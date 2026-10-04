package judge

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/infra/logger"
	"zoj/internal/model"
	judgeapi "zoj/pkg/judge"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// saveResultGuarded 乐观锁写回评测结果：仅当 submissions.version 未被重判改动时才落库。
// 只更新结果相关列，绝不动 version（version 仅由重判 +1）。
// version 不符返回 errSubmissionSuperseded；数据库错误会进入 DLQ 并原样返回。
func (w *Worker) saveResultGuarded(ctx context.Context, s *model.Submission, stage string, log *zap.SugaredLogger) error {
	res := w.db.WithContext(ctx).Model(&model.Submission{}).
		Where("id = ? AND version = ?", s.ID, s.Version).
		Updates(map[string]any{
			"status":         s.Status,
			"time_used":      s.TimeUsed,
			"memory_used":    s.MemoryUsed,
			"compile_output": s.CompileOutput,
			"score":          s.Score,
		})
	if res.Error != nil {
		log.Errorw("save submission failed", "err", res.Error)
		w.sendToDLQ(s.ID, stage, res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		// version 对不上：判题期间被重判，这一轮已过时，安静丢弃，不发 SSE、不更新聚合
		log.Infow("submission superseded by rejudge, drop stale result", "submissionID", s.ID, "version", s.Version)
		return errSubmissionSuperseded
	}
	return nil
}

// finalizeSubmission 汇总 max(time/memory) → save 主提交 → 发 done
// saved=false 且 err=nil 表示该版本已被重判取代，不再推送或更新聚合。
func (w *Worker) finalizeSubmission(
	ctx context.Context,
	s *model.Submission,
	results []*model.JudgeResult,
	log *zap.SugaredLogger,
) (bool, error) {
	var timeUsed, memoryUsed int64
	for _, cr := range results {
		if t := cr.TimeUsed / 1_000_000; t > timeUsed {
			timeUsed = t
		}
		if m := cr.MemoryUsed / 1024; m > memoryUsed {
			memoryUsed = m
		}
	}
	s.TimeUsed = timeUsed
	s.MemoryUsed = memoryUsed

	saved, err := w.subRepo.SaveJudgingResult(ctx, s, results)
	if err != nil {
		log.Errorw("save submission and judge results failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageCaseSave, err)
		return false, err
	}
	if !saved {
		log.Infow("submission superseded by rejudge, drop stale result", "submissionID", s.ID, "version", s.Version)
		return false, nil
	}

	caseDtos := make([]SubmissionCaseResult, 0, len(results))
	for i, cr := range results {
		caseDtos = append(caseDtos, SubmissionCaseResult{
			ID:         i + 1,
			Status:     cr.Status,
			TimeUsed:   cr.TimeUsed / 1_000_000,
			MemoryUsed: cr.MemoryUsed / 1024,
		})
	}
	mp := map[string]any{
		"submission": SubmissionEventInfo{
			ID:         s.PublicID,
			Language:   s.Language,
			Status:     s.Status,
			TimeUsed:   s.TimeUsed,
			MemoryUsed: s.MemoryUsed,
			CreatedAt:  s.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"caseResults": caseDtos,
	}

	data, err := json.Marshal(mp)
	if err != nil {
		log.Errorw("marshal final result failed", "err", err)
		return true, nil
	}
	if err := w.mq.PublishDone(context.Background(), s.ID, string(data)); err != nil {
		log.Errorw("publish submission done failed", "err", err)
	}
	return true, nil
}

// updateUserProblem 非比赛提交：从 submissions 明细重算该用户该题的做题统计（幂等，重判安全）。
// 不再用 ++ 累加 —— 那样重判会重复计数；改为按当前所有有效提交现算。
func (w *Worker) updateUserProblem(s model.Submission, log *zap.SugaredLogger) {
	// AC 可能改变全站排名 → 换代使用户榜所有分页缓存失效
	if s.Status == judgeapi.Accepted {
		defer w.cache.Incr(context.Background(), cache.UserRankGen())
	}

	subs, err := w.subRepo.ListUserProblemSubmissionsForRecompute(context.Background(), s.UserID, s.ProblemID)
	if err != nil {
		log.Errorw("recompute user problem: list submissions failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageUserProblem, err)
		return
	}
	if len(subs) == 0 {
		return // 当前这条（非 CE、已判）理应在内；为空则无需处理
	}
	up := model.UserProblem{
		UserID:      s.UserID,
		ProblemID:   s.ProblemID,
		SubmitCount: len(subs),
	}
	for _, r := range subs {
		if r.Status == judgeapi.Accepted {
			up.AcCount++
		}
	}
	if up.AcCount > 0 {
		up.Status = judgeapi.Accepted
	} else {
		up.Status = subs[0].Status
	}
	if err := w.problemRepo.UpsertUserProblem(context.Background(), &up); err != nil {
		log.Errorw("recompute user problem: upsert failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageUserProblem, err)
	}
}

// caseAverageScore 测试点均分：通过点数 / 总点数 × 满分（四舍五入）。
// 比赛和作业共用，区别只在满分从哪来。
func caseAverageScore(results []*model.JudgeResult, full int) int {
	total := len(results)
	if total == 0 || full <= 0 {
		return 0
	}
	passed := 0
	for _, r := range results {
		if r.Status == judgeapi.Accepted {
			passed++
		}
	}
	return int(math.Round(float64(passed) / float64(total) * float64(full)))
}

// contestProblemFullScore 取比赛某题的满分，未配置则按 100。
func (w *Worker) contestProblemFullScore(contestID, problemID int64) int {
	if cp, err := w.contestRepo.GetContestProblem(context.Background(), model.ContestProblem{
		ContestID: contestID,
		ProblemID: problemID,
	}); err == nil && cp != nil && cp.Score > 0 {
		return cp.Score
	}
	return 100
}

// submissionScore 按提交归属算本次得分：
//   - 比赛提交：测试点均分 × 该题满分（OI/IOI 榜单用；ACM 忽略，但一并存下便于展示）
//   - 作业提交：测试点均分 × 100（作业不做逐题配分）
//   - 其他（普通练习）：不计分，恒 0
func (w *Worker) submissionScore(s *model.Submission, results []*model.JudgeResult) int {
	switch {
	case s.ContestID != 0:
		return caseAverageScore(results, w.contestProblemFullScore(s.ContestID, s.ProblemID))
	case s.HomeworkID != 0:
		return caseAverageScore(results, model.HomeworkProblemFullScore)
	default:
		return 0
	}
}

// updateScoreContestStats OI/IOI：按得分维护 UserContestProblem.Score
// IOI 取历史最高分；OI 取最后一次提交分。行锁串行化同一 (contest,user,problem)。
func (w *Worker) updateScoreContestStats(s model.Submission, contest model.Contest, log *zap.SugaredLogger) {
	submittedAt := s.CreatedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now()
	}
	err := w.db.Transaction(func(tx *gorm.DB) error {
		var ucp model.UserContestProblem
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND user_id = ? AND problem_id = ?", s.ContestID, s.UserID, s.ProblemID).
			First(&ucp).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ucp = model.UserContestProblem{
				ContestID: s.ContestID,
				UserID:    s.UserID,
				ProblemID: s.ProblemID,
				Status:    s.Status,
				Score:     s.Score,
				SubCount:  1,
				AcTime:    &submittedAt,
			}
			if s.Status == judgeapi.Accepted {
				ucp.AcCount = 1
			} else {
				ucp.UnAcCount = 1
			}
			return tx.Create(&ucp).Error
		}
		if err != nil {
			return err
		}

		ucp.SubCount++
		ucp.Status = s.Status
		if s.Status == judgeapi.Accepted {
			ucp.AcCount++
		} else {
			ucp.UnAcCount++
		}
		if contest.Type == model.ContestIOI {
			// IOI：分数提高才更新（记录达成时间，用于同分排序）
			if s.Score > ucp.Score {
				ucp.Score = s.Score
				ucp.AcTime = &submittedAt
			}
		} else {
			// OI：取最后一次提交分
			ucp.Score = s.Score
			ucp.AcTime = &submittedAt
		}
		return tx.Save(&ucp).Error
	})
	if err != nil {
		log.Errorw("update score contest problem failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageContestStats, err)
	}
}

// updateContestStats 比赛提交：维护比赛榜
// 注意：报名走 JoinContest 接口，这里不再补写 ContestUser
//
// 用事务 + 行锁串行化同一 (contest,user,problem) 的并发提交，避免并发下
// “都读到还没 AC → 都新增/都累加”造成的丢失更新或重复行；配合
// user_contest_problem 上的唯一索引 (contest_id,user_id,problem_id) 兜底。
func (w *Worker) updateContestStats(s model.Submission, log *zap.SugaredLogger) {
	// 比赛榜数据变了 → 让排行榜缓存失效，下次查实时重算
	defer w.cache.Delete(context.Background(), cache.ContestRank(s.ContestID))

	// OI/IOI 走得分榜：本次得分已在 processSubmission 落到 s.Score，这里维护 UserContestProblem.Score
	if contest, err := w.contestRepo.GetContestByID(context.Background(), s.ContestID); err == nil &&
		(contest.Type == model.ContestOI || contest.Type == model.ContestIOI) {
		if s.Status == judgeapi.CompileError {
			return
		}
		w.updateScoreContestStats(s, *contest, log)
		return
	}

	submittedAt := s.CreatedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now()
	}
	var ucp model.UserContestProblem
	isNewAccepted := false

	err := w.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND user_id = ? AND problem_id = ?", s.ContestID, s.UserID, s.ProblemID).
			First(&ucp).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 该用户该题第一次有记录
			ucp = model.UserContestProblem{
				ContestID: s.ContestID,
				UserID:    s.UserID,
				ProblemID: s.ProblemID,
				Status:    s.Status,
			}
			if s.Status == judgeapi.Accepted {
				ucp.AcCount = 1
				ucp.AcTime = &submittedAt
				isNewAccepted = true
			} else if s.Status != judgeapi.CompileError {
				ucp.UnAcCount = 1 // CE 不计入错误提交（对齐 CF/ICPC 惯例）
			}
			// 唯一索引兜底：并发下若另一 worker 抢先插入，这里会撞唯一键报错，
			// 事务回滚后交给 DLQ 重试，重试时改走下面“已存在→更新”分支。
			return tx.Create(&ucp).Error
		}
		if err != nil {
			return err
		}

		// 已有行且已加锁，安全地读-改-写
		if s.Status == judgeapi.Accepted && ucp.Status != judgeapi.Accepted {
			ucp.Status = judgeapi.Accepted
			ucp.AcCount = 1
			ucp.AcTime = &submittedAt
			isNewAccepted = true
		} else if s.Status != judgeapi.Accepted && ucp.Status != judgeapi.Accepted {
			ucp.Status = s.Status
			if s.Status != judgeapi.CompileError {
				ucp.UnAcCount++ // CE 不计入错误提交
			}
		}
		return tx.Save(&ucp).Error
	})
	if err != nil {
		log.Errorw("update user contest problem failed", "err", err)
		w.sendToDLQ(s.ID, dlqStageContestStats, err)
		return
	}

	w.updateContestCache(s, ucp, isNewAccepted, log)
}

