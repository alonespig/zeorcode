package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/common/logger"
	"zoj/internal/dto"
	"zoj/internal/infra/cache"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

const (
	nsPerMs        = 1_000_000
	bytesPerKB     = 1024
	submitInterval = 3 * time.Second
)

type SubmissionService struct {
	subRepo   SubmissionStore
	userRepo  SubmissionUserReader
	proRepo   SubmissionProblemReader
	testData  ProblemTestDataReader
	cache     SubmissionCache
	mq        SubmissionQueue
	languages LanguageResolver
}

// SubmissionStore 由提交用例定义，只暴露当前用例实际使用的持久化能力。
// 具体的 GORM repository 通过 bootstrap 注入并隐式实现该接口。
type SubmissionStore interface {
	PublicIDExists(ctx context.Context, publicID int64) (bool, error)
	CreatePending(ctx context.Context, submission *model.Submission) (repository.SubmissionDispatch, error)
	GetRejudgeTargets(ctx context.Context, scope repository.RejudgeScope) ([]repository.RejudgeTarget, error)
	ResetForRejudge(ctx context.Context, ids []int64) ([]repository.SubmissionDispatch, error)
	MarkDispatchPublished(ctx context.Context, dispatch repository.SubmissionDispatch) error
	ListWithInfo(ctx context.Context, page, pageSize int, status *int, username string, userID, problemID int64, includeHidden bool) ([]repository.SubmissionListItem, int64, error)
	ResolveIDByPublicID(ctx context.Context, publicID int64) (int64, error)
	GetSubmissionByPublicID(ctx context.Context, publicID int64) (*model.Submission, error)
	GetCaseResultBySubID(ctx context.Context, submissionID int64) ([]model.JudgeResult, error)
	GetUserRecentAcceptedSubmissions(ctx context.Context, userID int64, start time.Time) ([]model.Submission, error)
}

type SubmissionUserReader interface {
	ResolveIDByUID(ctx context.Context, uid int64) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
}

type SubmissionProblemReader interface {
	GetByDisplayID(ctx context.Context, displayID string) (*model.Problem, error)
	ResolveID(ctx context.Context, displayID string) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Problem, error)
}

type SubmissionQueue interface {
	EnqueueSubmission(ctx context.Context, submissionID int64, version int) error
}

type SubmissionCache interface {
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)
	Delete(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) (int64, error)
}

func NewSubmissionService(subRepo SubmissionStore,
	userRepo SubmissionUserReader,
	proRepo SubmissionProblemReader,
	testData ProblemTestDataReader,
	cache SubmissionCache,
	mq SubmissionQueue,
	languages LanguageResolver) *SubmissionService {
	return &SubmissionService{
		subRepo:   subRepo,
		userRepo:  userRepo,
		proRepo:   proRepo,
		testData:  testData,
		cache:     cache,
		mq:        mq,
		languages: languages,
	}
}

func (s *SubmissionService) CreateSubmission(ctx context.Context, req *dto.SubmitCodeReq, userID int64, isAdmin bool) (int64, error) {
	if err := submitRateLimit(ctx, s.cache, userID); err != nil {
		return 0, err
	}
	// req.ProblemID 是对外题号，解析成内部题目；隐藏题非管理员不能提交（当作不存在）
	problem, err := s.proRepo.GetByDisplayID(ctx, req.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrProblemNotFound
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	if problem.Hidden && !isAdmin {
		return 0, errcode.ErrProblemNotFound
	}
	if err := requireProblemTestData(ctx, s.testData, problem); err != nil {
		return 0, err
	}
	language, err := s.languages.ResolveEnabledName(ctx, req.Language)
	if err != nil {
		return 0, err
	}
	publicID, err := newUniquePublicID(ctx, s.subRepo.PublicIDExists)
	if err != nil {
		return 0, err
	}
	submission := model.Submission{
		PublicID:  publicID,
		ProblemID: problem.ID, // 存内部主键
		UserID:    userID,
		Language:  language,
		Code:      req.Code,
		Status:    judge.Pending,
	}
	dispatch, err := s.subRepo.CreatePending(ctx, &submission)
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}

	if err := s.mq.EnqueueSubmission(ctx, dispatch.SubmissionID, dispatch.Version); err != nil {
		return 0, errcode.ErrInternal.WithMsg("提交入队失败").Wrap(err)
	}
	if err := s.subRepo.MarkDispatchPublished(ctx, dispatch); err != nil {
		logger.Warnw("mark submission dispatch published failed",
			"submissionID", dispatch.SubmissionID, "version", dispatch.Version, "err", err)
	}
	return submission.PublicID, nil
}

