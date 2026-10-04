package judge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/infra/logger"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	judgeapi "zoj/pkg/judge"

	"github.com/spf13/viper"
	"gorm.io/gorm"
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

// Worker 判题机：dispatcher 单线程拉 Redis，N 个 worker 并发处理
type Worker struct {
	db          *gorm.DB
	judgePool   *judgeapi.Pool
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
func NewWorker(db *gorm.DB, pool *judgeapi.Pool, repos Repositories, concurrency int, m *mq.MQ, c *cache.Cache) *Worker {
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
			s.Status = judgeapi.UnknownError
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
		s.Status = judgeapi.UnknownError
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
		s.Status = judgeapi.UnknownError
		if err := w.saveResultGuarded(ctx, &s, dlqStageCompileSave, log); err != nil && !errors.Is(err, errSubmissionSuperseded) {
			return taskProcessDeferred, err
		}
		log.Errorw("no judge instance configured")
		return taskProcessCompleted, nil
	}

	artifact, err := client.Compile(s.Language, s.Code)
	// 编译时若该机不可达（网络层），标记下线并换一台重试一次；用户代码编译失败不算不可达
	if err != nil && judgeapi.IsUnavailable(err) {
		log.Warnw("judge unavailable on compile, failover", "url", err)
		w.judgePool.MarkDown(client)
		if next := w.judgePool.Next(); next != nil {
			client = next
			artifact, err = client.Compile(s.Language, s.Code)
		}
	}
	if err != nil {
		s.Status = judgeapi.CompileError
		s.TimeUsed = 0
		s.MemoryUsed = 0
		s.Score = 0
		s.CompileOutput = judgeapi.CompilationOutput(err)
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
