package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"zoj/internal/common/consts"
	"zoj/internal/common/errcode"
	"zoj/internal/common/logger"
	"zoj/internal/dto"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"
	"zoj/pkg/rating"
	"zoj/pkg/scoring"
	"zoj/pkg/util"

	"gorm.io/gorm"
)

// contestRankTTL 排行榜缓存有效期：轮询期间命中缓存，最多滞后这么久
const contestRankTTL = 5 * time.Second

// initialRating 首次参加 rated 比赛的起算基线（rating 为 0=未定级时按此计算）
const initialRating = 1200

type ContestService struct {
	repo        *repository.ContestRepo
	problemRepo *repository.ProblemRepo
	submitRepo  *repository.SubmissionRepo
	userRepo    *repository.UserRepo
	cache       *cache.Cache
	mq          *mq.MQ
	notify      *NotificationService
	testData    ProblemTestDataReader
	languages   LanguageResolver
}

func NewContestService(repo *repository.ContestRepo,
	problem *repository.ProblemRepo,
	submitRepo *repository.SubmissionRepo,
	userRepo *repository.UserRepo,
	cache *cache.Cache,
	mq *mq.MQ,
	notify *NotificationService,
	testData ProblemTestDataReader,
	languages LanguageResolver) *ContestService {
	return &ContestService{repo: repo, problemRepo: problem, submitRepo: submitRepo, userRepo: userRepo, cache: cache, mq: mq, notify: notify, testData: testData, languages: languages}
}

func (s *ContestService) CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.CreateContestResp, error) {
	data := req.ContestDate + " " + req.ContestTime + ":00"
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", data, time.Local)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("时间格式错误")
	}
	publicID, err := newUniquePublicID(ctx, s.repo.PublicIDExists)
	if err != nil {
		return nil, err
	}
	contest := &model.Contest{
		PublicID:    publicID,
		Name:        req.Name,
		Description: req.Description,
		CoverURL:    strings.TrimSpace(req.CoverURL),
		Type:        consts.ContestType(req.Type),
		StartTime:   startTime,
		EndTime:     startTime.Add(time.Duration(req.Duration) * time.Minute), // Duration 单位为分钟
		Duration:    req.Duration,
		InviteCode:  strings.TrimSpace(req.InviteCode),
		Rated:       req.Rated,
	}
	problems, err := s.resolveContestProblems(ctx, req.Problems)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateContestWithProblems(ctx, contest, problems); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	return &dto.CreateContestResp{
		ID: contest.PublicID,
	}, nil
}

// defaultProblemScore 满分兜底：未填或非正一律按 100（ACM 用不到，OI/IOI 保证有值）
func defaultProblemScore(s int) int {
	if s <= 0 {
		return 100
	}
	return s
}

func (s *ContestService) resolveContestProblems(ctx context.Context, inputs []dto.ContestProblemInput) ([]model.ContestProblem, error) {
	problems := make([]model.ContestProblem, 0, len(inputs))
	for idx, p := range inputs {
		// p.ProblemID 是对外题号，解析成内部主键存 contest_problems。
		pk, err := s.problemRepo.ResolveID(ctx, p.ProblemID)
		if err != nil {
			return nil, errcode.ErrInvalidParams.WithMsg(fmt.Sprintf("题号 %s 不存在", p.ProblemID))
		}
		problems = append(problems, model.ContestProblem{
			ProblemID: pk,
			Label:     contestProblemLabel(idx),
			Color:     p.Color,
			Score:     defaultProblemScore(p.Score),
		})
	}
	return problems, nil
}

func contestProblemLabel(index int) string {
	if index < 0 {
		return ""
	}
	label := ""
	for index >= 0 {
		label = string(rune('A'+index%26)) + label
		index = index/26 - 1
	}
	return label
}

func (s *ContestService) UpdateContest(ctx context.Context, id int64,
	form *dto.UpdateContestReq) (*dto.CreateContestResp, error) {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return nil, err
	}

	start := time.UnixMilli(form.StartTime)
	end := time.UnixMilli(form.EndTime)
	if !end.After(start) {
		return nil, errcode.ErrInvalidParams.WithMsg("结束时间必须晚于开始时间")
	}

	// 比赛一旦开始，赛制不可再改（ACM/OI/IOI/CF 计分逻辑不同，改了会让已判成绩/榜单错乱）
	started := time.Now().After(contest.StartTime)
	if started && consts.ContestType(form.Type) != contest.Type {
		return nil, errcode.ErrInvalidParams.WithMsg("比赛已开始，不能修改赛制")
	}

	contest.Name = form.Name
	contest.Description = form.Description
	contest.CoverURL = strings.TrimSpace(form.CoverURL)
	contest.Type = consts.ContestType(form.Type)
	contest.StartTime = start
	contest.EndTime = end
	contest.Duration = int(end.Sub(start).Minutes())
	contest.Rated = form.Rated
	problems, err := s.resolveContestProblems(ctx, form.Problems)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateContestWithProblems(ctx, contest, problems); err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	return &dto.CreateContestResp{ID: contest.PublicID}, nil
}

func contestStatus(c model.Contest) consts.ContestStatus {
	now := time.Now()
	if now.Before(c.StartTime) {
		return consts.ContestNotStarted
	}
	if now.Before(contestEndTime(c)) {
		return consts.ContestRunning
	}
	return consts.ContestFinished
}

