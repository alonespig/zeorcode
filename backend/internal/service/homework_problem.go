package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"zoj/internal/model"
	"zoj/pkg/errcode"

	"gorm.io/gorm"
)

func (s *HomeworkService) ProblemDetail(
	ctx context.Context,
	homeworkID int64,
	problemDisplayID string,
	userID int64,
	isSiteAdmin bool,
) (*ProblemDetail, error) {
	hw, access, err := s.loadWithAccess(ctx, homeworkID, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	if hw.Status(time.Now()) == model.HomeworkNotStarted && canEditHomework(access, hw) != nil && !access.CanManage() {
		return nil, errcode.ErrHomeworkNotStarted
	}

	problemID, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(problemDisplayID))
	if err != nil {
		return nil, errcode.ErrProblemNotFound
	}
	refs, err := s.repo.ProblemRefsByHomeworkIDs(ctx, []int64{homeworkID})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	inHomework := false
	for _, ref := range refs {
		if ref.ProblemID == problemID {
			inHomework = true
			break
		}
	}
	if !inHomework {
		return nil, errcode.ErrProblemNotFound.WithMsg("该题不属于本作业")
	}

	problem, err := s.problemRepo.GetByID(ctx, problemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	hasTestData, err := problemTestDataReady(ctx, s.testData, problem)
	if err != nil {
		return nil, err
	}

	rawTags, err := s.problemRepo.GetProblemTagByID(ctx, problem.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	tags := make([]TagItem, 0, len(rawTags))
	for _, tag := range rawTags {
		tags = append(tags, TagItem{ID: tag.ID, Name: tag.Name})
	}
	rawSamples, err := s.problemRepo.GetProblemSamplesByID(ctx, problem.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	samples := make([]ProblemSample, 0, len(rawSamples))
	for _, sample := range rawSamples {
		samples = append(samples, ProblemSample{
			Input: sample.Input, Output: sample.Output, Explain: sample.Explain,
		})
	}

	stats, err := s.submitRepo.GetHomeworkProblemStats(ctx, homeworkID, []int64{problem.ID})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	acceptedCount, submitCount := 0, 0
	if len(stats) > 0 {
		acceptedCount = int(stats[0].AcceptedCount)
		submitCount = int(stats[0].SubmitCount)
	}

	return &ProblemDetail{
		ID:              problem.DisplayID,
		Name:            problem.Name,
		Difficulty:      problem.Difficulty,
		TimeLimit:       problem.TimeLimit,
		MemoryLimit:     problem.MemoryLimit,
		Description:     problem.Description,
		InputFormat:     problem.InputFormat,
		OutputFormat:    problem.OutputFormat,
		SubmitCount:     submitCount,
		AcceptedCount:   acceptedCount,
		Hint:            problem.Hint,
		Samples:         samples,
		Tags:            tags,
		OJ:              problem.OJ,
		RemoteProblemID: problem.RemoteProblemID,
		Hidden:          problem.Hidden,
		HasTestData:     hasTestData,
	}, nil
}

// Save 布置（id==0）或编辑作业。
// 布置需要团队管理员及以上；编辑按方案 B 只允许布置者本人。

func (s *HomeworkService) resolveProblems(ctx context.Context, displayIDs []string) ([]model.HomeworkProblem, error) {
	problems := make([]model.HomeworkProblem, 0, len(displayIDs))
	seen := make(map[int64]struct{}, len(displayIDs))
	for i, displayID := range displayIDs {
		pid, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(displayID))
		if err != nil {
			return nil, errcode.ErrProblemNotFound.WithMsg("题目不存在：" + displayID)
		}
		if _, dup := seen[pid]; dup {
			return nil, errcode.ErrInvalidParams.WithMsg("题目重复：" + displayID)
		}
		seen[pid] = struct{}{}
		problems = append(problems, model.HomeworkProblem{ProblemID: pid, Sort: i})
	}
	return problems, nil
}
