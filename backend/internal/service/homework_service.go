package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/common/logger"
	"zoj/internal/dto"
	"zoj/internal/infra/cache"
	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"

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
func (s *HomeworkService) ListByTeam(ctx context.Context, teamID, userID int64, isSiteAdmin bool, page, pageSize int) (*dto.HomeworkListResp, error) {
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
	resp := &dto.HomeworkListResp{Total: int(total), List: make([]dto.HomeworkItemResp, 0, len(list))}
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
		item := dto.HomeworkItemResp{
			ID:           hw.PublicID,
			Title:        hw.Title,
			Status:       hw.Status(now),
			ProblemCount: countByHW[hw.ID],
			StartTime:    hw.StartTime.Format(homeworkTimeLayout),
			EndTime:      hw.EndTime.Format(homeworkTimeLayout),
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
func (s *HomeworkService) Detail(ctx context.Context, id, userID int64, isSiteAdmin bool) (*dto.HomeworkDetailResp, error) {
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
	resp := &dto.HomeworkDetailResp{
		ID:            hw.PublicID,
		TeamID:        team.PublicID,
		TeamName:      team.Name,
		Title:         hw.Title,
		Description:   hw.Description,
		Status:        hw.Status(now),
		StartTime:     hw.StartTime.Format(homeworkTimeLayout),
		EndTime:       hw.EndTime.Format(homeworkTimeLayout),
		Creator:       creators[hw.CreatedBy].Username,
		CreatorAvatar: creators[hw.CreatedBy].Avatar,
		ProblemCount:  len(refs),
		TotalScore:    len(refs) * model.HomeworkProblemFullScore,
		CanEdit:       canEditHomework(access, hw) == nil,
		CanSeeAll:     access.CanManage(),
		Problems:      make([]dto.HomeworkProblemResp, 0, len(refs)),
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

	tagsByProblem := make(map[int64][]dto.TagItem)
	if rawTags, err := s.problemRepo.GetProblemTagsByIDs(ctx, problemIDs); err == nil {
		for _, tg := range rawTags {
			tagsByProblem[tg.ProblemID] = append(tagsByProblem[tg.ProblemID], dto.TagItem{ID: tg.ID, Name: tg.Name})
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

	// 我在时间窗内每题的最高分与整体通过状态
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
		statusMap, err = s.problemRepo.GetUserProblemStatusByIDs(ctx, userID, problemIDs)
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
		item := dto.HomeworkProblemResp{
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
func (s *HomeworkService) ProblemDetail(
	ctx context.Context,
	homeworkID int64,
	problemDisplayID string,
	userID int64,
	isSiteAdmin bool,
) (*dto.ProblemDetailResp, error) {
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
	tags := make([]dto.TagItem, 0, len(rawTags))
	for _, tag := range rawTags {
		tags = append(tags, dto.TagItem{ID: tag.ID, Name: tag.Name})
	}
	rawSamples, err := s.problemRepo.GetProblemSamplesByID(ctx, problem.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	samples := make([]dto.ProblemSample, 0, len(rawSamples))
	for _, sample := range rawSamples {
		samples = append(samples, dto.ProblemSample{
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

	return &dto.ProblemDetailResp{
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
func (s *HomeworkService) Save(ctx context.Context, teamID, id int64, req *dto.SaveHomeworkReq, userID int64, isSiteAdmin bool) (int64, error) {
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
func (s *HomeworkService) CreateFromAgent(ctx context.Context, teamID int64, req *dto.SaveHomeworkReq, userID int64, isSiteAdmin bool, actionID int64) (int64, error) {
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
func (s *HomeworkService) Submit(ctx context.Context, id int64, req *dto.HomeworkSubmitReq, userID int64, isSiteAdmin bool) (int64, error) {
	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return 0, err
	}
	if !access.CanView() {
		return 0, errcode.ErrPermissionDenied.WithMsg("仅团队成员可提交")
	}
	if hw.Status(time.Now()) == model.HomeworkNotStarted {
		return 0, errcode.ErrHomeworkNotStarted
	}
	if err := submitRateLimit(ctx, s.cache, userID); err != nil {
		return 0, err
	}

	// 题目必须在本作业内，否则作业提交可以被用来给任意题刷分
	problemID, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(req.ProblemID))
	if err != nil {
		return 0, errcode.ErrProblemNotFound
	}
	refs, err := s.repo.ProblemRefsByHomeworkIDs(ctx, []int64{id})
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	inHomework := false
	for _, ref := range refs {
		if ref.ProblemID == problemID {
			inHomework = true
			break
		}
	}
	if !inHomework {
		return 0, errcode.ErrProblemNotFound.WithMsg("该题不属于本作业")
	}

	problem, err := s.problemRepo.GetByID(ctx, problemID)
	if err != nil {
		return 0, errcode.ErrProblemNotFound
	}
	if err := requireProblemTestData(ctx, s.testData, problem); err != nil {
		return 0, err
	}
	language, err := s.languages.ResolveEnabledName(ctx, req.Language)
	if err != nil {
		return 0, err
	}
	publicID, err := newUniquePublicID(ctx, s.submitRepo.PublicIDExists)
	if err != nil {
		return 0, err
	}

	submission := &model.Submission{
		PublicID:   publicID,
		UserID:     userID,
		ProblemID:  problemID,
		Code:       req.Code,
		Status:     judge.Pending,
		Language:   language,
		HomeworkID: id,
		CreatedAt:  time.Now(),
	}
	dispatch, err := s.submitRepo.CreatePending(ctx, submission)
	if err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	if err := s.mq.EnqueueSubmission(ctx, dispatch.SubmissionID, dispatch.Version); err != nil {
		return 0, errcode.ErrInternal.WithMsg("提交入队失败").Wrap(err)
	}
	if err := s.submitRepo.MarkDispatchPublished(ctx, dispatch); err != nil {
		logger.Warnw("mark homework submission dispatch published failed",
			"submissionID", dispatch.SubmissionID, "version", dispatch.Version, "err", err)
	}
	return submission.PublicID, nil
}

// Rank 排行榜（IOI）：每题取时间窗内最高分，总分为各题之和；
// 按总分降序，同分时达成时间早者靠前。
func (s *HomeworkService) Rank(ctx context.Context, id, userID int64, isSiteAdmin bool) (*dto.HomeworkRankResp, error) {
	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	if !access.CanView() {
		return nil, errcode.ErrPermissionDenied.WithMsg("仅团队成员可查看")
	}

	refs, err := s.repo.ProblemRefsByHomeworkIDs(ctx, []int64{id})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.HomeworkRankResp{
		StartTime:  hw.StartTime.Format(homeworkTimeLayout),
		EndTime:    hw.EndTime.Format(homeworkTimeLayout),
		ProblemIDs: make([]string, 0, len(refs)),
		TotalScore: len(refs) * model.HomeworkProblemFullScore,
		List:       []dto.HomeworkRankRow{},
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
	displayByID := make(map[int64]string, len(problems))
	for _, p := range problems {
		displayByID[p.ID] = p.DisplayID
	}
	// 列头顺序与作业题目顺序一致
	for _, ref := range refs {
		if d, ok := displayByID[ref.ProblemID]; ok {
			resp.ProblemIDs = append(resp.ProblemIDs, d)
		}
	}

	best, err := s.repo.BestScores(ctx, id, hw.StartTime, hw.EndTime)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	// 排行榜以当前团队成员为基准：没有提交的人也显示 0 分；
	// 退队的人即使有历史成绩，也不再占榜位。
	members, err := s.teamRepo.ListMembers(ctx, hw.TeamID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}
	if len(userIDs) == 0 {
		return resp, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp.List = buildHomeworkRankRows(refs, displayByID, members, users, best)
	return resp, nil
}

type homeworkRankAggregate struct {
	total         int
	solved        int
	scores        map[int64]int
	achievedAt    time.Time // 各题达成最高分的最晚时刻，用于同分排序
	hasSubmission bool
}

type homeworkRankEntry struct {
	row           dto.HomeworkRankRow
	achievedAt    time.Time
	hasSubmission bool
}

func buildHomeworkRankRows(
	refs []repository.HomeworkProblemRef,
	displayByID map[int64]string,
	members []model.TeamMember,
	users []model.User,
	best []repository.HomeworkBestScore,
) []dto.HomeworkRankRow {
	byUser := make(map[int64]*homeworkRankAggregate, len(members))
	for _, member := range members {
		byUser[member.UserID] = &homeworkRankAggregate{scores: make(map[int64]int, len(refs))}
	}
	for _, score := range best {
		a, isMember := byUser[score.UserID]
		if !isMember {
			continue
		}
		a.hasSubmission = true
		a.scores[score.ProblemID] = score.Best
		a.total += score.Best
		if score.Best >= model.HomeworkProblemFullScore {
			a.solved++
		}
		if score.AchievedAt.After(a.achievedAt) {
			a.achievedAt = score.AchievedAt
		}
	}

	userByID := make(map[int64]model.User, len(users))
	for _, user := range users {
		userByID[user.ID] = user
	}
	entries := make([]homeworkRankEntry, 0, len(members))
	for _, member := range members {
		user, exists := userByID[member.UserID]
		if !exists {
			continue
		}
		a := byUser[member.UserID]
		row := dto.HomeworkRankRow{
			UID:         user.UID,
			Username:    user.Username,
			StudentNo:   studentNoValue(user.StudentNo),
			RealName:    user.RealName,
			Avatar:      user.Avatar,
			TotalScore:  a.total,
			SolvedCount: a.solved,
			Cells:       make([]dto.RankCell, 0, len(refs)),
		}
		for _, ref := range refs {
			display, exists := displayByID[ref.ProblemID]
			if !exists {
				continue
			}
			score := a.scores[ref.ProblemID]
			row.Cells = append(row.Cells, dto.RankCell{
				ProblemID: display,
				Score:     score,
				Solved:    score >= model.HomeworkProblemFullScore,
			})
		}
		entries = append(entries, homeworkRankEntry{
			row:           row,
			achievedAt:    a.achievedAt,
			hasSubmission: a.hasSubmission,
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].row.TotalScore != entries[j].row.TotalScore {
			return entries[i].row.TotalScore > entries[j].row.TotalScore
		}
		// 同分时有提交者在完全未提交者之前；两人都有提交时再比较达成时间。
		if entries[i].hasSubmission != entries[j].hasSubmission {
			return entries[i].hasSubmission
		}
		if entries[i].hasSubmission && !entries[i].achievedAt.Equal(entries[j].achievedAt) {
			return entries[i].achievedAt.Before(entries[j].achievedAt)
		}
		if entries[i].row.StudentNo != entries[j].row.StudentNo {
			return entries[i].row.StudentNo < entries[j].row.StudentNo
		}
		return entries[i].row.UID < entries[j].row.UID
	})

	rows := make([]dto.HomeworkRankRow, 0, len(entries))
	for i := range entries {
		entries[i].row.Rank = i + 1
		rows = append(rows, entries[i].row)
	}
	return rows
}

// Submissions 作业提交列表。
// 普通成员只看自己（忽略 uid 筛选）；团队管理员及以上可看全部并按成员筛选。
func (s *HomeworkService) Submissions(ctx context.Context, id, userID int64, isSiteAdmin bool, q *dto.HomeworkSubmissionQuery) (*dto.HomeworkSubmissionListResp, error) {
	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	if !access.CanView() {
		return nil, errcode.ErrPermissionDenied.WithMsg("仅团队成员可查看")
	}

	f := &repository.HomeworkSubmissionFilter{
		HomeworkID: id,
		Status:     q.Status,
		Page:       q.Page,
		PageSize:   q.PageSize,
	}
	canSeeAll := access.CanManage()
	if canSeeAll {
		// 只有能看全部的人，uid 筛选才生效
		if q.UID != 0 {
			targetID, err := s.userRepo.ResolveIDByUID(ctx, q.UID)
			if err != nil {
				return nil, errcode.ErrUserNotFound
			}
			f.UserID = targetID
		}
	} else {
		f.UserID = userID // 普通成员强制只看自己
	}
	if strings.TrimSpace(q.ProblemID) != "" {
		pid, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(q.ProblemID))
		if err != nil {
			return nil, errcode.ErrProblemNotFound
		}
		f.ProblemID = pid
	}

	list, total, err := s.repo.ListSubmissions(ctx, f)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &dto.HomeworkSubmissionListResp{
		Total:     int(total),
		CanSeeAll: canSeeAll,
		List:      make([]dto.HomeworkSubmissionItem, 0, len(list)),
	}
	if len(list) == 0 {
		return resp, nil
	}

	userIDs := make([]int64, 0, len(list))
	problemIDs := make([]int64, 0, len(list))
	for _, sub := range list {
		userIDs = append(userIDs, sub.UserID)
		problemIDs = append(problemIDs, sub.ProblemID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userByID := make(map[int64]model.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}
	problems, err := s.problemRepo.FindByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemByID := make(map[int64]model.Problem, len(problems))
	for _, p := range problems {
		problemByID[p.ID] = p
	}
	for _, sub := range list {
		u := userByID[sub.UserID]
		p := problemByID[sub.ProblemID]
		resp.List = append(resp.List, dto.HomeworkSubmissionItem{
			ID:          sub.PublicID,
			UID:         u.UID,
			Username:    u.Username,
			StudentNo:   studentNoValue(u.StudentNo),
			RealName:    u.RealName,
			Avatar:      u.Avatar,
			ProblemID:   p.DisplayID,
			ProblemName: p.Name,
			Status:      sub.Status,
			Score:       sub.Score,
			Language:    sub.Language,
			TimeUsed:    sub.TimeUsed,
			MemoryUsed:  sub.MemoryUsed,
			InWindow:    hw.InWindow(sub.CreatedAt),
			CreatedAt:   sub.CreatedAt.Format(homeworkTimeLayout),
		})
	}
	return resp, nil
}

// SubmissionDetail 返回作业提交的摘要与源代码。
// 普通成员只能查看自己的提交；团队所有者、团队管理员和站点管理员可以查看全部。
func (s *HomeworkService) SubmissionDetail(
	ctx context.Context,
	homeworkID, submissionPublicID, userID int64,
	isSiteAdmin bool,
) (*dto.HomeworkSubmissionDetailResp, error) {
	hw, access, err := s.loadWithAccess(ctx, homeworkID, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	submission, err := s.submitRepo.GetSubmissionByPublicID(ctx, submissionPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrSubmissionNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if submission.HomeworkID != homeworkID {
		return nil, errcode.ErrSubmissionNotFound
	}
	if submission.UserID != userID && !access.CanManage() {
		return nil, errcode.ErrSubmissionNotFound
	}

	user, err := s.userRepo.GetUserByID(ctx, submission.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problem, err := s.problemRepo.GetByID(ctx, submission.ProblemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	caseList, err := s.submitRepo.GetCaseResultBySubID(ctx, submission.ID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	sort.Slice(caseList, func(i, j int) bool {
		return caseList[i].ID < caseList[j].ID
	})
	caseResults := make([]dto.SubmissionCaseResult, 0, len(caseList))
	for idx, caseResult := range caseList {
		caseResults = append(caseResults, dto.SubmissionCaseResult{
			ID:         idx + 1,
			Status:     caseResult.Status,
			TimeUsed:   caseResult.TimeUsed / nsPerMs,
			MemoryUsed: caseResult.MemoryUsed / bytesPerKB,
		})
	}

	return &dto.HomeworkSubmissionDetailResp{
		ID:          submission.PublicID,
		UID:         user.UID,
		Username:    user.Username,
		StudentNo:   studentNoValue(user.StudentNo),
		RealName:    user.RealName,
		Avatar:      user.Avatar,
		ProblemID:   problem.DisplayID,
		ProblemName: problem.Name,
		OJ:          problem.OJ,
		Status:      submission.Status,
		Score:       submission.Score,
		Language:    submission.Language,
		Code:        submission.Code,
		TimeUsed:    submission.TimeUsed,
		MemoryUsed:  submission.MemoryUsed,
		CaseResults: caseResults,
		InWindow:    hw.InWindow(submission.CreatedAt),
		CreatedAt:   submission.CreatedAt.Format(homeworkTimeLayout),
	}, nil
}

// ===== 内部工具 =====

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

func (s *HomeworkService) teamAccess(ctx context.Context, teamID, userID int64, isSiteAdmin bool) (TeamAccess, error) {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TeamAccess{}, errcode.ErrTeamNotFound
		}
		return TeamAccess{}, errcode.ErrDatabase.Wrap(err)
	}
	return s.teamSrv.Access(ctx, teamID, userID, isSiteAdmin)
}

// loadWithAccess 取作业并算出调用方在其所属团队中的权限。
// 非团队成员一律返回「作业不存在」，不泄露作业的存在。
func (s *HomeworkService) loadWithAccess(ctx context.Context, id, userID int64, isSiteAdmin bool) (*model.Homework, TeamAccess, error) {
	hw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, TeamAccess{}, errcode.ErrHomeworkNotFound
		}
		return nil, TeamAccess{}, errcode.ErrDatabase.Wrap(err)
	}
	access, err := s.teamSrv.Access(ctx, hw.TeamID, userID, isSiteAdmin)
	if err != nil {
		return nil, TeamAccess{}, err
	}
	if !access.CanView() {
		return nil, TeamAccess{}, errcode.ErrHomeworkNotFound
	}
	return hw, access, nil
}

// resolveProblems 把对外题号列表转成关联行，数组下标即 Sort。
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
func parseHomeworkWindow(req *dto.SaveHomeworkReq) (time.Time, time.Time, error) {
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
