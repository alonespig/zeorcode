package judgeworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"zoj/internal/common/consts"
	"zoj/internal/common/logger"
	"zoj/internal/dto"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DefaultConcurrency 评测并发 worker 数量默认值
const DefaultConcurrency = 4

// 死信阶段标记，便于人工排查
const (
	dlqStageCompileSave  = "compile_save"
	dlqStageRunCase      = "run_case"
	dlqStageCaseSave     = "case_save"
	dlqStageFinalSave    = "final_save"
	dlqStageUserProblem  = "user_problem"
	dlqStageContestStats = "contest_stats"
	dlqStageRemoteSubmit = "remote_submit"
	dlqStageTestData     = "test_data"
	dlqStageProblemLoad  = "problem_load"
)

// TestCase 单个测试点：输入内容喂沙箱，期望哈希用于 worker 侧判 AC/PE/WA
type TestCase struct {
	ID             int
	Input          string
	StrippedMd5    string // 去行末/文末空白后 md5 —— 判 AC
	AllStrippedMd5 string // 去所有空白后 md5 —— 判 PE
}

// Worker 判题机：dispatcher 单线程拉 Redis，N 个 worker 并发处理
type Worker struct {
	db          *gorm.DB
	judgePool   *judge.Pool
	subRepo     *repository.SubmissionRepo
	problemRepo *repository.ProblemRepo
	contestRepo *repository.ContestRepo
	accRepo     *repository.RemoteAccountRepo
	cache       *cache.Cache
	mq          *mq.MQ

	concurrency int
	tasks       chan mq.StreamTask  // dispatcher → workers（带 Stream 消息 id，处理完 ACK）
	remoteTasks chan *pendingRemote // worker → 远程轮询协程
}

// Repositories 汇总 Worker 使用的持久化适配器，由进程入口负责构造。
// Worker 不再持有“如何从数据库连接创建 repository”的装配职责。
type Repositories struct {
	Submissions    *repository.SubmissionRepo
	Problems       *repository.ProblemRepo
	Contests       *repository.ContestRepo
	RemoteAccounts *repository.RemoteAccountRepo
}

// NewWorker concurrency<=0 时回退默认值。mq / cache 由调用方（判题进程入口）注入。
// pool 为判题机负载均衡池（一台或多台 go-judge）。
func NewWorker(db *gorm.DB, pool *judge.Pool, repos Repositories, concurrency int, m *mq.MQ, c *cache.Cache) *Worker {
	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}
	return &Worker{
		db:          db,
		judgePool:   pool,
		subRepo:     repos.Submissions,
		problemRepo: repos.Problems,
		contestRepo: repos.Contests,
		accRepo:     repos.RemoteAccounts,
		cache:       c,
		mq:          m,
		concurrency: concurrency,
		tasks:       make(chan mq.StreamTask, concurrency),
		remoteTasks: make(chan *pendingRemote, 256),
	}
}

// Concurrency 当前并发数
func (w *Worker) Concurrency() int { return w.concurrency }

// Run 启动 1 个 dispatcher + N 个 worker，阻塞直到 ctx 被取消并优雅退出
func (w *Worker) Run(ctx context.Context) {
	// 单实例可在启动时回收自身上次崩溃遗留的占用；多实例绝不能清理其他实例正在使用的账号。
	if !viper.GetBool("judge.multi_instance") {
		if err := w.accRepo.ResetBusy(context.Background()); err != nil {
			logger.Errorw("reset remote account busy failed", "err", err)
		}
	}

	var wg sync.WaitGroup

	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w.workLoop(ctx, id)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(w.tasks)
		w.dispatch(ctx)
	}()

	// 远程评测轮询协程：worker 提交后立即释放，由它异步轮询出结果再落库
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.remotePollLoop(ctx)
	}()

	wg.Wait()
}