// Rejudge 重判：重置目标提交(version+1/Pending/清结果) → 清相关缓存 → 重新入队。
// 榜单路①现算，重判后无需任何"重算触发"；这里删缓存只为让下次查榜/统计立即回落到 DB。
// 返回实际重判的提交条数。
func (s *SubmissionService) Rejudge(ctx context.Context, scope repository.RejudgeScope) (int, error) {
	targets, err := s.subRepo.GetRejudgeTargets(ctx, scope)
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	if len(targets) == 0 {
		return 0, errcode.ErrRejudgeTargetEmpty
	}
	ids := make([]int64, len(targets))
	for i, t := range targets {
		ids[i] = t.ID
	}
	dispatches, err := s.subRepo.ResetForRejudge(ctx, ids)
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	s.invalidateRejudgeCaches(ctx, targets)

	// 保留即时入队以维持当前延迟；失败的 Outbox 仍为 pending，后续由 Relay 重试。
	failed := 0
	for _, dispatch := range dispatches {
		if err := s.mq.EnqueueSubmission(ctx, dispatch.SubmissionID, dispatch.Version); err != nil {
			failed++
			continue
		}
		if err := s.subRepo.MarkDispatchPublished(ctx, dispatch); err != nil {
			logger.Warnw("mark rejudge dispatch published failed",
				"submissionID", dispatch.SubmissionID, "version", dispatch.Version, "err", err)
		}
	}
	if failed > 0 {
		return len(dispatches) - failed, errcode.ErrInternal.WithMsg("部分提交重新入队失败，请重试")
	}
	return len(dispatches), nil
}

// RejudgeSubmission 按公开提交编号单条重判。
func (s *SubmissionService) RejudgeSubmission(ctx context.Context, publicID int64) (int, error) {
	internalID, err := s.subRepo.ResolveIDByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrSubmissionNotFound
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return s.Rejudge(ctx, repository.RejudgeScope{SubmissionID: internalID})
}

// RejudgeContest 整场重判
func (s *SubmissionService) RejudgeContest(ctx context.Context, contestID int64) (int, error) {
	return s.Rejudge(ctx, repository.RejudgeScope{ContestID: contestID})
}

// RejudgeContestProblem 某比赛内某题的全部提交重判
func (s *SubmissionService) RejudgeContestProblem(ctx context.Context, contestID, problemID int64) (int, error) {
	return s.Rejudge(ctx, repository.RejudgeScope{ContestID: contestID, ProblemID: problemID})
}

// invalidateRejudgeCaches 删掉受影响的比赛/用户榜缓存，让读取回落到 DB（现算）。
func (s *SubmissionService) invalidateRejudgeCaches(ctx context.Context, targets []repository.RejudgeTarget) {
	contests := make(map[int64]struct{})
	hasNonContest := false
	for _, t := range targets {
		if t.ContestID == 0 {
			hasNonContest = true
			continue
		}
		contests[t.ContestID] = struct{}{}
		_ = s.cache.Delete(ctx,
			cache.ContestUserProblem(t.ContestID, t.UserID, t.ProblemID),
			cache.ContestProblemStats(t.ContestID, t.ProblemID),
			cache.ContestUserStats(t.ContestID, t.UserID),
			cache.ContestProblemSubmittedUsers(t.ContestID, t.ProblemID),
			cache.ContestProblemAcceptedUsers(t.ContestID, t.ProblemID),
			cache.ContestUserAcceptedProblems(t.ContestID, t.UserID),
		)
	}
	for cid := range contests {
		_ = s.cache.Delete(ctx, cache.ContestRank(cid))
	}
	if hasNonContest {
		_, _ = s.cache.Incr(ctx, cache.UserRankGen()) // 非比赛：换代失效用户榜缓存
	}
}

func submitRateLimit(ctx context.Context, c SubmissionCache, userID int64) error {
	ok, err := c.SetNX(ctx, cache.SubmitLock(userID), 1, submitInterval)
	if err != nil {
		return errcode.ErrRedis.Wrap(err)
	}
	if !ok {
		return errcode.ErrTooManyRequests.WithMsg("提交过于频繁，请稍后再试")
	}
	return nil
}

