package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"

	"gorm.io/gorm"
)

type HomeworkService struct {
	repo        *repository.HomeworkRepo
	teamRepo    *repository.TeamRepo
	problemRepo *repository.ProblemRepo
	submitRepo  *repository.SubmissionRepo
	userRepo    *repository.UserRepo
	teamSrv     *TeamService
	cache       *cache.Cache
	mq          *mq.MQ
	testData    ProblemTestDataReader
	languages   LanguageResolver
}

func NewHomeworkService(
	repo *repository.HomeworkRepo,
	teamRepo *repository.TeamRepo,
	problemRepo *repository.ProblemRepo,
	submitRepo *repository.SubmissionRepo,
	userRepo *repository.UserRepo,
	teamSrv *TeamService,
	c *cache.Cache,
	q *mq.MQ,
	testData ProblemTestDataReader,
	languages LanguageResolver,
) *HomeworkService {
	return &HomeworkService{
		repo: repo, teamRepo: teamRepo, problemRepo: problemRepo, submitRepo: submitRepo,
		userRepo: userRepo, teamSrv: teamSrv, cache: c, mq: q, testData: testData, languages: languages,
	}
}

const homeworkTimeLayout = "2006-01-02 15:04:05"

// ListByTeam 团队作业列表，仅团队成员可见。
func (s *HomeworkService) ListByTeam(ctx context.Context, teamID, userID int64, isSiteAdmin bool, page, pageSize int) (*HomeworkList, error) {
	access, err := s.teamAccess(ctx, teamID, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	if !access.CanView() {
		return nil, errcode.ErrPermissionDenied.WithMsg("仅团队成员可查看")
	}

	list, total, err := s.repo.ListByTeam(ctx, teamID, page, pageSize)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &HomeworkList{Total: total, List: make([]HomeworkItem, 0, len(list))}
	if len(list) == 0 {
		return resp, nil
	}

	ids := make([]int64, 0, len(list))
	for _, hw := range list {
		ids = append(ids, hw.ID)
	}
	refs, err := s.repo.ProblemRefsByHomeworkIDs(ctx, ids)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	countByHW := make(map[int64]int, len(list))
	for _, ref := range refs {
		countByHW[ref.HomeworkID]++
	}
	solved, err := s.repo.SolvedCountsByUser(ctx, ids, userID, model.HomeworkProblemFullScore)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	creators, err := s.creatorUsers(ctx, list)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, hw := range list {
		item := HomeworkItem{
			ID:           hw.PublicID,
			Title:        hw.Title,
			Status:       hw.Status(now),
			ProblemCount: countByHW[hw.ID],
			StartTime:    hw.StartTime,
			EndTime:      hw.EndTime,
			Creator:      creators[hw.CreatedBy].Username,
			CanEdit:      canEditHomework(access, &hw) == nil,
			CanDelete:    canDeleteHomework(access, &hw) == nil,
		}
		if userID != 0 {
			n := solved[hw.ID]
			item.SolvedCount = &n
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}

// Detail 作业详情。未开始时普通成员看不到题目列表（防提前刷题），
// 布置者和团队管理员不受限。
func (s *HomeworkService) Detail(ctx context.Context, id, userID int64, isSiteAdmin bool) (*HomeworkDetail, error) {
	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	team, err := s.teamRepo.GetByID(ctx, hw.TeamID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	creators, err := s.creatorUsers(ctx, []model.Homework{*hw})
	if err != nil {
		return nil, err
	}

	refs, err := s.repo.ProblemRefsByHomeworkIDs(ctx, []int64{id})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	now := time.Now()
	resp := &HomeworkDetail{
		ID:            hw.PublicID,
		TeamID:        team.PublicID,
		TeamName:      team.Name,
		Title:         hw.Title,
		Description:   hw.Description,
		Status:        hw.Status(now),
		StartTime:     hw.StartTime,
		EndTime:       hw.EndTime,
		Creator:       creators[hw.CreatedBy].Username,
		CreatorAvatar: creators[hw.CreatedBy].Avatar,
		ProblemCount:  len(refs),
		TotalScore:    len(refs) * model.HomeworkProblemFullScore,
		CanEdit:       canEditHomework(access, hw) == nil,
		CanSeeAll:     access.CanManage(),
		Problems:      make([]HomeworkProblem, 0, len(refs)),
	}

	// 未开始：只有能编辑的人（布置者/超管）和团队管理员可以提前看题
	if hw.Status(now) == model.HomeworkNotStarted && !resp.CanEdit && !access.CanManage() {
		resp.Locked = true
		return resp, nil
	}
	if len(refs) == 0 {
		return resp, nil
	}

	problemIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		problemIDs = append(problemIDs, ref.ProblemID)
	}
	problems, err := s.problemRepo.FindByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemByID := make(map[int64]model.Problem, len(problems))
	for _, p := range problems {
		problemByID[p.ID] = p
	}

	tagsByProblem := make(map[int64][]TagItem)
	if rawTags, err := s.problemRepo.GetProblemTagsByIDs(ctx, problemIDs); err == nil {
		for _, tg := range rawTags {
			tagsByProblem[tg.ProblemID] = append(tagsByProblem[tg.ProblemID], TagItem{ID: tg.ID, Name: tg.Name})
		}
	}
	stats, err := s.submitRepo.GetHomeworkProblemStats(ctx, id, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	statsByProblem := make(map[int64]repository.ProblemSubmissionStat, len(stats))
	for _, st := range stats {
		statsByProblem[st.ProblemID] = st
	}

	// 时间窗内的最高分，以及仅限本作业的提交状态（包含补交）。
	myBest := map[int64]int{}
	statusMap := map[int64]int{}
	if userID != 0 {
		best, err := s.repo.BestScoresByUser(ctx, id, userID, hw.StartTime, hw.EndTime)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
		for _, b := range best {
			myBest[b.ProblemID] = b.Best
		}
		statusMap, err = s.submitRepo.GetHomeworkUserProblemStatuses(ctx, id, userID, problemIDs)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
	}

	solvedCount, myScore := 0, 0
	// 按 refs 顺序输出（repo 已按 sort 升序）
	for _, ref := range refs {
		p, ok := problemByID[ref.ProblemID]
		if !ok {
			continue // 题目被删了，跳过而不是整单打不开
		}
		st := statsByProblem[p.ID]
		item := HomeworkProblem{
			ID:            p.DisplayID,
			Name:          p.Name,
			Difficulty:    p.Difficulty,
			Tags:          tagsByProblem[p.ID],
			AcceptedCount: int(st.AcceptedCount),
			SubmitCount:   int(st.SubmitCount),
		}
		if status, ok := statusMap[p.ID]; ok {
			item.Status = &status
		}
		if userID != 0 {
			score := myBest[p.ID]
			item.MyScore = &score
			myScore += score
			if score >= model.HomeworkProblemFullScore {
				solvedCount++
			}
		}
		resp.Problems = append(resp.Problems, item)
	}
	if userID != 0 {
		resp.SolvedCount = &solvedCount
		resp.MyScore = &myScore
	}
	return resp, nil
}

// ProblemDetail 返回作业上下文中的单题题面。
// 它与 Detail 使用同一套成员权限和未开始保护，避免通过单题 URL 绕过作业锁定。

func (s *HomeworkService) Save(ctx context.Context, teamID, id int64, req SaveHomeworkParams, userID int64, isSiteAdmin bool) (int64, error) {
	start, end, err := parseHomeworkWindow(req)
	if err != nil {
		return 0, err
	}
	problems, err := s.resolveProblems(ctx, req.Problems)
	if err != nil {
		return 0, err
	}

	if id == 0 {
		access, err := s.teamAccess(ctx, teamID, userID, isSiteAdmin)
		if err != nil {
			return 0, err
		}
		if !access.CanManage() {
			return 0, errcode.ErrPermissionDenied.WithMsg("只有团队所有者和管理员能布置作业")
		}
		publicID, err := newUniquePublicID(ctx, s.repo.PublicIDExists)
		if err != nil {
			return 0, err
		}
		hw := &model.Homework{
			PublicID:    publicID,
			TeamID:      teamID,
			Title:       strings.TrimSpace(req.Title),
			Description: req.Description,
			StartTime:   start,
			EndTime:     end,
			CreatedBy:   userID,
		}
		if err := s.repo.CreateWithProblems(ctx, hw, problems); err != nil {
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		return hw.PublicID, nil
	}

	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return 0, err
	}
	// 方案 B：只有布置者本人能改，团队管理员也不行
	if err := canEditHomework(access, hw); err != nil {
		return 0, err
	}
	updated := &model.Homework{
		ID:          id,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		StartTime:   start,
		EndTime:     end,
	}
	if err := s.repo.UpdateWithProblems(ctx, updated, problems); err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return hw.PublicID, nil
}

// CreateFromAgent creates one homework for one confirmed Agent action. The
// source action unique key makes retries safe after an uncertain network result.
func (s *HomeworkService) CreateFromAgent(ctx context.Context, teamID int64, req SaveHomeworkParams, userID int64, isSiteAdmin bool, actionID int64) (int64, error) {
	if actionID <= 0 {
		return 0, errcode.ErrInvalidParams.WithMsg("非法 Agent 操作")
	}
	if existing, err := s.repo.GetBySourceActionID(ctx, actionID); err == nil {
		return existing.PublicID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrDatabase.Wrap(err)
	}

	start, end, err := parseHomeworkWindow(req)
	if err != nil {
		return 0, err
	}
	problems, err := s.resolveProblems(ctx, req.Problems)
	if err != nil {
		return 0, err
	}
	access, err := s.teamAccess(ctx, teamID, userID, isSiteAdmin)
	if err != nil {
		return 0, err
	}
	if !access.CanManage() {
		return 0, errcode.ErrPermissionDenied.WithMsg("当前账号不能在该团队布置作业")
	}
	publicID, err := newUniquePublicID(ctx, s.repo.PublicIDExists)
	if err != nil {
		return 0, err
	}
	homework := &model.Homework{
		PublicID: publicID, TeamID: teamID, Title: strings.TrimSpace(req.Title),
		Description: req.Description, StartTime: start, EndTime: end,
		CreatedBy: userID, SourceActionID: &actionID,
	}
	if err := s.repo.CreateWithProblems(ctx, homework, problems); err != nil {
		// The unique source_action_id may have been committed by a concurrent retry.
		if existing, lookupErr := s.repo.GetBySourceActionID(ctx, actionID); lookupErr == nil {
			return existing.PublicID, nil
		}
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return homework.PublicID, nil
}

// Delete 删除作业：布置者本人，或团队所有者兜底（防布置人离队后没人能清理）。
func (s *HomeworkService) Delete(ctx context.Context, id, userID int64, isSiteAdmin bool) error {
	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return err
	}
	if err := canDeleteHomework(access, hw); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// Submit 作业内提交。截止后仍可提交（不计入排行榜），但未开始不能提交。

func (s *HomeworkService) ResolveID(ctx context.Context, publicID int64) (int64, error) {
	id, err := s.repo.ResolveIDByPublicID(ctx, publicID)
	if err == nil {
		return id, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrHomeworkNotFound
	}
	return 0, errcode.ErrDatabase.Wrap(err)
}

func (s *HomeworkService) ResolveTeamID(ctx context.Context, publicID int64) (int64, error) {
	return s.teamSrv.ResolveID(ctx, publicID)
}

func (s *HomeworkService) creatorUsers(ctx context.Context, list []model.Homework) (map[int64]model.User, error) {
	ids := make([]int64, 0, len(list))
	seen := make(map[int64]struct{}, len(list))
	for _, hw := range list {
		if hw.CreatedBy == 0 {
			continue
		}
		if _, ok := seen[hw.CreatedBy]; ok {
			continue
		}
		seen[hw.CreatedBy] = struct{}{}
		ids = append(ids, hw.CreatedBy)
	}
	if len(ids) == 0 {
		return map[int64]model.User{}, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	usersByID := make(map[int64]model.User, len(users))
	for _, u := range users {
		usersByID[u.ID] = u
	}
	return usersByID, nil
}

// parseHomeworkWindow 解析并校验起止时间。
func parseHomeworkWindow(req SaveHomeworkParams) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation(homeworkTimeLayout, strings.TrimSpace(req.StartTime), time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, errcode.ErrInvalidParams.WithMsg("开始时间格式错误")
	}
	end, err := time.ParseInLocation(homeworkTimeLayout, strings.TrimSpace(req.EndTime), time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, errcode.ErrInvalidParams.WithMsg("结束时间格式错误")
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, errcode.ErrInvalidParams.WithMsg("结束时间必须晚于开始时间")
	}
	return start, end, nil
}
