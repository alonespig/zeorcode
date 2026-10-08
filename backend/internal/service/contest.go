package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"

	"gorm.io/gorm"
)

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

func (s *ContestService) CreateContest(ctx context.Context, req CreateContestParams) (*CreateContestResult, error) {
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
		Type:        model.ContestType(req.Type),
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

	return &CreateContestResult{
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

func (s *ContestService) resolveContestProblems(ctx context.Context, inputs []ContestProblemInput) ([]model.ContestProblem, error) {
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
	form UpdateContestParams) (*CreateContestResult, error) {
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
	if started && model.ContestType(form.Type) != contest.Type {
		return nil, errcode.ErrInvalidParams.WithMsg("比赛已开始，不能修改赛制")
	}

	contest.Name = form.Name
	contest.Description = form.Description
	contest.CoverURL = strings.TrimSpace(form.CoverURL)
	contest.Type = model.ContestType(form.Type)
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

	return &CreateContestResult{ID: contest.PublicID}, nil
}

func contestStatus(c model.Contest) model.ContestStatus {
	now := time.Now()
	if now.Before(c.StartTime) {
		return model.ContestNotStarted
	}
	if now.Before(contestEndTime(c)) {
		return model.ContestRunning
	}
	return model.ContestFinished
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

func (s *ContestService) ResolveAdminID(ctx context.Context, publicID int64) (int64, error) {
	id, err := s.repo.ResolveAnyIDByPublicID(ctx, publicID)
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

func (s *ContestService) ListContests(ctx context.Context, userID int64, page, pageSize int, keyword string, ctype int, status *int) (*ContestList, error) {
	archived := false
	return s.listContests(ctx, userID, page, pageSize, keyword, ctype, status, &archived)
}

func (s *ContestService) ListAdminContests(ctx context.Context, page, pageSize int, keyword string, ctype int, status *int, archived *bool) (*ContestList, error) {
	return s.listContests(ctx, 0, page, pageSize, keyword, ctype, status, archived)
}

func (s *ContestService) listContests(ctx context.Context, userID int64, page, pageSize int, keyword string, ctype int, status *int, archived *bool) (*ContestList, error) {
	contests, total, err := s.repo.ListContests(ctx, page, pageSize, keyword, ctype, status, archived)
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

	resp := &ContestList{
		Total: total,
		List:  make([]ContestItem, 0, len(contests)),
	}
	for _, c := range contests {
		count := userCountMap[c.ID]
		registered := registeredMap[c.ID]
		resp.List = append(resp.List, ContestItem{
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
			Archived:       c.Archived,
		})
	}
	return resp, nil
}

func (s *ContestService) SetContestArchived(ctx context.Context, contestID int64, archived bool) error {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return err
	}
	if archived && contestStatus(*contest) == model.ContestRunning {
		return errcode.ErrInvalidParams.WithMsg("进行中的比赛不能归档")
	}
	if contest.Archived == archived {
		return nil
	}
	if err := s.repo.SetContestArchived(ctx, contestID, archived); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
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

func (s *ContestService) GetContestDesc(ctx context.Context, id int64) (*ContestDesc, error) {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := &ContestDesc{
		Name:        contest.Name,
		Description: contest.Description,
		StartTime:   contest.StartTime,
		EndTime:     contestEndTime(*contest),
	}
	return resp, nil
}

// GetContestDetail userID=0 表示未登录场景
func (s *ContestService) GetContestDetail(ctx context.Context, id int64, userID int64) (*ContestDetail, error) {
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
	resp := &ContestDetail{
		Name:           contest.Name,
		StartTime:      contest.StartTime,
		EndTime:        contestEndTime(*contest),
		IsRegistered:   registered,
		NeedInviteCode: contest.InviteCode != "",
		Rule:           contest.Type.String(),
		Rated:          contest.Rated,
		Settled:        contest.Settled,
	}
	return resp, nil
}

func (s *ContestService) EditContest(ctx context.Context, contestID int64) (*ContestEditInfo, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	contestProblems, err := s.repo.GetContestProblemsByContestID(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &ContestEditInfo{
		Name:        contest.Name,
		Description: contest.Description,
		CoverURL:    contest.CoverURL,
		Type:        int(contest.Type),
		StartTime:   contest.StartTime,
		EndTime:     contestEndTime(*contest),
		ProblemList: make([]ContestEasyProblem, 0, len(contestProblems)),
		Rated:       contest.Rated,
	}

	sort.Slice(contestProblems, func(i, j int) bool {
		return contestProblems[i].Label < contestProblems[j].Label
	})
	problemMap := s.problemsByIDs(ctx, contestProblems)
	for _, p := range contestProblems {
		pr := problemMap[p.ProblemID]
		resp.ProblemList = append(resp.ProblemList, ContestEasyProblem{
			ID:    pr.DisplayID, // 对外题号（编辑表单据此回显 + 重新提交）
			Name:  pr.Name,
			Color: p.Color,
			Score: p.Score,
		})
	}

	return resp, nil
}

// problemNamesByIDs 一次性取出比赛各题的题名，避免循环里逐题查库（N+1）
