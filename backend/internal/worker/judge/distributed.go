package judge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alonespig/zeorcode/backend/internal/infra/mq"
	judgev1 "github.com/alonespig/zeorcode/backend/internal/judgeproto"
	"github.com/alonespig/zeorcode/backend/internal/model"
	judgeapi "github.com/alonespig/zeorcode/backend/pkg/judge"

	"go.uber.org/zap"
)

func (w *Worker) processDistributedSubmission(
	ctx context.Context,
	task mq.StreamTask,
	s model.Submission,
	limits judgeapi.Limits,
	log *zap.SugaredLogger,
) (taskProcessOutcome, error) {
	if w.artifacts == nil {
		return taskProcessDeferred, errors.New("distributed judge artifact store is not configured")
	}
	descriptor, err := w.artifacts.Describe(ctx, s.ProblemID)
	if err != nil {
		s.Status = judgeapi.UnknownError
		if saveErr := w.saveResultGuarded(ctx, &s, dlqStageTestData, log); saveErr != nil {
			if errors.Is(saveErr, errSubmissionSuperseded) {
				return taskProcessCompleted, nil
			}
			return taskProcessDeferred, saveErr
		}
		w.sendToDLQ(s.ID, dlqStageTestData, err)
		return taskProcessCompleted, nil
	}

	req := &judgev1.JudgeRequest{
		SubmissionId:      s.ID,
		SubmissionVersion: int64(task.Version),
		Problem: &judgev1.ProblemArtifact{
			ProblemId: descriptor.ProblemID,
			Version:   descriptor.Version,
			Digest:    descriptor.Digest,
		},
		Language:         s.Language,
		SourceCode:       s.Code,
		TimeLimitMs:      int64(limits.TimeMs),
		MemoryLimitMb:    int64(limits.MemoryMB),
		OutputLimitBytes: int64(limits.OutputBytes),
	}
	progress := make([]SubmissionCaseResult, 0)
	final, err := w.agentPool.Judge(ctx, req, func(event *judgev1.JudgeEvent) error {
		if event.Type != judgev1.JudgeEventType_JUDGE_EVENT_TYPE_CASE_RESULT || event.CaseResult == nil {
			return nil
		}
		result := event.CaseResult
		progress = append(progress, SubmissionCaseResult{
			ID:         int(result.CaseId),
			Status:     int(result.Status),
			TimeUsed:   result.TimeNs / 1_000_000,
			MemoryUsed: result.MemoryBytes / 1024,
		})
		if err := w.publishDistributedProgress(s, progress); err != nil {
			log.Warnw("publish distributed judge progress failed", "err", err)
		}
		return nil
	})
	if err != nil {
		return taskProcessDeferred, fmt.Errorf("distributed judge: %w", err)
	}
	if final.Status == int32(judgeapi.CompileError) {
		s.Status = judgeapi.CompileError
		s.TimeUsed = 0
		s.MemoryUsed = 0
		s.Score = 0
		s.CompileOutput = final.CompileOutput
		if saveErr := w.saveResultGuarded(ctx, &s, dlqStageCompileSave, log); saveErr != nil && !errors.Is(saveErr, errSubmissionSuperseded) {
			return taskProcessDeferred, saveErr
		}
		return taskProcessCompleted, nil
	}

	caseResults := make([]*model.JudgeResult, 0, len(final.Cases))
	for _, result := range final.Cases {
		if result == nil {
			continue
		}
		caseResults = append(caseResults, &model.JudgeResult{
			SubmissionID: s.ID,
			CaseID:       int64(result.CaseId),
			Status:       int(result.Status),
			TimeUsed:     result.TimeNs,
			MemoryUsed:   result.MemoryBytes,
		})
	}
	if len(caseResults) == 0 {
		return taskProcessDeferred, errors.New("distributed judge returned no testcase results")
	}
	s.Status = int(final.Status)
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

func (w *Worker) publishDistributedProgress(s model.Submission, results []SubmissionCaseResult) error {
	payload, err := json.Marshal(map[string]any{
		"submission": SubmissionEventInfo{
			ID:        s.PublicID,
			Language:  s.Language,
			Status:    judgeapi.Accepted,
			CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"caseResults": results,
	})
	if err != nil {
		return err
	}
	return w.mq.PublishUpdate(context.Background(), s.ID, string(payload))
}