func contestEndTime(contest model.Contest) time.Time {
	if !contest.EndTime.IsZero() {
		return contest.EndTime
	}
	return contest.StartTime.Add(time.Duration(contest.Duration) * time.Minute)
}

func (s *ContestService) getContest(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.repo.GetContestByID(ctx, contestID)
	if err == nil {
		return contest, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrContestNotFound
	}
	return nil, errcode.ErrDatabase.Wrap(err)
}

func (s *ContestService) ResolveID(ctx context.Context, publicID int64) (int64, error) {
	id, err := s.repo.ResolveIDByPublicID(ctx, publicID)
	if err == nil {
		return id, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrContestNotFound
	}
	return 0, errcode.ErrDatabase.Wrap(err)
}

func (s *ContestService) ResolveProblemID(ctx context.Context, displayID string) (int64, error) {
	id, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(displayID))
	if err != nil {
		return 0, errcode.ErrProblemNotFound
	}
	return id, nil
}

func contestProblemAccessPolicy(contest *model.Contest, registered, isAdmin bool, now time.Time) error {
	if isAdmin {
		return nil
	}
	if now.Before(contest.StartTime) {
		return errcode.ErrContestNotStart
	}
	if !registered {
		return errcode.ErrContestNotRegistered
	}
	return nil
}

func contestSubmissionPolicy(contest *model.Contest, registered bool, now time.Time) error {
	if now.Before(contest.StartTime) {
		return errcode.ErrContestNotStart
	}
	if !now.Before(contestEndTime(*contest)) {
		return errcode.ErrContestFinished
	}
	if !registered {
		return errcode.ErrContestNotRegistered
	}
	return nil
}

