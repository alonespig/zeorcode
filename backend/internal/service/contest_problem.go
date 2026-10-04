package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

func (s *ContestService) GetContestProblemList(ctx context.Context, contestID, userID int64, isAdmin bool) (*ContestProblemList, error) {
	if _, err := s.authorizeContestProblemAccess(ctx, contestID, userID, isAdmin); err != nil {
		return nil, err
	}

	problems, err := s.repo.GetContestProblems(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemIDs := make([]int64, 0, len(problems))
	for _, p := range problems {
		problemIDs = append(problemIDs, p.ProblemID)
	}

	resp := &ContestProblemList{
		ProblemList: make([]ContestProblem, 0, len(problems)),
	}

	filter := repository.ContestSubmissionFilter{
		UserID:     &userID,
		ContestID:  contestID,
		ProblemIDs: problemIDs,
	}

	submissionList, err := s.submitRepo.ContestSubmissionList(ctx, &filter)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	problemStatus := map[int64]int{}

	for _, submission := range submissionList {
		status, ok := problemStatus[submission.ProblemID]
		if submission.Status == int(judge.Accepted) ||
			(ok && status == judge.Accepted) {
			problemStatus[submission.ProblemID] = int(judge.Accepted)
		} else if submission.Status == int(judge.Pending) {
			problemStatus[submission.ProblemID] = int(judge.Pending)
		} else {
			problemStatus[submission.ProblemID] = int(judge.WrongAnswer)
		}
	}

	problemList, err := s.problemRepo.FindByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemMap := make(map[int64]model.Problem, len(problemList))
	for _, problem := range problemList {
		problemMap[problem.ID] = problem
	}

	statsList, err := s.submitRepo.GetContestProblemStats(ctx, contestID, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	statsMap := make(map[int64]repository.ProblemSubmissionStat, len(statsList))
	for _, stat := range statsList {
		statsMap[stat.ProblemID] = stat
	}

	for _, p := range problems {
		problem := problemMap[p.ProblemID]
		status := problemStatus[p.ProblemID]
		stat := statsMap[p.ProblemID]
		// 路①：status / 提交通过数均已从 submissions 现算（DB），不再叠 Redis 增量覆盖
		// —— 否则重判后这些计数会漂移。

		resp.ProblemList = append(resp.ProblemList, ContestProblem{
			ProblemID:     problem.DisplayID,
			Name:          problem.Name,
			Label:         p.Label,
			Color:         p.Color,
			Status:        status,
			TimeLimit:     1000,
			MemoryLimit:   128,
			TotalCount:    stat.SubmitCount,
			AcceptedCount: stat.AcceptedCount,
		})
	}
	return resp, nil
}

func (s *ContestService) applyContestProblemCache(
	ctx context.Context,
	contestID, userID, problemID int64,
	status *int,
	stat *repository.ProblemSubmissionStat,
) {
	if cachedStatus, ok := s.getCachedContestProblemStatus(ctx, contestID, userID, problemID); ok {
		*status = cachedStatus
	}
	s.applyCachedContestProblemStat(ctx, contestID, problemID, stat)
}

func (s *ContestService) getCachedContestProblemStatus(ctx context.Context, contestID, userID, problemID int64) (int, bool) {
	fields, err := s.cache.HGetAll(ctx, cache.ContestUserProblem(contestID, userID, problemID))
	if err != nil || len(fields) == 0 {
		return 0, false
	}
	status, err := strconv.Atoi(fields["status"])
	if err != nil {
		return 0, false
	}
	return status, true
}

func (s *ContestService) applyCachedContestProblemStat(ctx context.Context, contestID, problemID int64, stat *repository.ProblemSubmissionStat) {
	fields, err := s.cache.HGetAll(ctx, cache.ContestProblemStats(contestID, problemID))
	if err != nil || len(fields) == 0 {
		return
	}
	if submitCount, err := strconv.ParseInt(fields["submit_count"], 10, 64); err == nil {
		stat.SubmitCount = submitCount
	}
	if acceptedCount, err := strconv.ParseInt(fields["accepted_count"], 10, 64); err == nil {
		stat.AcceptedCount = acceptedCount
	}
}

func (s *ContestService) getCachedContestUserProblem(ctx context.Context, contestID, userID, problemID int64) (model.UserContestProblem, bool) {
	fields, err := s.cache.HGetAll(ctx, cache.ContestUserProblem(contestID, userID, problemID))
	if err != nil || len(fields) == 0 {
		return model.UserContestProblem{}, false
	}
	status, err := strconv.Atoi(fields["status"])
	if err != nil {
		return model.UserContestProblem{}, false
	}
	tries, _ := strconv.Atoi(fields["tries"])
	record := model.UserContestProblem{
		ContestID: contestID,
		UserID:    userID,
		ProblemID: problemID,
		Status:    status,
		UnAcCount: tries,
	}
	if acUnix, err := strconv.ParseInt(fields["ac_time"], 10, 64); err == nil && acUnix > 0 {
		acTime := time.Unix(acUnix, 0)
		record.AcTime = &acTime
	}
	return record, true
}

func (s *ContestService) problemNamesByIDs(ctx context.Context, contestProblems []model.ContestProblem) map[int64]string {
	ids := make([]int64, 0, len(contestProblems))
	for _, p := range contestProblems {
		ids = append(ids, p.ProblemID)
	}
	problems, _ := s.problemRepo.FindByIDs(ctx, ids)
	names := make(map[int64]string, len(problems))
	for _, p := range problems {
		names[p.ID] = p.Name
	}
	return names
}

// problemsByIDs 一次性取出比赛各题的完整信息（含对外题号），按主键索引
func (s *ContestService) problemsByIDs(ctx context.Context, contestProblems []model.ContestProblem) map[int64]model.Problem {
	ids := make([]int64, 0, len(contestProblems))
	for _, p := range contestProblems {
		ids = append(ids, p.ProblemID)
	}
	problems, _ := s.problemRepo.FindByIDs(ctx, ids)
	m := make(map[int64]model.Problem, len(problems))
	for _, p := range problems {
		m[p.ID] = p
	}
	return m
}

func (s *ContestService) GetContestProblemDetail(
	ctx context.Context,
	contestID int64,
	label string,
	userID int64,
	isAdmin bool,
) (*ContestProblemDetail, error) {
	if _, err := s.authorizeContestProblemAccess(ctx, contestID, userID, isAdmin); err != nil {
		return nil, err
	}
	contestProblem, err := s.repo.GetContestProblem(ctx, model.ContestProblem{
		ContestID: contestID,
		Label:     label,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problem, err := s.problemRepo.GetByID(ctx, contestProblem.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	hasTestData, err := problemTestDataReady(ctx, s.testData, problem)
	if err != nil {
		return nil, err
	}

	resp := &ContestProblemDetail{
		Name:         label + ". " + problem.Name,
		TimeLimit:    problem.TimeLimit,
		MemoryLimit:  problem.MemoryLimit,
		Description:  problem.Description,
		InputFormat:  problem.InputFormat,
		OutputFormat: problem.OutputFormat,
		Hint:         problem.Hint,
		Samples:      make([]ProblemSample, 0),
		HasTestData:  hasTestData,
	}
	samples, err := s.problemRepo.GetProblemSamplesByID(ctx, problem.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	for _, sample := range samples {
		resp.Samples = append(resp.Samples, ProblemSample{
			Input:   sample.Input,
			Output:  sample.Output,
			Explain: sample.Explain,
		})
	}

	return resp, nil
}

func (s *ContestService) GetContestSubmitInfo(
	ctx context.Context,
	contestID, userID int64,
	isAdmin bool,
) (*ContestSubmitInfo, error) {
	if _, err := s.authorizeContestProblemAccess(ctx, contestID, userID, isAdmin); err != nil {
		return nil, err
	}
	problem, err := s.repo.GetContestProblems(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &ContestSubmitInfo{
		ProblemList: make([]ContestProblemItem, 0, len(problem)),
	}
	problemNames := s.problemNamesByIDs(ctx, problem)
	for _, p := range problem {
		resp.ProblemList = append(resp.ProblemList, ContestProblemItem{
			Name:  p.Label + ". " + problemNames[p.ProblemID],
			Label: p.Label,
		})
	}
	return resp, nil
}