// dispatch 单线程从 Redis Stream(消费者组)取任务，下发给 worker。
// Stream 语义下：读到的消息在被 ACK 前一直挂在消费者的 PEL 里，
// 进程崩溃/关机时未 ACK 的任务不会丢——重启后先 drain 本消费者的 pending 重投，再读新消息。
func (w *Worker) dispatch(ctx context.Context) {
	consumer := streamConsumerName()

	// 重启恢复：把上次残留、未 ACK 的任务先重新下发（避免丢单）
	if pending, err := w.mq.ReadPendingSubmissions(context.Background(), consumer, 10_000); err != nil {
		logger.Errorw("read pending submissions failed", "err", err)
	} else {
		for _, t := range pending {
			logger.Infow("recover pending submission", "submissionID", t.SubmissionID)
			select {
			case <-ctx.Done():
				return
			case w.tasks <- t:
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 阻塞读新任务（无需 1s 轮询）；未 ACK 的任务留在 PEL，重启会恢复
		task, err := w.mq.ReadNewSubmission(context.Background(), consumer, 5*time.Second)
		if err != nil {
			logger.Errorw("read submission failed", "err", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
			continue
		}
		if task == nil {
			continue // block 超时，无新消息
		}

		select {
		case <-ctx.Done():
			return // 不 ACK，留在 PEL，下次启动 drain 恢复
		case w.tasks <- *task:
		}
	}
}

// workLoop 单个 worker 主循环：只有任务已可靠完成、已确认重复或已被新版本取代时才 ACK。
func (w *Worker) workLoop(ctx context.Context, workerID int) {
	for t := range w.tasks {
		for {
			logger.Infow("picked up submission", "workerID", workerID, "submissionID", t.SubmissionID, "version", t.Version)
			err := handleSubmissionTask(
				ctx,
				t,
				w.subRepo,
				w.mq,
				func(ctx context.Context, task mq.StreamTask, sub model.Submission) (taskProcessOutcome, error) {
					return w.processSubmission(ctx, task, sub)
				},
			)
			if err == nil {
				break
			}
			logger.Errorw("handle submission task failed; retrying", "workerID", workerID, "submissionID", t.SubmissionID, "msgID", t.MsgID, "err", err)
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

type submissionTaskStore interface {
	ClaimDispatchForJudging(ctx context.Context, dispatch repository.SubmissionDispatch, messageID string) (repository.DispatchClaimResult, error)
	GetSubmissionByID(ctx context.Context, submissionID int64) (*model.Submission, error)
	CompleteDispatchForJudging(ctx context.Context, dispatch repository.SubmissionDispatch, messageID string) (bool, error)
}

type submissionTaskAcker interface {
	AckSubmission(ctx context.Context, msgID string) error
}

type taskProcessOutcome uint8

const (
	taskProcessCompleted taskProcessOutcome = iota
	taskProcessDeferred
	taskProcessRequeued
)

// handleSubmissionTask 以 (submission_id, version, Redis message ID) 原子认领任务。
// 返回错误或异步处理中时消息保留在 PEL；重复、过时和可靠完成的消息才会 ACK。
func handleSubmissionTask(
	ctx context.Context,
	task mq.StreamTask,
	store submissionTaskStore,
	acker submissionTaskAcker,
	process func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error),
) error {
	dispatch := repository.SubmissionDispatch{SubmissionID: task.SubmissionID, Version: task.Version}
	claim, err := store.ClaimDispatchForJudging(ctx, dispatch, task.MsgID)
	if err != nil {
		return fmt.Errorf("claim submission %d version %d: %w", task.SubmissionID, task.Version, err)
	}
	switch claim {
	case repository.DispatchDuplicate,
		repository.DispatchAlreadyCompleted,
		repository.DispatchStale,
		repository.DispatchSubmissionMissing:
		return ackSubmissionTask(ctx, task, acker)
	case repository.DispatchClaimed:
	default:
		return fmt.Errorf("claim submission %d version %d returned unknown result %d", task.SubmissionID, task.Version, claim)
	}

	sub, err := store.GetSubmissionByID(ctx, task.SubmissionID)
	if err != nil {
		return fmt.Errorf("load submission %d: %w", task.SubmissionID, err)
	}
	outcome, err := process(ctx, task, *sub)
	if err != nil {
		return fmt.Errorf("process submission %d: %w", task.SubmissionID, err)
	}
	switch outcome {
	case taskProcessDeferred:
		return nil
	case taskProcessRequeued:
		return ackSubmissionTask(ctx, task, acker)
	case taskProcessCompleted:
		completed, err := store.CompleteDispatchForJudging(ctx, dispatch, task.MsgID)
		if err != nil {
			return fmt.Errorf("complete submission %d version %d: %w", task.SubmissionID, task.Version, err)
		}
		if !completed {
			return fmt.Errorf("complete submission %d version %d: dispatch ownership lost", task.SubmissionID, task.Version)
		}
		return ackSubmissionTask(ctx, task, acker)
	default:
		return fmt.Errorf("process submission %d returned unknown outcome %d", task.SubmissionID, outcome)
	}
}

func ackSubmissionTask(ctx context.Context, task mq.StreamTask, acker submissionTaskAcker) error {
	if err := acker.AckSubmission(ctx, task.MsgID); err != nil {
		return fmt.Errorf("ack submission %d: %w", task.SubmissionID, err)
	}
	return nil
}

// streamConsumerName 消费者名：同一台机器上稳定不变，重启后能 drain 到上次的 pending。
func streamConsumerName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "worker"
	}
	return "judge-" + host
}

var errSubmissionSuperseded = errors.New("submission superseded by rejudge")

// processSubmission 单次评测的主流程编排。
// completed 表示最终结果已落库或本版本已被重判取代；deferred 表示远程判题仍在异步轮询。
func (w *Worker) processSubmission(
	ctx context.Context,
	task mq.StreamTask,
	s model.Submission,
) (taskProcessOutcome, error) {
	log := logger.S().With("submissionID", s.ID, "problemID", s.ProblemID)

	// 远程题：提交到远程 OJ，交给轮询协程异步出结果，不走本地沙箱
	problem, err := w.problemRepo.GetByID(ctx, s.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.Status = judge.UnknownError
			s.TimeUsed = 0
			s.MemoryUsed = 0
			s.Score = 0
			if saveErr := w.saveResultGuarded(ctx, &s, dlqStageProblemLoad, log); saveErr != nil {
				if errors.Is(saveErr, errSubmissionSuperseded) {
					return taskProcessCompleted, nil
				}
				return taskProcessDeferred, saveErr
			}
			w.sendToDLQ(s.ID, dlqStageProblemLoad, err)
			return taskProcessCompleted, nil
		}
		return taskProcessDeferred, fmt.Errorf("load problem %d: %w", s.ProblemID, err)
	}
	if problem.OJ != "" {
		return w.dispatchRemote(ctx, task, s, problem, log)
	}

	lim := w.loadLimits(s.ProblemID, log)
	cases, err := w.loadCases(s.ProblemID)
	if err != nil {
		s.Status = judge.UnknownError
		s.TimeUsed = 0
		s.MemoryUsed = 0
		s.Score = 0
		if saveErr := w.saveResultGuarded(ctx, &s, dlqStageTestData, log); saveErr != nil {
			if errors.Is(saveErr, errSubmissionSuperseded) {
				return taskProcessCompleted, nil
			}
			return taskProcessDeferred, saveErr
		}
		log.Errorw("test data unavailable", "err", err)
		w.sendToDLQ(s.ID, dlqStageTestData, err)
		return taskProcessCompleted, nil
	}

	// 从池子挑一台判题机；一次提交的编译/运行/删除都用这一台（编译产物在它上面）
	client := w.judgePool.Next()
	if client == nil {
		s.Status = judge.UnknownError
		if err := w.saveResultGuarded(ctx, &s, dlqStageCompileSave, log); err != nil && !errors.Is(err, errSubmissionSuperseded) {
			return taskProcessDeferred, err
		}
		log.Errorw("no judge instance configured")
		return taskProcessCompleted, nil
	}

	artifact, err := client.Compile(s.Language, s.Code)
	// 编译时若该机不可达（网络层），标记下线并换一台重试一次；用户代码编译失败不算不可达
	if err != nil && judge.IsUnavailable(err) {
		log.Warnw("judge unavailable on compile, failover", "url", err)
		w.judgePool.MarkDown(client)
		if next := w.judgePool.Next(); next != nil {
			client = next
			artifact, err = client.Compile(s.Language, s.Code)
		}
	}
	if err != nil {
		s.Status = judge.CompileError
		s.TimeUsed = 0
		s.MemoryUsed = 0
		s.Score = 0
		s.CompileOutput = judge.CompilationOutput(err)
		if saveErr := w.saveResultGuarded(ctx, &s, dlqStageCompileSave, log); saveErr != nil && !errors.Is(saveErr, errSubmissionSuperseded) {
			return taskProcessDeferred, saveErr
		}
		log.Warnw("compile failed", "language", s.Language, "err", err)
		return taskProcessCompleted, nil
	}
	defer func() {
		if err := client.DeleteArtifact(artifact); err != nil {
			log.Warnw("delete cached artifact failed", "err", err)
		}
	}()

	runResult, err := w.runAllCases(ctx, &s, client, artifact, cases, lim, log)
	if err != nil {
		return taskProcessDeferred, err
	}
	if runResult.terminal {
		return taskProcessCompleted, nil
	}
	caseResults := runResult.results
	s.Status = runResult.finalStatus

	// 比赛/作业提交按测试点均分算本次得分；普通练习恒 0
	s.Score = w.submissionScore(&s, caseResults)

	saved, err := w.finalizeSubmission(ctx, &s, caseResults, log)
	if err != nil {
		return taskProcessDeferred, err
	}
	if !saved {
		return taskProcessCompleted, nil
	}

	if s.ContestID == 0 {
		w.updateUserProblem(s, log)
	} else {
		w.updateContestStats(s, log)
	}
	return taskProcessCompleted, nil
}

// saveResultGuarded 乐观锁写回评测结果：仅当 submissions.version 未被重判改动时才落库。
// 只更新结果相关列，绝不动 version（version 仅由重判 +1）。
// version 不符返回 errSubmissionSuperseded；数据库错误会进入 DLQ 并原样返回。
func (w *Worker) saveResultGuarded(ctx context.Context, s *model.Submission, stage string, log *zap.SugaredLogger) error {
	res := w.db.WithContext(ctx).Model(&model.Submission{}).
		Where("id = ? AND version = ?", s.ID, s.Version).
		Updates(map[string]any{
			"status":         s.Status,
			"time_used":      s.TimeUsed,
			"memory_used":    s.MemoryUsed,
			"compile_output": s.CompileOutput,
			"score":          s.Score,
		})
	if res.Error != nil {
		log.Errorw("save submission failed", "err", res.Error)
		w.sendToDLQ(s.ID, stage, res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		// version 对不上：判题期间被重判，这一轮已过时，安静丢弃，不发 SSE、不更新聚合
		log.Infow("submission superseded by rejudge, drop stale result", "submissionID", s.ID, "version", s.Version)
		return errSubmissionSuperseded
	}
	return nil
}

// loadLimits 读题目限制；失败回退默认
func (w *Worker) loadLimits(problemID int64, log *zap.SugaredLogger) judge.Limits {
	lim := judge.Limits{TimeMs: 1000, MemoryMB: 128}
	problem, err := w.problemRepo.GetByID(context.Background(), problemID)
	if err != nil {
		log.Warnw("get problem limits failed, fallback to defaults", "err", err)
		return lim
	}
	if problem.TimeLimit > 0 {
		lim.TimeMs = problem.TimeLimit
	}
	if problem.MemoryLimit > 0 {
		lim.MemoryMB = problem.MemoryLimit
	}
	return lim
}

// loadCases 读 info.json 拿测试点清单（缺失/损坏则扫描目录懒生成），
// 再逐点读输入内容喂沙箱；期望输出以清单里的哈希形式随 case 带回。
func (w *Worker) loadCases(problemID int64) ([]TestCase, error) {
	testDir := fmt.Sprintf("%s/%d/", viper.GetString("judge.data_dir"), problemID)
	return loadCasesFromDir(testDir)
}

func loadCasesFromDir(testDir string) ([]TestCase, error) {
	info, err := loadOrGenInfo(testDir)
	if err != nil {
		return nil, fmt.Errorf("load testcase info from %q: %w", testDir, err)
	}
	if info.Count <= 0 || len(info.Cases) == 0 {
		return nil, fmt.Errorf("test data directory %q contains no complete cases", testDir)
	}
	if info.Count != len(info.Cases) {
		return nil, fmt.Errorf("testcase count mismatch in %q: count=%d cases=%d", testDir, info.Count, len(info.Cases))
	}

	cases := make([]TestCase, 0, len(info.Cases))
	for _, c := range info.Cases {
		if c.Input == "" || filepath.Base(c.Input) != c.Input {
			return nil, fmt.Errorf("invalid testcase input path for case %d", c.ID)
		}
		if c.StrippedMd5 == "" || c.AllStrippedMd5 == "" {
			return nil, fmt.Errorf("missing expected output hash for case %d", c.ID)
		}
		input, err := os.ReadFile(filepath.Join(testDir, c.Input))
		if err != nil {
			return nil, fmt.Errorf("read input for case %d: %w", c.ID, err)
		}
		cases = append(cases, TestCase{
			ID:             c.ID,
			Input:          string(input),
			StrippedMd5:    c.StrippedMd5,
			AllStrippedMd5: c.AllStrippedMd5,
		})
	}
	return cases, nil
}

type runCasesResult struct {
	results     []*model.JudgeResult
	finalStatus int
	terminal    bool
}

// runAllCases 逐点执行、写 JudgeResult、推送进度。
// terminal=true 表示运行错误已经作为 UnknownError 可靠落库，调用方可以结束当前任务。
func (w *Worker) runAllCases(
	ctx context.Context,
	s *model.Submission,
	client *judge.Client,
	artifact *judge.Artifact,
	cases []TestCase,
	lim judge.Limits,
	log *zap.SugaredLogger,
) (runCasesResult, error) {
	if len(cases) == 0 {
		return runCasesResult{}, errors.New("cannot judge submission without test cases")
	}

	results := make([]*model.JudgeResult, 0, len(cases))
	caseResultDto := make([]dto.SubmissionCaseResult, 0, len(cases))

	subDto := dto.SubmissionEventInfo{
		ID:        s.PublicID,
		Language:  s.Language,
		Status:    judge.Accepted,
		CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	mp := map[string]any{"submission": subDto}

	finalStatus := judge.Accepted

	for index, tc := range cases {
		verdict, err := client.Run(artifact, tc.Input, lim)
		if err != nil {
			log.Errorw("judge run failed", "caseID", tc.ID, "err", err)
			s.Status = judge.UnknownError
			s.TimeUsed = 0
			s.MemoryUsed = 0
			s.Score = 0
			saveErr := w.saveResultGuarded(ctx, s, dlqStageRunCase, log)
			if saveErr != nil && !errors.Is(saveErr, errSubmissionSuperseded) {
				return runCasesResult{}, saveErr
			}
			return runCasesResult{terminal: true}, nil
		}

		// 沙箱跑干净(Accepted)时，用清单哈希判 AC/PE/WA；TLE/MLE/RE 等运行态直接沿用
		status := verdict.Status
		if status == judge.Accepted {
			status = judgeOutput(verdict.Stdout, tc.StrippedMd5, tc.AllStrippedMd5)
		}

		log.Infow("case judged",
			"caseID", tc.ID,
			"status", status,
			"timeNs", verdict.TimeNs,
			"memoryBytes", verdict.MemoryByte,
		)

		if status != judge.Accepted && finalStatus == judge.Accepted {
			finalStatus = status
		}

		jr := model.JudgeResult{
			SubmissionID: s.ID,
			CaseID:       int64(index + 1),
			Status:       status,
			TimeUsed:     verdict.TimeNs,
			MemoryUsed:   verdict.MemoryByte,
		}
		caseResultDto = append(caseResultDto, dto.SubmissionCaseResult{
			ID:         index + 1,
			Status:     jr.Status,
			TimeUsed:   jr.TimeUsed / 1_000_000,
			MemoryUsed: jr.MemoryUsed / 1024,
		})

		results = append(results, &jr)

		// 逐点推送进度（SSE 实时性靠这个，不依赖落库）
		mp["caseResults"] = caseResultDto
		data, err := json.Marshal(mp)
		if err != nil {
			log.Errorw("marshal case result failed", "caseID", index+1, "err", err)
			continue
		}
		if err := w.mq.PublishUpdate(context.Background(), s.ID, string(data)); err != nil {
			log.Errorw("publish case update failed", "caseID", index+1, "err", err)
		}
	}

	return runCasesResult{results: results, finalStatus: finalStatus}, nil
}

// finalizeSubmission 汇总 max(time/memory) → save 主提交 → 发 done
// saved=false 且 err=nil 表示该版本已被重判取代，不再推送或更新聚合。
func (w *Worker) finalizeSubmission(
	ctx context.Context,
	s *model.Submission,
	results []*model.JudgeResult,
	log *zap.SugaredLogger,
) (bool, error) {
	var timeUsed, memoryUsed int64
	for _, cr := range results {
		if t := cr.TimeUsed / 1_000_000; t > timeUsed {
			timeUsed = t
		}
		if m := cr.MemoryUsed / 1024; m > memoryUsed {
			memoryUsed = m
		}
	}
	s.TimeUsed = timeUsed
	s.MemoryUsed = memoryUsed

	saved, err := w.subRepo.SaveJudgingResult(ctx, s, results)
	if err != nil {
		log.Errorw("save submission and judge results failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageCaseSave, err)
		return false, err
	}
	if !saved {
		log.Infow("submission superseded by rejudge, drop stale result", "submissionID", s.ID, "version", s.Version)
		return false, nil
	}

	caseDtos := make([]dto.SubmissionCaseResult, 0, len(results))
	for i, cr := range results {
		caseDtos = append(caseDtos, dto.SubmissionCaseResult{
			ID:         i + 1,
			Status:     cr.Status,
			TimeUsed:   cr.TimeUsed / 1_000_000,
			MemoryUsed: cr.MemoryUsed / 1024,
		})
	}
	mp := map[string]any{
		"submission": dto.SubmissionEventInfo{
			ID:         s.PublicID,
			Language:   s.Language,
			Status:     s.Status,
			TimeUsed:   s.TimeUsed,
			MemoryUsed: s.MemoryUsed,
			CreatedAt:  s.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"caseResults": caseDtos,
	}

	data, err := json.Marshal(mp)
	if err != nil {
		log.Errorw("marshal final result failed", "err", err)
		return true, nil
	}
	if err := w.mq.PublishDone(context.Background(), s.ID, string(data)); err != nil {
		log.Errorw("publish submission done failed", "err", err)
	}
	return true, nil
}

// updateUserProblem 非比赛提交：从 submissions 明细重算该用户该题的做题统计（幂等，重判安全）。
// 不再用 ++ 累加 —— 那样重判会重复计数；改为按当前所有有效提交现算。
func (w *Worker) updateUserProblem(s model.Submission, log *zap.SugaredLogger) {
	// AC 可能改变全站排名 → 换代使用户榜所有分页缓存失效
	if s.Status == judge.Accepted {
		defer w.cache.Incr(context.Background(), cache.UserRankGen())
	}

	subs, err := w.subRepo.ListUserProblemSubmissionsForRecompute(context.Background(), s.UserID, s.ProblemID)
	if err != nil {
		log.Errorw("recompute user problem: list submissions failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageUserProblem, err)
		return
	}
	if len(subs) == 0 {
		return // 当前这条（非 CE、已判）理应在内；为空则无需处理
	}
	up := model.UserProblem{
		UserID:      s.UserID,
		ProblemID:   s.ProblemID,
		SubmitCount: len(subs),
	}
	for _, r := range subs {
		if r.Status == judge.Accepted {
			up.AcCount++
		}
	}
	if up.AcCount > 0 {
		up.Status = judge.Accepted
	} else {
		up.Status = subs[0].Status
	}
	if err := w.problemRepo.UpsertUserProblem(context.Background(), &up); err != nil {
		log.Errorw("recompute user problem: upsert failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageUserProblem, err)
	}
}

// caseAverageScore 测试点均分：通过点数 / 总点数 × 满分（四舍五入）。
// 比赛和作业共用，区别只在满分从哪来。
func caseAverageScore(results []*model.JudgeResult, full int) int {
	total := len(results)
	if total == 0 || full <= 0 {
		return 0
	}
	passed := 0
	for _, r := range results {
		if r.Status == judge.Accepted {
			passed++
		}
	}
	return int(math.Round(float64(passed) / float64(total) * float64(full)))
}

// contestProblemFullScore 取比赛某题的满分，未配置则按 100。
func (w *Worker) contestProblemFullScore(contestID, problemID int64) int {
	if cp, err := w.contestRepo.GetContestProblem(context.Background(), model.ContestProblem{
		ContestID: contestID,
		ProblemID: problemID,
	}); err == nil && cp != nil && cp.Score > 0 {
		return cp.Score
	}
	return 100
}

// submissionScore 按提交归属算本次得分：
//   - 比赛提交：测试点均分 × 该题满分（OI/IOI 榜单用；ACM 忽略，但一并存下便于展示）
//   - 作业提交：测试点均分 × 100（作业不做逐题配分）
//   - 其他（普通练习）：不计分，恒 0
func (w *Worker) submissionScore(s *model.Submission, results []*model.JudgeResult) int {
	switch {
	case s.ContestID != 0:
		return caseAverageScore(results, w.contestProblemFullScore(s.ContestID, s.ProblemID))
	case s.HomeworkID != 0:
		return caseAverageScore(results, model.HomeworkProblemFullScore)
	default:
		return 0
	}
}

// updateScoreContestStats OI/IOI：按得分维护 UserContestProblem.Score
// IOI 取历史最高分；OI 取最后一次提交分。行锁串行化同一 (contest,user,problem)。
func (w *Worker) updateScoreContestStats(s model.Submission, contest model.Contest, log *zap.SugaredLogger) {
	submittedAt := s.CreatedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now()
	}
	err := w.db.Transaction(func(tx *gorm.DB) error {
		var ucp model.UserContestProblem
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND user_id = ? AND problem_id = ?", s.ContestID, s.UserID, s.ProblemID).
			First(&ucp).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ucp = model.UserContestProblem{
				ContestID: s.ContestID,
				UserID:    s.UserID,
				ProblemID: s.ProblemID,
				Status:    s.Status,
				Score:     s.Score,
				SubCount:  1,
				AcTime:    &submittedAt,
			}
			if s.Status == judge.Accepted {
				ucp.AcCount = 1
			} else {
				ucp.UnAcCount = 1
			}
			return tx.Create(&ucp).Error
		}
		if err != nil {
			return err
		}

		ucp.SubCount++
		ucp.Status = s.Status
		if s.Status == judge.Accepted {
			ucp.AcCount++
		} else {
			ucp.UnAcCount++
		}
		if contest.Type == consts.ContestIOI {
			// IOI：分数提高才更新（记录达成时间，用于同分排序）
			if s.Score > ucp.Score {
				ucp.Score = s.Score
				ucp.AcTime = &submittedAt
			}
		} else {
			// OI：取最后一次提交分
			ucp.Score = s.Score
			ucp.AcTime = &submittedAt
		}
		return tx.Save(&ucp).Error
	})
	if err != nil {
		log.Errorw("update score contest problem failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageContestStats, err)
	}
}

// updateContestStats 比赛提交：维护比赛榜
// 注意：报名走 JoinContest 接口，这里不再补写 ContestUser
//
// 用事务 + 行锁串行化同一 (contest,user,problem) 的并发提交，避免并发下
// “都读到还没 AC → 都新增/都累加”造成的丢失更新或重复行；配合
// user_contest_problem 上的唯一索引 (contest_id,user_id,problem_id) 兜底。
func (w *Worker) updateContestStats(s model.Submission, log *zap.SugaredLogger) {
	// 比赛榜数据变了 → 让排行榜缓存失效，下次查实时重算
	defer w.cache.Delete(context.Background(), cache.ContestRank(s.ContestID))

	// OI/IOI 走得分榜：本次得分已在 processSubmission 落到 s.Score，这里维护 UserContestProblem.Score
	if contest, err := w.contestRepo.GetContestByID(context.Background(), s.ContestID); err == nil &&
		(contest.Type == consts.ContestOI || contest.Type == consts.ContestIOI) {
		if s.Status == judge.CompileError {
			return
		}
		w.updateScoreContestStats(s, *contest, log)
		return
	}

	submittedAt := s.CreatedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now()
	}
	var ucp model.UserContestProblem
	isNewAccepted := false

	err := w.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND user_id = ? AND problem_id = ?", s.ContestID, s.UserID, s.ProblemID).
			First(&ucp).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 该用户该题第一次有记录
			ucp = model.UserContestProblem{
				ContestID: s.ContestID,
				UserID:    s.UserID,
				ProblemID: s.ProblemID,
				Status:    s.Status,
			}
			if s.Status == judge.Accepted {
				ucp.AcCount = 1
				ucp.AcTime = &submittedAt
				isNewAccepted = true
			} else if s.Status != judge.CompileError {
				ucp.UnAcCount = 1 // CE 不计入错误提交（对齐 CF/ICPC 惯例）
			}
			// 唯一索引兜底：并发下若另一 worker 抢先插入，这里会撞唯一键报错，
			// 事务回滚后交给 DLQ 重试，重试时改走下面“已存在→更新”分支。
			return tx.Create(&ucp).Error
		}
		if err != nil {
			return err
		}

		// 已有行且已加锁，安全地读-改-写
		if s.Status == judge.Accepted && ucp.Status != judge.Accepted {
			ucp.Status = judge.Accepted
			ucp.AcCount = 1
			ucp.AcTime = &submittedAt
			isNewAccepted = true
		} else if s.Status != judge.Accepted && ucp.Status != judge.Accepted {
			ucp.Status = s.Status
			if s.Status != judge.CompileError {
				ucp.UnAcCount++ // CE 不计入错误提交
			}
		}
		return tx.Save(&ucp).Error
	})
	if err != nil {
		log.Errorw("update user contest problem failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageContestStats, err)
		return
	}

	w.updateContestCache(s, ucp, isNewAccepted, log)
}

func (w *Worker) updateContestCache(s model.Submission, ucp model.UserContestProblem, isNewAccepted bool, log *zap.SugaredLogger) {
	ctx := context.Background()

	// Redis stores running contest counters for polling-heavy pages. The database
	// remains the source of truth, so cache failures are warnings only.
	if err := w.cache.SAdd(ctx, cache.ContestProblemSubmittedUsers(s.ContestID, s.ProblemID), s.UserID); err != nil {
		log.Warnw("cache submitted user failed", "err", err)
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestProblemStats(s.ContestID, s.ProblemID), "submit_count", 1); err != nil {
		log.Warnw("cache submit count failed", "err", err)
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestUserStats(s.ContestID, s.UserID), "submit_count", 1); err != nil {
		log.Warnw("cache user submit count failed", "err", err)
	}

	acTime := int64(0)
	if ucp.AcTime != nil {
		acTime = ucp.AcTime.Unix()
	}
	if err := w.cache.HSet(ctx, cache.ContestUserProblem(s.ContestID, s.UserID, s.ProblemID),
		"status", ucp.Status,
		"tries", ucp.UnAcCount,
		"ac_time", acTime,
		"updated_at", time.Now().Unix(),
	); err != nil {
		log.Warnw("cache user problem status failed", "err", err)
	}

	if s.Status != judge.Accepted {
		return
	}
	if err := w.cache.SAdd(ctx, cache.ContestProblemAcceptedUsers(s.ContestID, s.ProblemID), s.UserID); err != nil {
		log.Warnw("cache accepted user failed", "err", err)
	}
	if err := w.cache.SAdd(ctx, cache.ContestUserAcceptedProblems(s.ContestID, s.UserID), s.ProblemID); err != nil {
		log.Warnw("cache accepted problem failed", "err", err)
	}
	if !isNewAccepted {
		return
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestProblemStats(s.ContestID, s.ProblemID), "accepted_count", 1); err != nil {
		log.Warnw("cache accepted count failed", "err", err)
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestUserStats(s.ContestID, s.UserID), "pass_count", 1); err != nil {
		log.Warnw("cache user pass count failed", "err", err)
	}
	if ucp.AcTime == nil {
		return
	}
	contest, err := w.contestRepo.GetContestByID(context.Background(), s.ContestID)
	if err != nil {
		log.Warnw("get contest for cache penalty failed", "err", err)
		return
	}
	penalty := int(ucp.AcTime.Sub(contest.StartTime).Minutes()) + 20*ucp.UnAcCount
	if _, err := w.cache.HIncrBy(ctx, cache.ContestUserStats(s.ContestID, s.UserID), "penalty", int64(penalty)); err != nil {
		log.Warnw("cache user penalty failed", "err", err)
	}
}

// sendToDLQ 把致命错误的上下文 JSON 推到死信队列
// 本函数不再上报错误（避免嵌套），只 log
func (w *Worker) sendToDLQ(submissionID int64, stage string, cause error) {
	payload, _ := json.Marshal(map[string]any{
		"submissionID": submissionID,
		"stage":        stage,
		"error":        cause.Error(),
		"ts":           time.Now().Unix(),
	})
	if err := w.mq.DeadLetterSubmission(context.Background(), string(payload)); err != nil {
		logger.Errorw("push DLQ failed", "submissionID", submissionID, "stage", stage, "err", err)
	}
}