func (s *ContestService) authorizeContestProblemAccess(
	ctx context.Context,
	contestID, userID int64,
	isAdmin bool,
) (*model.Contest, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		return contest, nil
	}
	registered, err := s.repo.ExistByID(ctx, contestID, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if err := contestProblemAccessPolicy(contest, registered, false, time.Now()); err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *ContestService) authorizeContestSubmission(
	ctx context.Context,
	contestID, userID int64,
) (*model.Contest, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	registered, err := s.repo.ExistByID(ctx, contestID, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if err := contestSubmissionPolicy(contest, registered, time.Now()); err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *ContestService) ListContests(ctx context.Context, userID int64, page, pageSize int, keyword string, ctype int, status *int) (*dto.ContestListResp, error) {
	contests, total, err := s.repo.ListContests(ctx, page, pageSize, keyword, ctype, status)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	unStartContest := make([]model.Contest, 0, len(contests))
	runningContest := make([]model.Contest, 0, len(contests))
	endContest := make([]model.Contest, 0, len(contests))

	sort.Slice(contests, func(i, j int) bool {
		return contests[i].StartTime.Before(contests[j].StartTime)
	})

	now := time.Now()
	for i := 0; i < len(contests); i++ {
		// 未开始
		if now.Before(contests[i].StartTime) {
			unStartContest = append(unStartContest, contests[i])
		} else if !now.Before(contestEndTime(contests[i])) {
			endContest = append(endContest, contests[i])
		} else {
			runningContest = append(runningContest, contests[i])
		}
	}

	// 进行中 → 未开始 → 已结束，同组内按开始时间升序

	contests = contests[:0]

	contests = append(contests, runningContest...)
	contests = append(contests, unStartContest...)
	contests = append(contests, endContest...)

	contestIDs := make([]int64, 0, len(contests))
	for _, c := range contests {
		contestIDs = append(contestIDs, c.ID)
	}
	userCountMap, err := s.repo.GetContestUserCounts(ctx, contestIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	registeredMap, err := s.repo.GetRegisteredContestIDs(ctx, userID, contestIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	resp := &dto.ContestListResp{
		Total: total,
		List:  make([]dto.ContestItem, 0, len(contests)),
	}
	for _, c := range contests {
		count := userCountMap[c.ID]
		registered := registeredMap[c.ID]
		resp.List = append(resp.List, dto.ContestItem{
			ID:             c.PublicID,
			Name:           c.Name,
			Description:    c.Description,
			CoverURL:       c.CoverURL,
			StartTime:      c.StartTime,
			EndTime:        contestEndTime(c),
			Type:           c.Type.String(),
			Status:         contestStatus(c),
			Rated:          c.Rated,
			Duration:       c.Duration,
			IsRegistered:   registered,
			NeedInviteCode: c.InviteCode != "",
			Participants:   int(count),
		})
	}
	return resp, nil
}

func (s *ContestService) JoinContest(ctx context.Context, userID, contestID int64, inviteCode string) error {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return err
	}
	if !time.Now().Before(contestEndTime(*contest)) {
		return errcode.ErrContestFinished
	}
	// 邀请码比赛：必须提供且匹配
	if contest.InviteCode != "" && strings.TrimSpace(inviteCode) != contest.InviteCode {
		return errcode.ErrContestInviteCodeInvalid
	}
	// 防重复报名
	exists, err := s.repo.ExistByID(ctx, contestID, userID)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	if exists {
		return errcode.ErrContestAlreadyJoined
	}
	if err := s.repo.CreateContestUser(ctx, userID, contestID); err != nil {
		// 唯一索引竞争时，尽量把重复报名稳定映射为业务错误。
		if exists, checkErr := s.repo.ExistByID(ctx, contestID, userID); checkErr == nil && exists {
			return errcode.ErrContestAlreadyJoined
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *ContestService) GetContestDesc(ctx context.Context, id int64) (*dto.ContestDesc, error) {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := &dto.ContestDesc{
		Name:        contest.Name,
		Description: contest.Description,
		StartTime:   contest.StartTime,
		EndTime:     contestEndTime(*contest),
	}
	return resp, nil
}

// GetContestDetail userID=0 表示未登录场景
func (s *ContestService) GetContestDetail(ctx context.Context, id int64, userID int64) (*dto.ContestDetailResp, error) {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return nil, err
	}
	registered := false
	if userID > 0 {
		registered, err = s.repo.ExistByID(ctx, id, userID)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
	}
	resp := &dto.ContestDetailResp{
		Name:           contest.Name,
		StartTime:      contest.StartTime.UnixMilli(),
		EndTime:        contestEndTime(*contest).UnixMilli(),
		IsRegistered:   registered,
		NeedInviteCode: contest.InviteCode != "",
		Rule:           contest.Type.String(),
		Rated:          contest.Rated,
		Settled:        contest.Settled,
	}
	return resp, nil
}

func (s *ContestService) GetContestProblemList(ctx context.Context, contestID, userID int64, isAdmin bool) (*dto.ContestProblemListResp, error) {
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

	resp := &dto.ContestProblemListResp{
		ProblemList: make([]dto.ContestProblem, 0, len(problems)),
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

		resp.ProblemList = append(resp.ProblemList, dto.ContestProblem{
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

func (s *ContestService) EditContest(ctx context.Context, contestID int64) (*dto.EditContestResp, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	contestProblems, err := s.repo.GetContestProblemsByContestID(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.EditContestResp{
		Name:        contest.Name,
		Description: contest.Description,
		CoverURL:    contest.CoverURL,
		Type:        int(contest.Type),
		StartTime:   contest.StartTime.UnixMilli(),
		EndTime:     contestEndTime(*contest).UnixMilli(),
		ProblemList: make([]dto.EasyInfo, 0, len(contestProblems)),
		Rated:       contest.Rated,
	}

	sort.Slice(contestProblems, func(i, j int) bool {
		return contestProblems[i].Label < contestProblems[j].Label
	})
	problemMap := s.problemsByIDs(ctx, contestProblems)
	for _, p := range contestProblems {
		pr := problemMap[p.ProblemID]
		resp.ProblemList = append(resp.ProblemList, dto.EasyInfo{
			ID:    pr.DisplayID, // 对外题号（编辑表单据此回显 + 重新提交）
			Name:  pr.Name,
			Color: p.Color,
			Score: p.Score,
		})
	}

	return resp, nil
}

// problemNamesByIDs 一次性取出比赛各题的题名，避免循环里逐题查库（N+1）
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
) (*dto.ContestProblemResp, error) {
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

	resp := &dto.ContestProblemResp{
		Name:         label + ". " + problem.Name,
		TimeLimit:    problem.TimeLimit,
		MemoryLimit:  problem.MemoryLimit,
		Description:  problem.Description,
		InputFormat:  problem.InputFormat,
		OutputFormat: problem.OutputFormat,
		Hint:         problem.Hint,
		Samples:      make([]dto.ProblemSample, 0),
		HasTestData:  hasTestData,
	}
	samples, err := s.problemRepo.GetProblemSamplesByID(ctx, problem.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	for _, sample := range samples {
		resp.Samples = append(resp.Samples, dto.ProblemSample{
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
) (*dto.ContestSubmitResp, error) {
	if _, err := s.authorizeContestProblemAccess(ctx, contestID, userID, isAdmin); err != nil {
		return nil, err
	}
	problem, err := s.repo.GetContestProblems(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.ContestSubmitResp{
		ProblemList: make([]dto.ContestProblemItem, 0, len(problem)),
	}
	problemNames := s.problemNamesByIDs(ctx, problem)
	for _, p := range problem {
		resp.ProblemList = append(resp.ProblemList, dto.ContestProblemItem{
			Name:  p.Label + ". " + problemNames[p.ProblemID],
			Label: p.Label,
		})
	}
	return resp, nil
}

func (s *ContestService) SubmitContestProblem(ctx context.Context, form *dto.ContestSubmitReq, userID int64) (int64, error) {
	if _, err := s.authorizeContestSubmission(ctx, form.ContestID, userID); err != nil {
		return 0, err
	}
	if err := submitRateLimit(ctx, s.cache, userID); err != nil {
		return 0, err
	}
	contestProblem, err := s.repo.GetContestProblem(ctx, model.ContestProblem{
		ContestID: form.ContestID,
		Label:     form.Label,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrContestProblemNotFound
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	problem, err := s.problemRepo.GetByID(ctx, contestProblem.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrContestProblemNotFound
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	if err := requireProblemTestData(ctx, s.testData, problem); err != nil {
		return 0, err
	}

	language, err := s.languages.ResolveEnabledName(ctx, form.Language)
	if err != nil {
		return 0, err
	}
	publicID, err := newUniquePublicID(ctx, s.submitRepo.PublicIDExists)
	if err != nil {
		return 0, err
	}

	submission := &model.Submission{
		PublicID:  publicID,
		UserID:    userID,
		ProblemID: contestProblem.ProblemID,
		Code:      form.Code,
		Status:    judge.Pending,
		Language:  language,
		ContestID: form.ContestID,
		CreatedAt: time.Now(),
	}
	dispatch, err := s.submitRepo.CreatePending(ctx, submission)
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	if err := s.mq.EnqueueSubmission(ctx, dispatch.SubmissionID, dispatch.Version); err != nil {
		return 0, errcode.ErrInternal.WithMsg("提交入队失败").Wrap(err)
	}
	if err := s.submitRepo.MarkDispatchPublished(ctx, dispatch); err != nil {
		logger.Warnw("mark contest submission dispatch published failed",
			"submissionID", dispatch.SubmissionID, "version", dispatch.Version, "err", err)
	}
	return submission.PublicID, nil
}

// GetContestSubmissions 比赛提交列表。
//   - 普通用户：只看自己（忽略筛选）。
//   - 管理员：看全场，支持按用户名/题目/结果筛选。
func (s *ContestService) GetContestSubmissions(ctx context.Context, contestID, userID int64, isAdmin bool, q *dto.ContestSubmissionQuery) (*dto.ContestSubmissionListResp, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	sq := &repository.SubmissionQuery{ContestID: contestID}
	if isAdmin {
		if q != nil {
			if q.Username != "" {
				name := q.Username
				sq.Username = &name
			}
			if strings.TrimSpace(q.ProblemID) != "" {
				pid, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(q.ProblemID))
				if err != nil {
					return nil, errcode.ErrProblemNotFound
				}
				sq.ProblemID = &pid
			}
			if q.Status != nil {
				sq.Status = q.Status
			}
		}
	} else {
		sq.UserID = &userID // 普通用户只看自己
	}

	submissions, err := s.submitRepo.GetContestSubmissionListWithInfo(ctx, sq)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.ContestSubmissionListResp{
		Total:       int64(len(submissions)),
		Submissions: make([]dto.ContestSubmissionItem, 0, len(submissions)),
	}
	for _, sub := range submissions {
		resp.Submissions = append(resp.Submissions, dto.ContestSubmissionItem{
			ID:          sub.PublicID,
			ProblemID:   sub.ProblemDisplayID,
			ProblemName: sub.ProblemLabel + ". " + sub.ProblemName,
			UserID:      sub.UserUID, // 对外用户号
			UserName:    sub.UserName,
			Rating:      sub.Rating,
			Result:      sub.Status,
			Language:    sub.Language,
			TimeUsed:    sub.TimeUsed,
			MemoryUsed:  sub.MemoryUsed,
			CreatedAt:   util.FormatDurationHHMMSS(sub.CreatedAt, contest.StartTime),
		})
	}
	return resp, nil
}

// markSelf 按当前查看者标记自己那一行（isSelf 不进缓存，按查看者 uid 临时标；
// 榜单里 User.ID 已是对外 uid，故这里也用 uid 比较）
func markSelf(resp *dto.ContestRankResp, viewerUID int64) {
	if viewerUID == 0 {
		return
	}
	for i := range resp.RankList {
		resp.RankList[i].IsSelf = resp.RankList[i].User.ID == viewerUID
	}
}

// uidOf 内部主键 → 对外用户号（0=未登录或查不到）
func (s *ContestService) uidOf(ctx context.Context, userPK int64) int64 {
	if userPK <= 0 {
		return 0
	}
	if u, err := s.userRepo.GetUserByID(ctx, userPK); err == nil {
		return u.UID
	}
	return 0
}

// GetMyContestRank 当前用户在某场比赛的名次（没参与返回 Found=false）
func (s *ContestService) GetMyContestRank(ctx context.Context, contestID, userID int64) (*dto.MyContestRankResp, error) {
	rank, err := s.GetContestRank(ctx, contestID, userID)
	if err != nil {
		return nil, err
	}
	for _, item := range rank.RankList {
		if item.IsSelf { // GetContestRank 已按查看者标好 isSelf
			return &dto.MyContestRankResp{
				Found:      true,
				Rank:       item.Rank,
				PassCount:  item.PassCount,
				Penalty:    item.Penalty,
				TotalScore: item.TotalScore,
				Rule:       rank.Contest.Rule,
			}, nil
		}
	}
	return &dto.MyContestRankResp{Found: false, Rule: rank.Contest.Rule}, nil
}

type standingEntry struct {
	UserID int64
	Rank   int
}

// settleRatingIfNeeded 惰性结算：rated + 已结束 + 未结算时，按最终名次算 CF rating 落库（幂等）。
func (s *ContestService) settleRatingIfNeeded(ctx context.Context, contestID int64) {
	contest, err := s.repo.GetContestByID(ctx, contestID)
	if err != nil || !contest.Rated || contest.Settled {
		return
	}
	if contestStatus(*contest) != consts.ContestFinished {
		return
	}
	ranks, err := s.contestStandings(ctx, contest)
	if err != nil || len(ranks) < 2 {
		return // 少于 2 名参赛者不计分
	}
	userIDs := make([]int64, 0, len(ranks))
	for _, e := range ranks {
		userIDs = append(userIDs, e.UserID)
	}

	type note struct {
		userID  int64
		old, nw int
		delta   int
	}
	var settled []note

	err = s.repo.SettleTx(ctx, contestID, func(tx *gorm.DB, c *model.Contest) error {
		if c.Settled { // 并发下别人已抢先结算
			return nil
		}
		var users []model.User
		if err := tx.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
		curRating := make(map[int64]int, len(users))
		curMax := make(map[int64]int, len(users))
		for _, u := range users {
			curRating[u.ID] = u.Rating
			curMax[u.ID] = u.MaxRating
		}

		players := make([]rating.Player, len(ranks))
		for i, e := range ranks {
			r := curRating[e.UserID]
			if r == 0 {
				r = initialRating
			}
			players[i] = rating.Player{Rating: r, Rank: e.Rank}
		}
		results := rating.Calculate(players)

		for i, e := range ranks {
			old := players[i].Rating
			nw := results[i].NewRating
			if err := tx.Create(&model.RatingChange{
				ContestID: contestID, UserID: e.UserID, Rank: e.Rank,
				OldRating: old, NewRating: nw, Delta: results[i].Delta,
			}).Error; err != nil {
				return err
			}
			maxr := nw
			if curMax[e.UserID] > maxr {
				maxr = curMax[e.UserID]
			}
			if err := tx.Model(&model.User{}).Where("id = ?", e.UserID).
				Updates(map[string]any{"rating": nw, "max_rating": maxr}).Error; err != nil {
				return err
			}
			settled = append(settled, note{userID: e.UserID, old: old, nw: nw, delta: results[i].Delta})
		}
		return tx.Model(&model.Contest{}).Where("id = ?", contestID).Update("settled", true).Error
	})
	if err != nil {
		logger.S().Warnw("settle rating failed", "contestID", contestID, "err", err)
		return
	}
	_ = s.cache.Delete(ctx, cache.ContestRank(contestID))

	// 结算后给每个参赛者发 rating 变化通知
	for _, n := range settled {
		s.notify.Notify(ctx, n.userID, 0, "rating",
			"Rating 结算 · "+contest.Name,
			fmt.Sprintf("%d → %d (%+d)", n.old, n.nw, n.delta),
			fmt.Sprintf("/contest/%d/rank", contest.PublicID), "contest", contestID)
	}
}

// contestStandings 计算比赛最终名次（仅有提交者，并列共享名次），供 rating 结算用。
// buildContestUCP 从一组「同一 user+problem、时间升序、已排除 Pending/CE」的提交明细，
// 重算出该 (contest,user,problem) 的聚合行。规则与 worker 的增量逻辑一一对应，保证结果一致：
//   - ACM / CF：Status/AcTime 取首个 AC；UnAcCount = 首 AC 之前的非 AC(非 CE)次数(罚时)；首 AC 后冻结。
//   - OI：取最后一次提交的 Score/Status/时间。
//   - IOI：取历史最高 Score 及其达成时间；Status 取最后一次。
//
// subs 不能为空。
func buildContestUCP(contestType consts.ContestType, contestID, userID, problemID int64, subs []repository.RecomputeSubmissionRow) model.UserContestProblem {
	ucp := model.UserContestProblem{
		ContestID: contestID,
		UserID:    userID,
		ProblemID: problemID,
		SubCount:  len(subs),
	}
	switch contestType {
	case consts.ContestOI:
		for _, r := range subs {
			if r.Status == judge.Accepted {
				ucp.AcCount++
			} else {
				ucp.UnAcCount++
			}
		}
		last := subs[len(subs)-1]
		ucp.Status = last.Status
		ucp.Score = last.Score
		t := last.CreatedAt
		ucp.AcTime = &t
	case consts.ContestIOI:
		maxScore := -1
		var maxTime time.Time
		for _, r := range subs {
			if r.Status == judge.Accepted {
				ucp.AcCount++
			} else {
				ucp.UnAcCount++
			}
			if r.Score > maxScore {
				maxScore = r.Score
				maxTime = r.CreatedAt
			}
		}
		if maxScore < 0 {
			maxScore = 0
		}
		ucp.Score = maxScore
		ucp.Status = subs[len(subs)-1].Status
		mt := maxTime
		ucp.AcTime = &mt
	default: // ACM / CF
		for _, r := range subs {
			if ucp.Status == judge.Accepted {
				break // 首个 AC 后冻结：罚时/状态不再变
			}
			if r.Status == judge.Accepted {
				ucp.Status = judge.Accepted
				ucp.AcCount = 1
				t := r.CreatedAt
				ucp.AcTime = &t
			} else {
				ucp.Status = r.Status
				ucp.UnAcCount++
			}
		}
	}
	return ucp
}

// computeContestUserProblems 路①（对齐 HOJ）：从 submissions 明细现算某比赛的全部 (user,problem) 聚合，不落库。
// 榜单 / 结算 / 重算都以它为唯一真值来源 —— 天然反映最新判题结果（含重判后），无需任何"重算触发"。
func (s *ContestService) computeContestUserProblems(ctx context.Context, contest *model.Contest) ([]model.UserContestProblem, error) {
	rows, err := s.submitRepo.ListContestSubmissionsForRecompute(ctx, contest.ID)
	if err != nil {
		return nil, err
	}
	// 按 (user,problem) 分组（rows 已按 user,problem,time 升序）
	type key struct{ u, p int64 }
	groups := make(map[key][]repository.RecomputeSubmissionRow)
	order := make([]key, 0)
	for _, r := range rows {
		k := key{r.UserID, r.ProblemID}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], r)
	}
	ucps := make([]model.UserContestProblem, 0, len(order))
	for _, k := range order {
		ucps = append(ucps, buildContestUCP(contest.Type, contest.ID, k.u, k.p, groups[k]))
	}
	return ucps, nil
}

// RecomputeContest 把现算结果物化落库到 UserContestProblem，并清相关缓存。
// 路①下榜单已直接现算、不依赖此表；保留该接口用于手动重建/导出/排查（admin 触发）。
func (s *ContestService) RecomputeContest(ctx context.Context, contestID int64) error {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return err
	}
	ucps, err := s.computeContestUserProblems(ctx, contest)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	if err := s.repo.ReplaceContestUserProblems(ctx, contestID, ucps); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	s.invalidateContestCaches(ctx, contestID, ucps)
	return nil
}

// invalidateContestCaches 清掉比赛的 rank 缓存 + 各覆盖缓存，让读取回落到刚重算好的 DB。
func (s *ContestService) invalidateContestCaches(ctx context.Context, contestID int64, ucps []model.UserContestProblem) {
	_ = s.cache.Delete(ctx, cache.ContestRank(contestID))
	problems := make(map[int64]struct{})
	users := make(map[int64]struct{})
	for _, u := range ucps {
		problems[u.ProblemID] = struct{}{}
		users[u.UserID] = struct{}{}
		_ = s.cache.Delete(ctx, cache.ContestUserProblem(contestID, u.UserID, u.ProblemID))
	}
	for pid := range problems {
		_ = s.cache.Delete(ctx, cache.ContestProblemStats(contestID, pid))
		_ = s.cache.Delete(ctx, cache.ContestProblemSubmittedUsers(contestID, pid))
		_ = s.cache.Delete(ctx, cache.ContestProblemAcceptedUsers(contestID, pid))
	}
	for uid := range users {
		_ = s.cache.Delete(ctx, cache.ContestUserStats(contestID, uid))
		_ = s.cache.Delete(ctx, cache.ContestUserAcceptedProblems(contestID, uid))
	}
}

// RecomputeUserProblem 从 submissions 明细重算某用户某题的非比赛做题聚合(UserProblem)，并失效用户榜缓存。
func (s *ContestService) RecomputeUserProblem(ctx context.Context, userID, problemID int64) error {
	subs, err := s.submitRepo.ListUserProblemSubmissionsForRecompute(ctx, userID, problemID)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	if len(subs) == 0 {
		if err := s.problemRepo.DeleteUserProblem(ctx, userID, problemID); err != nil {
			return errcode.ErrDatabase.Wrap(err)
		}
		_, _ = s.cache.Incr(ctx, cache.UserRankGen())
		return nil
	}
	up := model.UserProblem{
		UserID:      userID,
		ProblemID:   problemID,
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
		up.Status = subs[0].Status // 从未 AC：沿用首次提交状态（对齐现有增量逻辑）
	}
	if err := s.problemRepo.UpsertUserProblem(ctx, &up); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	_, _ = s.cache.Incr(ctx, cache.UserRankGen())
	return nil
}

func (s *ContestService) contestStandings(ctx context.Context, contest *model.Contest) ([]standingEntry, error) {
	contestUsers, err := s.repo.GetContestUserByContestID(ctx, contest.ID)
	if err != nil {
		return nil, err
	}
	records, err := s.computeContestUserProblems(ctx, contest)
	if err != nil {
		return nil, err
	}
	byUser := make(map[int64][]model.UserContestProblem)
	for _, r := range records {
		byUser[r.UserID] = append(byUser[r.UserID], r)
	}
	acm := contest.Type != consts.ContestOI && contest.Type != consts.ContestIOI && contest.Type != consts.ContestCF
	isCF := contest.Type == consts.ContestCF
	cfDur := int(contestEndTime(*contest).Sub(contest.StartTime).Minutes())
	cpScore := map[int64]int{} // CF：题目初始分 x
	if isCF {
		if cps, e := s.repo.GetContestProblemsByContestID(ctx, contest.ID); e == nil {
			for _, cp := range cps {
				cpScore[cp.ProblemID] = cp.Score
			}
		}
	}

	type row struct {
		userID   int64
		pass     int
		penalty  int
		score    int
		lastTime int64
	}
	rows := make([]row, 0, len(contestUsers))
	for _, cu := range contestUsers {
		recs := byUser[cu.UserID]
		if len(recs) == 0 {
			continue // 只算有提交者
		}
		rr := row{userID: cu.UserID}
		for _, rec := range recs {
			if acm {
				if rec.Status == judge.Accepted && rec.AcTime != nil {
					rr.pass++
					rr.penalty += int(rec.AcTime.Sub(contest.StartTime).Minutes()) + 20*rec.UnAcCount
				}
			} else if isCF {
				accepted := rec.Status == judge.Accepted && rec.AcTime != nil
				t := 0
				if rec.AcTime != nil {
					t = int(rec.AcTime.Sub(contest.StartTime).Minutes())
				}
				rr.score += scoring.CFScore(cpScore[rec.ProblemID], t, cfDur, rec.UnAcCount, accepted)
				if accepted && rec.AcTime.Unix() > rr.lastTime {
					rr.lastTime = rec.AcTime.Unix()
				}
			} else {
				rr.score += rec.Score
				if rec.AcTime != nil && rec.AcTime.Unix() > rr.lastTime {
					rr.lastTime = rec.AcTime.Unix()
				}
			}
		}
		rows = append(rows, rr)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if acm {
			if rows[i].pass != rows[j].pass {
				return rows[i].pass > rows[j].pass
			}
			return rows[i].penalty < rows[j].penalty
		}
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return rows[i].lastTime < rows[j].lastTime
	})

	out := make([]standingEntry, len(rows))
	for i := range rows {
		rank := i + 1
		if i > 0 {
			var eq bool
			if acm {
				eq = rows[i].pass == rows[i-1].pass && rows[i].penalty == rows[i-1].penalty
			} else {
				eq = rows[i].score == rows[i-1].score && rows[i].lastTime == rows[i-1].lastTime
			}
			if eq {
				rank = out[i-1].Rank
			}
		}
		out[i] = standingEntry{UserID: rows[i].userID, Rank: rank}
	}
	return out, nil
}

func (s *ContestService) GetContestRank(ctx context.Context, contestID, userID int64) (*dto.ContestRankResp, error) {
	s.settleRatingIfNeeded(ctx, contestID)
	viewerUID := s.uidOf(ctx, userID) // 查看者的对外用户号，用于标记 isSelf
	cacheKey := cache.ContestRank(contestID)
	// 命中缓存直接返回（轮询期间不打 DB）；isSelf 按查看者临时标，不进缓存
	var cachedResp dto.ContestRankResp
	if ok, err := s.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil && ok {
		markSelf(&cachedResp, viewerUID)
		return &cachedResp, nil
	}

	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	contestProblems, err := s.repo.GetContestProblemsByContestID(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemIDs := make([]int64, 0, len(contestProblems))
	for _, problem := range contestProblems {
		problemIDs = append(problemIDs, problem.ProblemID)
	}

	sort.Slice(contestProblems, func(i, j int) bool {
		return contestProblems[i].Label < contestProblems[j].Label
	})

	contestUsers, err := s.repo.GetContestUserByContestID(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userIDs := make([]int64, 0, len(contestUsers))
	for _, user := range contestUsers {
		userIDs = append(userIDs, user.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userMap := make(map[int64]model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	records, err := s.computeContestUserProblems(ctx, contest)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	recordMap := make(map[int64]map[int64]model.UserContestProblem, len(contestUsers))
	firstBloodMap := map[int64]int64{}
	firstBloodTime := map[int64]time.Time{}
	for _, record := range records {
		if recordMap[record.UserID] == nil {
			recordMap[record.UserID] = map[int64]model.UserContestProblem{}
		}
		recordMap[record.UserID][record.ProblemID] = record
		if record.Status == judge.Accepted && record.AcTime != nil {
			if t, ok := firstBloodTime[record.ProblemID]; !ok || record.AcTime.Before(t) {
				firstBloodTime[record.ProblemID] = *record.AcTime
				firstBloodMap[record.ProblemID] = record.UserID
			}
		}
	}

	resp := &dto.ContestRankResp{
		Contest: dto.Contest{
			ID:     contest.PublicID,
			Title:  contest.Name,
			Status: int(contestStatus(*contest)),
		},
		Problems: make([]dto.ProblemLabel, 0, len(contestProblems)),
		RankList: make([]dto.ContestRankItem, 0, len(contestUsers)),
	}

	resp.Contest.Rule = contest.Type.String()
	for i := 0; i < len(contestProblems); i++ {
		resp.Problems = append(resp.Problems, dto.ProblemLabel{
			Label:    contestProblems[i].Label,
			Color:    contestProblems[i].Color,
			MaxScore: contestProblems[i].Score,
		})
	}

	if contest.Type != consts.ContestOI && contest.Type != consts.ContestIOI && contest.Type != consts.ContestCF {
		// ===== ACM（含未知/历史 type）：通过数 + 罚时 =====
		for _, user := range contestUsers {
			u := userMap[user.UserID]
			item := dto.ContestRankItem{
				Problems: make([]dto.ContestProblemStatus, 0, len(contestProblems)),
			}
			item.User.ID = u.UID // 对外用户号
			item.User.Name = u.Username
			item.User.Avatar = u.Avatar
			item.User.Rating = u.Rating

			for index, problem := range contestProblems {
				// 路①：records 已是从 submissions 现算的最新真值，无需再叠 Redis 覆盖缓存。
				userContestProblem, ok := recordMap[user.UserID][problem.ProblemID]
				if ok {
					item.Problems = append(item.Problems, dto.ContestProblemStatus{
						Label:  problem.Label,
						Status: userContestProblem.Status,
						Tries:  userContestProblem.UnAcCount,
					})
					if userContestProblem.UserID == firstBloodMap[problem.ProblemID] {
						item.Problems[index].FirstBlood = 1
					}
					if userContestProblem.Status == judge.Accepted && userContestProblem.AcTime != nil {
						diff := userContestProblem.AcTime.Sub(contest.StartTime)
						hours := int(diff.Hours())
						minutes := int(diff.Minutes()) % 60
						timeStr := fmt.Sprintf("%02d:%02d", hours, minutes)
						item.PassCount++
						item.Problems[index].AcTime = &timeStr
						item.Penalty += int(userContestProblem.AcTime.Sub(contest.StartTime).Minutes()) + 20*userContestProblem.UnAcCount
					}
				} else {
					item.Problems = append(item.Problems, dto.ContestProblemStatus{
						Label:  problem.Label,
						Status: 0,
						Tries:  0,
					})
				}
			}
			// PassCount/Penalty 直接用上面从 DB(recordMap) 算出的值；去掉 ContestUserStats 累加缓存覆盖(无 TTL、并发/重复消费会多加,漂移后永久污染榜单)。
			resp.RankList = append(resp.RankList, item)
		}

		sort.Slice(resp.RankList, func(i, j int) bool {
			if resp.RankList[i].PassCount != resp.RankList[j].PassCount {
				return resp.RankList[i].PassCount > resp.RankList[j].PassCount
			}
			return resp.RankList[i].Penalty < resp.RankList[j].Penalty
		})
	} else {
		// ===== OI / IOI：每题得分 + 总分（IOI 取最高分、OI 取最后一次，已在判题时落库）=====
		// OI 赛中封榜：只列参赛者、不公布分数，结束后再放开
		frozen := contest.Type == consts.ContestOI && contestStatus(*contest) == consts.ContestRunning
		resp.Contest.Frozen = frozen
		isCF := contest.Type == consts.ContestCF
		cfDur := int(contestEndTime(*contest).Sub(contest.StartTime).Minutes()) // CF 动态分用的赛长（分钟）

		type scoreRow struct {
			item dto.ContestRankItem
			last time.Time
		}
		rows := make([]scoreRow, 0, len(contestUsers))
		for _, user := range contestUsers {
			u := userMap[user.UserID]
			item := dto.ContestRankItem{
				Problems: make([]dto.ContestProblemStatus, 0, len(contestProblems)),
			}
			item.User.ID = u.UID // 对外用户号
			item.User.Name = u.Username
			item.User.Avatar = u.Avatar
			item.User.Rating = u.Rating

			var last time.Time
			for _, problem := range contestProblems {
				cell := dto.ContestProblemStatus{Label: problem.Label}
				if !frozen {
					score := 0
					if ucp, ok := recordMap[user.UserID][problem.ProblemID]; ok {
						cell.Status = ucp.Status
						if isCF {
							accepted := ucp.Status == judge.Accepted && ucp.AcTime != nil
							t := 0
							if ucp.AcTime != nil {
								t = int(ucp.AcTime.Sub(contest.StartTime).Minutes())
							}
							score = scoring.CFScore(problem.Score, t, cfDur, ucp.UnAcCount, accepted)
							if accepted {
								// CF：分数下面展示 AC 时间（相对开赛 HH:MM）
								diff := ucp.AcTime.Sub(contest.StartTime)
								ts := fmt.Sprintf("%02d:%02d", int(diff.Hours()), int(diff.Minutes())%60)
								cell.AcTime = &ts
								if ucp.AcTime.After(last) {
									last = *ucp.AcTime
								}
							}
						} else {
							score = ucp.Score
							if ucp.AcTime != nil && ucp.AcTime.After(last) {
								last = *ucp.AcTime
							}
						}
					}
					cell.Score = &score
					item.TotalScore += score
				}
				item.Problems = append(item.Problems, cell)
			}
			rows = append(rows, scoreRow{item: item, last: last})
		}

		if !frozen {
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].item.TotalScore != rows[j].item.TotalScore {
					return rows[i].item.TotalScore > rows[j].item.TotalScore
				}
				return rows[i].last.Before(rows[j].last) // 同分早达者靠前
			})
		}
		for i := range rows {
			resp.RankList = append(resp.RankList, rows[i].item)
		}
	}

	for i := 0; i < len(resp.RankList); i++ {
		resp.RankList[i].Rank = i + 1
	}

	// rated 且已结算：给每行附上本场 rating 变化（±delta / 新分）
	if contest.Rated && contest.Settled {
		if changes, err := s.repo.GetRatingChangesByContest(ctx, contestID); err == nil {
			byUser := make(map[int64]model.RatingChange, len(changes))
			for _, c := range changes {
				byUser[c.UserID] = c
			}
			for i := range resp.RankList {
				if rc, ok := byUser[resp.RankList[i].User.ID]; ok {
					d, nr := rc.Delta, rc.NewRating
					resp.RankList[i].RatingDelta = &d
					resp.RankList[i].NewRating = &nr
				}
			}
		}
	}

	// 写缓存（短 TTL，存的是不含 isSelf 的基准榜）；失败不影响返回
	_ = s.cache.SetJSON(ctx, cacheKey, resp, contestRankTTL)

	markSelf(resp, viewerUID)
	return resp, nil
}