func (w *Worker) updateContestCache(s model.Submission, ucp model.UserContestProblem, isNewAccepted bool, log *zap.SugaredLogger) {
	ctx := context.Background()

	// Redis stores running contest counters for polling-heavy pages. The database
	// remains the source of truth, so cache failures are warnings only.
	if err := w.cache.SAdd(ctx, cache.ContestProblemSubmittedUsers(s.ContestID, s.ProblemID), s.UserID); err != nil {
		log.Warnw("cache submitted user failed", "err", err)
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestProblemStats(s.ContestID, s.ProblemID), "submit_count", 1); err != nil {
		log.Warnw("cache submit count failed", "err", err)
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestUserStats(s.ContestID, s.UserID), "submit_count", 1); err != nil {
		log.Warnw("cache user submit count failed", "err", err)
	}

	acTime := int64(0)
	if ucp.AcTime != nil {
		acTime = ucp.AcTime.Unix()
	}
	if err := w.cache.HSet(ctx, cache.ContestUserProblem(s.ContestID, s.UserID, s.ProblemID),
		"status", ucp.Status,
		"tries", ucp.UnAcCount,
		"ac_time", acTime,
		"updated_at", time.Now().Unix(),
	); err != nil {
		log.Warnw("cache user problem status failed", "err", err)
	}

	if s.Status != judgeapi.Accepted {
		return
	}
	if err := w.cache.SAdd(ctx, cache.ContestProblemAcceptedUsers(s.ContestID, s.ProblemID), s.UserID); err != nil {
		log.Warnw("cache accepted user failed", "err", err)
	}
	if err := w.cache.SAdd(ctx, cache.ContestUserAcceptedProblems(s.ContestID, s.UserID), s.ProblemID); err != nil {
		log.Warnw("cache accepted problem failed", "err", err)
	}
	if !isNewAccepted {
		return
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestProblemStats(s.ContestID, s.ProblemID), "accepted_count", 1); err != nil {
		log.Warnw("cache accepted count failed", "err", err)
	}
	if _, err := w.cache.HIncrBy(ctx, cache.ContestUserStats(s.ContestID, s.UserID), "pass_count", 1); err != nil {
		log.Warnw("cache user pass count failed", "err", err)
	}
	if ucp.AcTime == nil {
		return
	}
	contest, err := w.contestRepo.GetContestByID(context.Background(), s.ContestID)
	if err != nil {
		log.Warnw("get contest for cache penalty failed", "err", err)
		return
	}
	penalty := int(ucp.AcTime.Sub(contest.StartTime).Minutes()) + 20*ucp.UnAcCount
	if _, err := w.cache.HIncrBy(ctx, cache.ContestUserStats(s.ContestID, s.UserID), "penalty", int64(penalty)); err != nil {
		log.Warnw("cache user penalty failed", "err", err)
	}
}

// sendToDLQ 把致命错误的上下文 JSON 推到死信队列
// 本函数不再上报错误（避免嵌套），只 log
func (w *Worker) sendToDLQ(submissionID int64, stage string, cause error) {
	payload, _ := json.Marshal(map[string]any{
		"submissionID": submissionID,
		"stage":        stage,
		"error":        cause.Error(),
		"ts":           time.Now().Unix(),
	})
	if err := w.mq.DeadLetterSubmission(context.Background(), string(payload)); err != nil {
		logger.Errorw("push DLQ failed", "submissionID", submissionID, "stage", stage, "err", err)
	}
}
