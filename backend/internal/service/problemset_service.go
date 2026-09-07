package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"

	"gorm.io/gorm"
)

type ProblemSetService struct {
	repo        *repository.ProblemSetRepo
	problemRepo *repository.ProblemRepo
	subRepo     *repository.SubmissionRepo
	userRepo    *repository.UserRepo
}

func NewProblemSetService(
	repo *repository.ProblemSetRepo,
	problemRepo *repository.ProblemRepo,
	subRepo *repository.SubmissionRepo,
	userRepo *repository.UserRepo,
) *ProblemSetService {
	return &ProblemSetService{repo: repo, problemRepo: problemRepo, subRepo: subRepo, userRepo: userRepo}
}

const problemSetTimeLayout = "2006-01-02 15:04:05"

// List 题单分页列表。includeDraft 仅后台传 true。
// userID 为 nil（未登录）时不查做题状态，SolvedCount 留空，前端不渲染进度。
func (s *ProblemSetService) List(ctx context.Context, form *dto.ProblemSetListReq, userID *int64, includeDraft bool) (*dto.ProblemSetListResp, error) {
	sets, total, err := s.repo.List(ctx, &repository.ProblemSetQuery{
		Page:         form.Page,
		PageSize:     form.PageSize,
		Keyword:      form.Keyword,
		TagIDs:       parseIDList(form.Tags),
		Visibility:   form.Visibility,
		IncludeDraft: includeDraft,
	})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	resp := &dto.ProblemSetListResp{Total: int(total), List: make([]dto.ProblemSetItemResp, 0, len(sets))}
	if len(sets) == 0 {
		return resp, nil
	}

	setIDs := make([]int64, 0, len(sets))
	for _, set := range sets {
		setIDs = append(setIDs, set.ID)
	}

	// 一次取回本页所有题单的题目引用，据此算题数和进度，避免逐个题单查询
	refs, err := s.repo.ProblemRefsBySetIDs(ctx, setIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemIDsBySet := make(map[int64][]int64, len(sets))
	allProblemIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		problemIDsBySet[ref.ProblemSetID] = append(problemIDsBySet[ref.ProblemSetID], ref.ProblemID)
		allProblemIDs = append(allProblemIDs, ref.ProblemID)
	}

	// 同理，当前用户在本页所有题目上的通过状态也只查一次
	statusMap := map[int64]int{}
	if userID != nil && len(allProblemIDs) > 0 {
		statusMap, err = s.problemRepo.GetUserProblemStatusByIDs(ctx, *userID, allProblemIDs)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
	}

	tagsBySet, err := s.tagsBySetIDs(ctx, setIDs)
	if err != nil {
		return nil, err
	}
	authors, err := s.authorNames(ctx, sets)
	if err != nil {
		return nil, err
	}

	for _, set := range sets {
		problemIDs := problemIDsBySet[set.ID]
		item := dto.ProblemSetItemResp{
			ID:           set.PublicID,
			Title:        set.Title,
			Tags:         tagsBySet[set.ID],
			ProblemCount: len(problemIDs),
			Visibility:   set.Visibility,
			Published:    set.Published,
			Author:       authors[set.CreatedBy],
			UpdatedAt:    set.UpdatedAt.Format(problemSetTimeLayout),
		}
		if userID != nil {
			solved := countSolved(problemIDs, statusMap)
			item.SolvedCount = &solved
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}

// Detail 题单详情。
// 邀请码题单未解锁时 Locked=true 且不返回题目列表，但标题/描述/标签/题数照常返回，
// 这样别人能知道题单存在、也能看到它讲什么。管理员免解锁。
func (s *ProblemSetService) Detail(ctx context.Context, id int64, userID *int64, isAdmin bool) (*dto.ProblemSetDetailResp, error) {
	set, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemSetNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	// 草稿对外不存在，返回「不存在」而不是「无权限」，避免泄露草稿的存在
	if set.Published != model.ProblemSetPublished && !isAdmin {
		return nil, errcode.ErrProblemSetNotFound
	}

	refs, err := s.repo.ProblemRefsBySetIDs(ctx, []int64{id})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		problemIDs = append(problemIDs, ref.ProblemID)
	}

	tagsBySet, err := s.tagsBySetIDs(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	authors, err := s.authorNames(ctx, []model.ProblemSet{*set})
	if err != nil {
		return nil, err
	}

	resp := &dto.ProblemSetDetailResp{
		ID:           set.PublicID,
		Title:        set.Title,
		Description:  set.Description,
		Tags:         tagsBySet[set.ID],
		Visibility:   set.Visibility,
		Published:    set.Published,
		ProblemCount: len(problemIDs),
		Author:       authors[set.CreatedBy],
		UpdatedAt:    set.UpdatedAt.Format(problemSetTimeLayout),
		Problems:     make([]dto.ProblemSetProblemResp, 0, len(problemIDs)),
	}

	unlocked, err := s.hasAccess(ctx, set, userID, isAdmin)
	if err != nil {
		return nil, err
	}
	if !unlocked {
		resp.Locked = true
		return resp, nil
	}
	if len(problemIDs) == 0 {
		return resp, nil
	}

	statusMap := map[int64]int{}
	if userID != nil {
		statusMap, err = s.problemRepo.GetUserProblemStatusByIDs(ctx, *userID, problemIDs)
		if err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
	}
	if userID != nil {
		solved := countSolved(problemIDs, statusMap)
		resp.SolvedCount = &solved
	}

	problems, err := s.problemRepo.FindByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemByID := make(map[int64]model.Problem, len(problems))
	for _, problem := range problems {
		problemByID[problem.ID] = problem
	}

	tagsByProblem := make(map[int64][]dto.TagItem)
	rawTags, err := s.problemRepo.GetProblemTagsByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	for _, tag := range rawTags {
		tagsByProblem[tag.ProblemID] = append(tagsByProblem[tag.ProblemID], dto.TagItem{ID: tag.ID, Name: tag.Name})
	}

	statsByProblem := make(map[int64]repository.ProblemSubmissionStat, len(problemIDs))
	stats, err := s.subRepo.GetProblemSubmissionStats(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	for _, stat := range stats {
		statsByProblem[stat.ProblemID] = stat
	}

	// 按 refs 的顺序输出（repo 已按 sort 升序），不要按 FindByIDs 的返回顺序
	for _, ref := range refs {
		problem, ok := problemByID[ref.ProblemID]
		if !ok {
			continue // 题目被删了，跳过而不是报错，避免整单打不开
		}
		stat := statsByProblem[problem.ID]
		item := dto.ProblemSetProblemResp{
			ID:            problem.DisplayID,
			Name:          problem.Name,
			Difficulty:    problem.Difficulty,
			Tags:          tagsByProblem[problem.ID],
			AcceptedCount: int(stat.AcceptedCount),
			SubmitCount:   int(stat.SubmitCount),
		}
		if status, ok := statusMap[problem.ID]; ok {
			item.Status = &status
		}
		resp.Problems = append(resp.Problems, item)
	}
	return resp, nil
}

// Unlock 校验邀请码并记录解锁。
func (s *ProblemSetService) Unlock(ctx context.Context, id, userID int64, code string) error {
	set, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrProblemSetNotFound
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	if set.Published != model.ProblemSetPublished {
		return errcode.ErrProblemSetNotFound
	}
	if set.Visibility != model.ProblemSetInviteOnly {
		return nil // 公开题单本就不需要解锁，按成功处理
	}
	if set.InviteCode == "" || set.InviteCode != strings.TrimSpace(code) {
		return errcode.ErrContestInviteCodeInvalid
	}
	if err := s.repo.CreateUnlock(ctx, userID, id); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// Save 新建（id==0）或更新题单。题目顺序取 req.Problems 的数组下标。
func (s *ProblemSetService) Save(ctx context.Context, id int64, req *dto.SaveProblemSetReq, actorID int64) (int64, error) {
	inviteCode := strings.TrimSpace(req.InviteCode)
	if id == 0 && req.Visibility == model.ProblemSetInviteOnly && inviteCode == "" {
		return 0, errcode.ErrInvalidParams.WithMsg("选择邀请码可见时必须填写邀请码")
	}
	var existing *model.ProblemSet
	if id != 0 {
		var err error
		existing, err = s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, errcode.ErrProblemSetNotFound
			}
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		if req.Visibility == model.ProblemSetInviteOnly && inviteCode == "" {
			inviteCode = existing.InviteCode
		}
		if req.Visibility == model.ProblemSetInviteOnly && inviteCode == "" {
			return 0, errcode.ErrInvalidParams.WithMsg("选择邀请码可见时必须填写邀请码")
		}
	}

	problems, err := s.resolveProblems(ctx, req.Problems)
	if err != nil {
		return 0, err
	}
	tags := make([]model.ProblemSetTag, 0, len(req.TagIDs))
	for _, tagID := range req.TagIDs {
		tags = append(tags, model.ProblemSetTag{TagID: tagID})
	}

	set := &model.ProblemSet{
		ID:          id,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Published:   req.Published,
		Visibility:  req.Visibility,
		InviteCode:  inviteCode,
	}
	if id == 0 {
		publicID, err := newUniquePublicID(ctx, s.repo.PublicIDExists)
		if err != nil {
			return 0, err
		}
		set.PublicID = publicID
		set.CreatedBy = actorID
		if err := s.repo.CreateWithRelations(ctx, set, problems, tags); err != nil {
			return 0, errcode.ErrDatabase.Wrap(err)
		}
		return set.PublicID, nil
	}
	if err := s.repo.UpdateWithRelations(ctx, set, problems, tags); err != nil {
		return 0, errcode.ErrDatabase.Wrap(err)
	}
	return existing.PublicID, nil
}

func (s *ProblemSetService) ResolveID(ctx context.Context, publicID int64) (int64, error) {
	id, err := s.repo.ResolveIDByPublicID(ctx, publicID)
	if err == nil {
		return id, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrProblemSetNotFound
	}
	return 0, errcode.ErrDatabase.Wrap(err)
}

func (s *ProblemSetService) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrProblemSetNotFound
		}
		return errcode.ErrDatabase.Wrap(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	return nil
}

// hasAccess 判断能否看到题目列表。
func (s *ProblemSetService) hasAccess(ctx context.Context, set *model.ProblemSet, userID *int64, isAdmin bool) (bool, error) {
	if isAdmin || set.Visibility != model.ProblemSetInviteOnly {
		return true, nil
	}
	if userID == nil {
		return false, nil // 未登录无法解锁
	}
	unlocked, err := s.repo.IsUnlocked(ctx, *userID, set.ID)
	if err != nil {
		return false, errcode.ErrDatabase.Wrap(err)
	}
	return unlocked, nil
}

// resolveProblems 把对外题号列表转成关联行，数组下标即 Sort。
func (s *ProblemSetService) resolveProblems(ctx context.Context, displayIDs []string) ([]model.ProblemSetProblem, error) {
	problems := make([]model.ProblemSetProblem, 0, len(displayIDs))
	seen := make(map[int64]struct{}, len(displayIDs))
	for i, displayID := range displayIDs {
		problemID, err := s.problemRepo.ResolveID(ctx, strings.TrimSpace(displayID))
		if err != nil {
			return nil, errcode.ErrProblemNotFound.WithMsg("题目不存在：" + displayID)
		}
		if _, dup := seen[problemID]; dup {
			return nil, errcode.ErrInvalidParams.WithMsg("题目重复：" + displayID)
		}
		seen[problemID] = struct{}{}
		problems = append(problems, model.ProblemSetProblem{ProblemID: problemID, Sort: i})
	}
	return problems, nil
}

func (s *ProblemSetService) tagsBySetIDs(ctx context.Context, setIDs []int64) (map[int64][]dto.TagItem, error) {
	rawTags, err := s.repo.TagsBySetIDs(ctx, setIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	result := make(map[int64][]dto.TagItem, len(setIDs))
	for _, tag := range rawTags {
		result[tag.ProblemSetID] = append(result[tag.ProblemSetID], dto.TagItem{ID: tag.ID, Name: tag.Name})
	}
	return result, nil
}

func (s *ProblemSetService) authorNames(ctx context.Context, sets []model.ProblemSet) (map[int64]string, error) {
	ids := make([]int64, 0, len(sets))
	seen := make(map[int64]struct{}, len(sets))
	for _, set := range sets {
		if set.CreatedBy == 0 {
			continue
		}
		if _, ok := seen[set.CreatedBy]; ok {
			continue
		}
		seen[set.CreatedBy] = struct{}{}
		ids = append(ids, set.CreatedBy)
	}
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	names := make(map[int64]string, len(users))
	for _, user := range users {
		names[user.ID] = user.Username
	}
	return names, nil
}

// countSolved 统计给定题目里状态为「已通过」的个数。
func countSolved(problemIDs []int64, statusMap map[int64]int) int {
	solved := 0
	for _, problemID := range problemIDs {
		if statusMap[problemID] == model.UserProblemSolved {
			solved++
		}
	}
	return solved
}

// parseIDList 解析逗号分隔的 id 串，如 "1,2,3"；非法项直接跳过。
func parseIDList(raw string) []int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}
