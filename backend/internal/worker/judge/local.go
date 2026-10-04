package judge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"zoj/internal/model"
	judgeapi "zoj/pkg/judge"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// TestCase 单个测试点：输入内容喂沙箱，期望哈希用于 worker 侧判 AC/PE/WA
type TestCase struct {
	ID             int
	Input          string
	StrippedMd5    string // 去行末/文末空白后 md5 —— 判 AC
	AllStrippedMd5 string // 去所有空白后 md5 —— 判 PE
}

// loadLimits 读题目限制；失败回退默认
func (w *Worker) loadLimits(problemID int64, log *zap.SugaredLogger) judgeapi.Limits {
	lim := judgeapi.Limits{TimeMs: 1000, MemoryMB: 128}
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
	client *judgeapi.Client,
	artifact *judgeapi.Artifact,
	cases []TestCase,
	lim judgeapi.Limits,
	log *zap.SugaredLogger,
) (runCasesResult, error) {
	if len(cases) == 0 {
		return runCasesResult{}, errors.New("cannot judge submission without test cases")
	}

	results := make([]*model.JudgeResult, 0, len(cases))
	caseResultDto := make([]SubmissionCaseResult, 0, len(cases))

	subDto := SubmissionEventInfo{
		ID:        s.PublicID,
		Language:  s.Language,
		Status:    judgeapi.Accepted,
		CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	mp := map[string]any{"submission": subDto}

	finalStatus := judgeapi.Accepted

	for index, tc := range cases {
		verdict, err := client.Run(artifact, tc.Input, lim)
		if err != nil {
			log.Errorw("judge run failed", "caseID", tc.ID, "err", err)
			s.Status = judgeapi.UnknownError
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
		if status == judgeapi.Accepted {
			status = judgeOutput(verdict.Stdout, tc.StrippedMd5, tc.AllStrippedMd5)
		}

		log.Infow("case judged",
			"caseID", tc.ID,
			"status", status,
			"timeNs", verdict.TimeNs,
			"memoryBytes", verdict.MemoryByte,
		)

		if status != judgeapi.Accepted && finalStatus == judgeapi.Accepted {
			finalStatus = status
		}

		jr := model.JudgeResult{
			SubmissionID: s.ID,
			CaseID:       int64(index + 1),
			Status:       status,
			TimeUsed:     verdict.TimeNs,
			MemoryUsed:   verdict.MemoryByte,
		}
		caseResultDto = append(caseResultDto, SubmissionCaseResult{
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
