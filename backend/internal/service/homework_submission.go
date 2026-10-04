package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"zoj/internal/infra/logger"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

func (s *HomeworkService) Submit(ctx context.Context, id int64, req HomeworkSubmitParams, userID int64, isSiteAdmin bool) (int64, error) {
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

func (s *HomeworkService) Submissions(ctx context.Context, id, userID int64, isSiteAdmin bool, q HomeworkSubmissionQueryParams) (*HomeworkSubmissionList, error) {
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
	resp := &HomeworkSubmissionList{
		Total:     total,
		CanSeeAll: canSeeAll,
		List:      make([]HomeworkSubmissionItem, 0, len(list)),
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
		resp.List = append(resp.List, HomeworkSubmissionItem{
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
			CreatedAt:   sub.CreatedAt,
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
) (*HomeworkSubmissionDetail, error) {
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
	caseResults := make([]SubmissionCaseResult, 0, len(caseList))
	for idx, caseResult := range caseList {
		caseResults = append(caseResults, SubmissionCaseResult{
			ID:         idx + 1,
			Status:     caseResult.Status,
			TimeUsed:   caseResult.TimeUsed / nsPerMs,
			MemoryUsed: caseResult.MemoryUsed / bytesPerKB,
		})
	}

	return &HomeworkSubmissionDetail{
		ID:            submission.PublicID,
		UID:           user.UID,
		Username:      user.Username,
		StudentNo:     studentNoValue(user.StudentNo),
		RealName:      user.RealName,
		Avatar:        user.Avatar,
		ProblemID:     problem.DisplayID,
		ProblemName:   problem.Name,
		OJ:            problem.OJ,
		Status:        submission.Status,
		Score:         submission.Score,
		Language:      submission.Language,
		Code:          submission.Code,
		CompileOutput: submission.CompileOutput,
		TimeUsed:      submission.TimeUsed,
		MemoryUsed:    submission.MemoryUsed,
		CaseResults:   caseResults,
		InWindow:      hw.InWindow(submission.CreatedAt),
		CreatedAt:     submission.CreatedAt,
	}, nil
}

// ===== 内部工具 =====
