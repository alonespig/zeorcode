package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/infra/logger"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
	"zoj/pkg/judge"
	"zoj/pkg/util"

	"gorm.io/gorm"
)

func (s *ContestService) SubmitContestProblem(ctx context.Context, form ContestSubmitParams, userID int64) (int64, error) {
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
		if errors.Is(err, repository.ErrContestSubmissionClosed) {
			return 0, errcode.ErrBadRequest.WithMsg("比赛当前不接受提交，请刷新比赛状态")
		}
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
func (s *ContestService) GetContestSubmissions(ctx context.Context, contestID, userID int64, isAdmin bool, q *ContestSubmissionQueryParams) (*ContestSubmissionList, error) {
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
	resp := &ContestSubmissionList{
		Total:       int64(len(submissions)),
		Submissions: make([]ContestSubmissionItem, 0, len(submissions)),
	}
	for _, sub := range submissions {
		resp.Submissions = append(resp.Submissions, ContestSubmissionItem{
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