func (s *SubmissionService) List(ctx context.Context, page, pageSize int, status *int, username string, userID int64, problemDisplayID string, isAdmin bool) (*dto.SubmitListResp, error) {
	// 按题目筛选时，problemDisplayID 是对外题号，先解析成内部主键；解析不到说明没这题，直接返回空
	var problemPK int64
	if problemDisplayID != "" {
		pk, err := s.proRepo.ResolveID(ctx, problemDisplayID)
		if err != nil {
			return &dto.SubmitListResp{Total: 0, List: []dto.SubmissionItemResp{}}, nil
		}
		problemPK = pk
	}
	// 按用户筛选时 userID 是对外用户号，解析成内部主键；解析不到直接返回空
	var userPK int64
	if userID > 0 {
		pk, err := s.userRepo.ResolveIDByUID(ctx, userID)
		if err != nil {
			return &dto.SubmitListResp{Total: 0, List: []dto.SubmissionItemResp{}}, nil
		}
		userPK = pk
	}
	submitList, total, err := s.subRepo.ListWithInfo(ctx, page, pageSize, status, username, userPK, problemPK, isAdmin)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := dto.SubmitListResp{
		Total: total,
		List:  make([]dto.SubmissionItemResp, 0, len(submitList)),
	}
	for _, sub := range submitList {
		item := dto.SubmissionItemResp{
			ID:          sub.PublicID,
			ProblemID:   sub.ProblemDisplayID, // 对外题号，链接跳 /problem/{题号}
			ProblemName: sub.ProblemName,
			UserID:      sub.UserUID, // 对外用户号，链接跳 /user/{uid}
			UserName:    sub.UserName,
			Rating:      sub.Rating,
			Language:    sub.Language,
			Result:      sub.Status,
			TimeUsed:    sub.TimeUsed,
			MemoryUsed:  sub.MemoryUsed,
			CreatedAt:   sub.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		resp.List = append(resp.List, item)
	}
	return &resp, nil
}

func (s *SubmissionService) GetSubmissionByPublicID(ctx context.Context, publicID, requesterID int64, isAdmin bool) (*dto.SubmissionResp, error) {
	submission, err := s.subRepo.GetSubmissionByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrSubmissionNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	// 比赛与作业提交必须走各自的专用详情接口，由对应业务校验报名或团队权限。
	// 通用详情仅接受普通练习提交，提交者本人和站点管理员除外。
	if (submission.ContestID != 0 || submission.HomeworkID != 0) && submission.UserID != requesterID && !isAdmin {
		return nil, errcode.ErrSubmissionNotFound
	}
	caseList, err := s.subRepo.GetCaseResultBySubID(ctx, submission.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problem, err := s.proRepo.GetByID(ctx, submission.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	// 隐藏题的通用提交详情对非管理员不可见；比赛、作业管理视角走各自的专用接口。
	if submission.ContestID == 0 && problem.Hidden && !isAdmin {
		return nil, errcode.ErrProblemNotFound
	}
	author, err := s.userRepo.GetUserByID(ctx, submission.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	canViewCode := submission.UserID == requesterID || isAdmin
	respDto := dto.SubmissionResp{
		User: dto.UserDetail{
			ID:     author.UID, // 对外用户号
			Name:   author.Username,
			Avatar: avatarOr(author.Avatar),
		},
		Problem: dto.SubmissionProblem{
			ID:          problem.DisplayID, // 对外题号
			Name:        problem.Name,
			Description: problem.Description,
			OJ:          problem.OJ,
		},
		Submission: dto.SubmissionInfo{
			ID:            submission.PublicID,
			Language:      submission.Language,
			Code:          submission.Code,
			CanViewCode:   canViewCode,
			Status:        submission.Status,
			TimeUsed:      submission.TimeUsed,
			MemoryUsed:    submission.MemoryUsed,
			CompileOutput: submission.CompileOutput,
			CreatedAt:     submission.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		CaseResults: make([]dto.SubmissionCaseResult, 0),
	}
	sort.Slice(caseList, func(i, j int) bool {
		return caseList[i].ID < caseList[j].ID
	})
	for idx, caseResult := range caseList {
		respDto.CaseResults = append(respDto.CaseResults, dto.SubmissionCaseResult{
			ID:         idx + 1,
			Status:     caseResult.Status,
			TimeUsed:   caseResult.TimeUsed / nsPerMs,
			MemoryUsed: caseResult.MemoryUsed / bytesPerKB,
		})
	}
	// 越权控制：仅本人或管理员可见源代码，其他人一律隐藏
	if !canViewCode {
		respDto.Submission.Code = ""
		respDto.Submission.CompileOutput = ""
	}
	return &respDto, nil
}

// ResolveSubmissionID 把外部 8 位提交编号解析为内部主键。
// 仅供需要订阅内部评测通道的入口使用，不应写入 HTTP 响应。
func (s *SubmissionService) ResolveSubmissionID(ctx context.Context, publicID int64) (int64, error) {
	id, err := s.subRepo.ResolveIDByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrSubmissionNotFound
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return id, nil
}

func (s *SubmissionService) GetUserRecent7DaysPassCount(ctx context.Context, userID int64) (*dto.DailyAcceptedCountResp, error) {
	var dailyCounts dto.DailyAcceptedCountResp
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, now.Location())
	result, err := s.subRepo.GetUserRecentAcceptedSubmissions(ctx, userID, start)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemsByDay := make(map[string]map[int64]struct{})
	for _, item := range result {
		key := item.CreatedAt.In(now.Location()).Format("2006-01-02")
		if problemsByDay[key] == nil {
			problemsByDay[key] = make(map[int64]struct{})
		}
		problemsByDay[key][item.ProblemID] = struct{}{}
	}
	for i := 0; i < 7; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		dailyCounts.Date = append(dailyCounts.Date, date)
		dailyCounts.Count = append(dailyCounts.Count, len(problemsByDay[date]))
	}

	return &dailyCounts, nil
}
