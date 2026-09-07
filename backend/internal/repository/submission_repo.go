package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"zoj/internal/model"
	"zoj/pkg/judge"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubmissionRepo struct {
	db *gorm.DB
}

func NewSubmissionRepo(db *gorm.DB) *SubmissionRepo {
	return &SubmissionRepo{db: db}
}

func (s *SubmissionRepo) PublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Submission{}).
		Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (s *SubmissionRepo) ResolveIDByPublicID(ctx context.Context, publicID int64) (int64, error) {
	var submission model.Submission
	err := s.db.WithContext(ctx).Select("id").Where("public_id = ?", publicID).First(&submission).Error
	return submission.ID, err
}

func (s *SubmissionRepo) GetSubmissionByPublicID(ctx context.Context, publicID int64) (*model.Submission, error) {
	var submission model.Submission
	err := s.db.WithContext(ctx).Where("public_id = ?", publicID).First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (s *SubmissionRepo) Create(ctx context.Context, submission *model.Submission) (int64, error) {
	err := s.db.WithContext(ctx).Create(submission).Error
	return submission.ID, err
}

// SubmissionDispatch 唯一标识一次需要投递的 Submission 版本。
type SubmissionDispatch struct {
	SubmissionID int64
	Version      int
}

// CreatePending 在同一事务内创建 Submission 和对应的 Outbox 记录。
// 任一步失败都会回滚，避免数据库已有 Pending Submission 却没有可恢复的投递记录。
func (s *SubmissionRepo) CreatePending(ctx context.Context, submission *model.Submission) (SubmissionDispatch, error) {
	dispatch := SubmissionDispatch{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(submission).Error; err != nil {
			return err
		}
		dispatch = SubmissionDispatch{SubmissionID: submission.ID, Version: submission.Version}
		return tx.Create(&model.SubmissionOutbox{
			SubmissionID:      dispatch.SubmissionID,
			SubmissionVersion: dispatch.Version,
			Status:            model.SubmissionOutboxPending,
			AvailableAt:       time.Now(),
		}).Error
	})
	return dispatch, err
}

// MarkDispatchPublished 标记指定 Submission 版本已经成功写入判题队列。
func (s *SubmissionRepo) MarkDispatchPublished(ctx context.Context, dispatch SubmissionDispatch) error {
	now := time.Now()
	return s.markDispatchPublishedQuery(ctx, dispatch, now).Error
}

func (s *SubmissionRepo) markDispatchPublishedQuery(
	ctx context.Context,
	dispatch SubmissionDispatch,
	publishedAt time.Time,
) *gorm.DB {
	return s.db.WithContext(ctx).Model(&model.SubmissionOutbox{}).
		Where("submission_id = ? AND submission_version = ? AND status IN ?",
			dispatch.SubmissionID, dispatch.Version,
			[]int{int(model.SubmissionOutboxPending), int(model.SubmissionOutboxProcessing)}).
		Updates(map[string]any{
			"status":       model.SubmissionOutboxPublished,
			"published_at": &publishedAt,
			"locked_by":    "",
			"locked_until": nil,
			"last_error":   "",
		})
}

// DispatchClaimResult 描述判题 worker 对某个 Submission 版本的原子认领结果。
type DispatchClaimResult uint8

const (
	DispatchClaimed DispatchClaimResult = iota
	DispatchDuplicate
	DispatchAlreadyCompleted
	DispatchStale
	DispatchSubmissionMissing
)

// ClaimDispatchForJudging 在同一事务中核对 Submission 版本和 Outbox 状态，并认领判题权。
// messageID 使用 Redis Stream 消息 ID：消息被其他消费者接管时 ID 不变，因此进程崩溃后可以续跑；
// 同一 Submission 版本的其他消息 ID 会被识别为重复投递。
func (s *SubmissionRepo) ClaimDispatchForJudging(
	ctx context.Context,
	dispatch SubmissionDispatch,
	messageID string,
) (DispatchClaimResult, error) {
	if dispatch.SubmissionID <= 0 || dispatch.Version < 0 || messageID == "" {
		return DispatchDuplicate, fmt.Errorf("invalid dispatch claim: submission=%d version=%d message=%q",
			dispatch.SubmissionID, dispatch.Version, messageID)
	}

	result := DispatchDuplicate
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var submission model.Submission
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "version", "status").
			First(&submission, dispatch.SubmissionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				result = DispatchSubmissionMissing
				return tx.Model(&model.SubmissionOutbox{}).
					Where("submission_id = ? AND submission_version = ? AND status <> ?",
						dispatch.SubmissionID, dispatch.Version, model.SubmissionOutboxCompleted).
					Updates(map[string]any{
						"status":       model.SubmissionOutboxCompleted,
						"locked_by":    "",
						"locked_until": nil,
					}).Error
			}
			return err
		}

		if submission.Version != dispatch.Version {
			result = DispatchStale
			return tx.Model(&model.SubmissionOutbox{}).
				Where("submission_id = ? AND submission_version = ? AND status <> ?",
					dispatch.SubmissionID, dispatch.Version, model.SubmissionOutboxCompleted).
				Updates(map[string]any{
					"status":       model.SubmissionOutboxCompleted,
					"locked_by":    "",
					"locked_until": nil,
				}).Error
		}

		var outbox model.SubmissionOutbox
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("submission_id = ? AND submission_version = ?", dispatch.SubmissionID, dispatch.Version).
			First(&outbox).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status := model.SubmissionOutboxJudging
			lockedBy := messageID
			now := time.Now()
			result = DispatchClaimed
			if submission.Status != judge.Pending {
				status = model.SubmissionOutboxCompleted
				lockedBy = ""
				result = DispatchAlreadyCompleted
			}
			return tx.Create(&model.SubmissionOutbox{
				SubmissionID:      dispatch.SubmissionID,
				SubmissionVersion: dispatch.Version,
				Status:            status,
				AvailableAt:       now,
				LockedBy:          lockedBy,
				PublishedAt:       &now,
			}).Error
		}
		if err != nil {
			return err
		}

		if submission.Status != judge.Pending {
			result = DispatchAlreadyCompleted
			if outbox.Status == model.SubmissionOutboxCompleted {
				return nil
			}
			return tx.Model(&model.SubmissionOutbox{}).Where("id = ?", outbox.ID).Updates(map[string]any{
				"status":       model.SubmissionOutboxCompleted,
				"locked_by":    "",
				"locked_until": nil,
			}).Error
		}

		switch outbox.Status {
		case model.SubmissionOutboxPending, model.SubmissionOutboxProcessing, model.SubmissionOutboxPublished:
			result = DispatchClaimed
			now := time.Now()
			return tx.Model(&model.SubmissionOutbox{}).Where("id = ?", outbox.ID).Updates(map[string]any{
				"status":       model.SubmissionOutboxJudging,
				"locked_by":    messageID,
				"locked_until": nil,
				"last_error":   "",
				"published_at": &now,
			}).Error
		case model.SubmissionOutboxJudging:
			if outbox.LockedBy == messageID {
				result = DispatchClaimed
			} else {
				result = DispatchDuplicate
			}
			return nil
		case model.SubmissionOutboxCompleted:
			result = DispatchAlreadyCompleted
			return nil
		default:
			return fmt.Errorf("unknown submission outbox status %d", outbox.Status)
		}
	})
	return result, err
}

// CompleteDispatchForJudging 仅允许当前 Redis 消息完成自己认领的判题任务。
// 已完成状态按幂等成功处理，便于结果已落库但 ACK 前崩溃的任务恢复。
func (s *SubmissionRepo) CompleteDispatchForJudging(
	ctx context.Context,
	dispatch SubmissionDispatch,
	messageID string,
) (bool, error) {
	result := s.db.WithContext(ctx).Model(&model.SubmissionOutbox{}).
		Where("submission_id = ? AND submission_version = ? AND status = ? AND locked_by = ?",
			dispatch.SubmissionID, dispatch.Version, model.SubmissionOutboxJudging, messageID).
		Updates(map[string]any{
			"status":       model.SubmissionOutboxCompleted,
			"locked_by":    "",
			"locked_until": nil,
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	var count int64
	err := s.db.WithContext(ctx).Model(&model.SubmissionOutbox{}).
		Where("submission_id = ? AND submission_version = ? AND status = ?",
			dispatch.SubmissionID, dispatch.Version, model.SubmissionOutboxCompleted).
		Count(&count).Error
	return count > 0, err
}

// ReleaseDispatchForRetry 释放当前消息的判题权，让一个新消息可以重新认领。
func (s *SubmissionRepo) ReleaseDispatchForRetry(
	ctx context.Context,
	dispatch SubmissionDispatch,
	messageID string,
) (bool, error) {
	result := s.db.WithContext(ctx).Model(&model.SubmissionOutbox{}).
		Where("submission_id = ? AND submission_version = ? AND status = ? AND locked_by = ?",
			dispatch.SubmissionID, dispatch.Version, model.SubmissionOutboxJudging, messageID).
		Updates(map[string]any{
			"status":       model.SubmissionOutboxPublished,
			"locked_by":    "",
			"locked_until": nil,
		})
	return result.RowsAffected > 0, result.Error
}

// SaveJudgingResult 原子写入本地判题终态与测试点结果。
// version 条件先锁定本次判题所属版本；若已被重判取代，事务不会触碰 JudgeResult。
func (s *SubmissionRepo) SaveJudgingResult(
	ctx context.Context,
	submission *model.Submission,
	results []*model.JudgeResult,
) (bool, error) {
	saved := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		update := tx.Model(&model.Submission{}).
			Where("id = ? AND version = ?", submission.ID, submission.Version).
			Updates(map[string]any{
				"status":         submission.Status,
				"time_used":      submission.TimeUsed,
				"memory_used":    submission.MemoryUsed,
				"compile_output": submission.CompileOutput,
				"score":          submission.Score,
			})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return nil
		}

		if err := tx.Where("submission_id = ?", submission.ID).Delete(&model.JudgeResult{}).Error; err != nil {
			return err
		}
		if len(results) > 0 {
			if err := tx.CreateInBatches(results, 100).Error; err != nil {
				return err
			}
		}
		saved = true
		return nil
	})
	return saved, err
}

// RecomputeSubmissionRow 重算聚合用的精简提交行
type RecomputeSubmissionRow struct {
	UserID    int64
	ProblemID int64
	Status    int
	Score     int
	CreatedAt time.Time
}

// ListContestSubmissionsForRecompute 取某比赛内所有"已判且非编译错误"的提交（排除 Pending / Compile Error，
// 因为二者本就不进比赛聚合），按 user、problem、提交时间升序，供从明细重算 UserContestProblem。
func (s *SubmissionRepo) ListContestSubmissionsForRecompute(ctx context.Context, contestID int64) ([]RecomputeSubmissionRow, error) {
	var rows []RecomputeSubmissionRow
	err := s.db.WithContext(ctx).
		Model(&model.Submission{}).
		Select("user_id, problem_id, status, score, created_at").
		Where("contest_id = ? AND status <> ? AND status <> ?", contestID, judge.Pending, judge.CompileError).
		Order("user_id ASC, problem_id ASC, created_at ASC").
		Scan(&rows).Error
	return rows, err
}

// ListUserProblemSubmissionsForRecompute 取某用户某题的非比赛(contest_id=0)、已判非CE提交，时间升序，供重算 UserProblem。
func (s *SubmissionRepo) ListUserProblemSubmissionsForRecompute(ctx context.Context, userID, problemID int64) ([]RecomputeSubmissionRow, error) {
	var rows []RecomputeSubmissionRow
	err := s.db.WithContext(ctx).
		Model(&model.Submission{}).
		Select("user_id, problem_id, status, score, created_at").
		Where("contest_id = 0 AND user_id = ? AND problem_id = ? AND status <> ? AND status <> ?",
			userID, problemID, judge.Pending, judge.CompileError).
		Order("created_at ASC").
		Scan(&rows).Error
	return rows, err
}

// ===== 重判（rejudge）=====

// RejudgeScope 重判范围：三种粒度二选一。
//   - SubmissionID>0：单条
//   - ContestID>0 且 ProblemID>0：某比赛内某题的全部提交
//   - ContestID>0 且 ProblemID=0：整场比赛
type RejudgeScope struct {
	SubmissionID int64
	ContestID    int64
	ProblemID    int64
}

// RejudgeTarget 待重判提交的最小信息（用于重新入队 + 精确清缓存）
type RejudgeTarget struct {
	ID        int64
	UserID    int64
	ProblemID int64
	ContestID int64
}

// GetRejudgeTargets 按范围取出要重判的提交清单。
func (s *SubmissionRepo) GetRejudgeTargets(ctx context.Context, scope RejudgeScope) ([]RejudgeTarget, error) {
	q := s.db.WithContext(ctx).Model(&model.Submission{}).Select("id, user_id, problem_id, contest_id")
	if scope.SubmissionID > 0 {
		q = q.Where("id = ?", scope.SubmissionID)
	} else {
		if scope.ContestID > 0 {
			q = q.Where("contest_id = ?", scope.ContestID)
		}
		if scope.ProblemID > 0 {
			q = q.Where("problem_id = ?", scope.ProblemID)
		}
	}
	var targets []RejudgeTarget
	err := q.Order("id ASC").Scan(&targets).Error
	return targets, err
}

// ResetForRejudge 事务内把一批提交重置为待重判，并为更新后的每个版本写入 Outbox：
// version+1（乐观锁，作废在途旧结果）、status→Pending、清结果、创建可靠投递记录。
func (s *SubmissionRepo) ResetForRejudge(ctx context.Context, ids []int64) ([]SubmissionDispatch, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	dispatches := make([]SubmissionDispatch, 0, len(ids))
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Submission{}).Where("id IN ?", ids).
			Updates(map[string]any{
				"version":        gorm.Expr("version + 1"),
				"status":         judge.Pending,
				"score":          0,
				"time_used":      0,
				"memory_used":    0,
				"compile_output": "",
			}).Error; err != nil {
			return err
		}
		if err := tx.Where("submission_id IN ?", ids).Delete(&model.JudgeResult{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Submission{}).
			Select("id AS submission_id, version").
			Where("id IN ?", ids).
			Order("id ASC").
			Scan(&dispatches).Error; err != nil {
			return err
		}
		if len(dispatches) == 0 {
			return nil
		}
		availableAt := time.Now()
		outboxes := make([]model.SubmissionOutbox, 0, len(dispatches))
		for _, dispatch := range dispatches {
			outboxes = append(outboxes, model.SubmissionOutbox{
				SubmissionID:      dispatch.SubmissionID,
				SubmissionVersion: dispatch.Version,
				Status:            model.SubmissionOutboxPending,
				AvailableAt:       availableAt,
			})
		}
		return tx.Create(&outboxes).Error
	})
	return dispatches, err
}

type ContestSubmissionFilter struct {
	UserID     *int64
	ContestID  int64
	ProblemIDs []int64
	Status     []int
}

func (s *SubmissionRepo) ContestSubmissionList(ctx context.Context, filter *ContestSubmissionFilter) ([]model.Submission, error) {
	var submissions []model.Submission
	query := s.db.WithContext(ctx).Model(&model.Submission{}).
		Select("id", "problem_id", "status", "created_at").
		Where("contest_id = ?", filter.ContestID).
		Order("created_at desc")

	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if len(filter.ProblemIDs) > 0 {
		query = query.Where("problem_id IN ?", filter.ProblemIDs)
	}

	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if err := query.Find(&submissions).Error; err != nil {
		return nil, err
	}

	return submissions, nil
}

func (s *SubmissionRepo) Delete(ctx context.Context, submissionID int64) error {
	return s.db.WithContext(ctx).Delete(&model.Submission{}, submissionID).Error
}

type SubmissionListItem struct {
	PublicID         int64     `gorm:"column:public_id"`
	ProblemID        int64     `gorm:"column:problem_id"`
	ProblemDisplayID string    `gorm:"column:problem_display_id"` // 对外题号
	ProblemName      string    `gorm:"column:problem_name"`
	UserID           int64     `gorm:"column:user_id"`
	UserUID          int64     `gorm:"column:user_uid"` // 对外用户号
	UserName         string    `gorm:"column:user_name"`
	Rating           int       `gorm:"column:rating"`
	Language         string    `gorm:"column:language"`
	Status           int       `gorm:"column:status"`
	TimeUsed         int64     `gorm:"column:time_used"`
	MemoryUsed       int64     `gorm:"column:memory_used"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

// ListWithInfo joins users and problems once for the submission list page,
// replacing the old per-row username/problem lookups.
// status 为 nil 时不按状态筛选；username 为空时不按用户名筛选（否则模糊匹配）。
// joins 放在 Count 之前，保证按 u.username 过滤时 count 与查询用的是同一组条件。
func (s *SubmissionRepo) ListWithInfo(ctx context.Context, page, pageSize int, status *int, username string, userID, problemID int64, includeHidden bool) ([]SubmissionListItem, int64, error) {
	var list []SubmissionListItem
	var total int64
	query := s.db.WithContext(ctx).Table("submissions AS s").
		Joins("JOIN users AS u ON u.id = s.user_id").
		Joins("JOIN problems AS p ON p.id = s.problem_id").
		Where("s.contest_id = 0 AND s.homework_id = 0")

	// 隐藏题的提交不进公开列表（管理员除外），否则会泄露隐藏题的存在
	if !includeHidden {
		query = query.Where("p.hidden = 0")
	}

	if status != nil {
		query = query.Where("s.status = ?", *status)
	}
	if username != "" {
		query = query.Where("u.username LIKE ?", "%"+username+"%")
	}
	if userID > 0 {
		query = query.Where("s.user_id = ?", userID)
	}
	if problemID > 0 {
		query = query.Where("s.problem_id = ?", problemID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Select(`s.public_id, s.problem_id, p.display_id AS problem_display_id, p.name AS problem_name,
			s.user_id, u.uid AS user_uid, u.username AS user_name, u.rating AS rating,
			s.language, s.status, s.time_used, s.memory_used, s.created_at`).
		Order("s.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func (s *SubmissionRepo) Update(ctx context.Context, submission *model.Submission) error {
	return s.db.WithContext(ctx).Save(submission).Error
}

func (s *SubmissionRepo) CreateSubmissionCaseResult(ctx context.Context, result *model.JudgeResult) error {
	return s.db.WithContext(ctx).Create(result).Error
}

func (s *SubmissionRepo) GetSubmissionCount(ctx context.Context, problemID int64, result *int) (int64, error) {
	var total int64
	query := s.db.WithContext(ctx).Model(&model.Submission{}).Where("problem_id = ?", problemID)
	if result != nil {
		query = query.Where("status = ?", *result)
	}
	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

type ProblemSubmissionStat struct {
	ProblemID     int64 `gorm:"column:problem_id"`
	SubmitCount   int64 `gorm:"column:submit_count"`
	AcceptedCount int64 `gorm:"column:accepted_count"`
}

// GetProblemSubmissionStats returns submit and accepted counts for many problems
// with one grouped query. The caller maps the rows by problem_id.
func (s *SubmissionRepo) GetProblemSubmissionStats(ctx context.Context, problemIDs []int64) ([]ProblemSubmissionStat, error) {
	if len(problemIDs) == 0 {
		return nil, nil
	}
	var stats []ProblemSubmissionStat
	err := s.db.WithContext(ctx).Model(&model.Submission{}).
		Select("problem_id, COUNT(*) AS submit_count, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS accepted_count", judge.Accepted).
		Where("problem_id IN ?", problemIDs).
		Group("problem_id").
		Scan(&stats).Error
	return stats, err
}

func (s *SubmissionRepo) GetContestProblemStats(ctx context.Context, contestID int64, problemIDs []int64) ([]ProblemSubmissionStat, error) {
	if len(problemIDs) == 0 {
		return nil, nil
	}
	var stats []ProblemSubmissionStat
	err := s.db.WithContext(ctx).Model(&model.Submission{}).
		Select("problem_id, COUNT(*) AS submit_count, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS accepted_count", judge.Accepted).
		Where("contest_id = ? AND problem_id IN ?", contestID, problemIDs).
		Group("problem_id").
		Scan(&stats).Error
	return stats, err
}

// GetHomeworkProblemStats returns submit and accepted counts scoped to one
// homework. Submissions made from the problem library, contests, or another
// homework are deliberately excluded.
func (s *SubmissionRepo) GetHomeworkProblemStats(ctx context.Context, homeworkID int64, problemIDs []int64) ([]ProblemSubmissionStat, error) {
	if len(problemIDs) == 0 {
		return nil, nil
	}
	var stats []ProblemSubmissionStat
	err := s.db.WithContext(ctx).Model(&model.Submission{}).
		Select("problem_id, COUNT(*) AS submit_count, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS accepted_count", judge.Accepted).
		Where("homework_id = ? AND problem_id IN ?", homeworkID, problemIDs).
		Group("problem_id").
		Scan(&stats).Error
	return stats, err
}

func (s *SubmissionRepo) GetSubmissionByID(ctx context.Context, submissionID int64) (*model.Submission, error) {
	var submission model.Submission
	err := s.db.WithContext(ctx).First(&submission, submissionID).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (s *SubmissionRepo) GetCaseResultBySubID(ctx context.Context, id int64) ([]model.JudgeResult, error) {
	var list []model.JudgeResult
	err := s.db.WithContext(ctx).Where("submission_id = ?", id).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

type SubmissionQuery struct {
	UserID    *int64
	ContestID int64
	Username  *string // 管理员按用户名模糊筛选
	ProblemID *int64  // 按题目筛选
	Status    *int    // 按判定结果筛选
}

type ContestSubmissionListItem struct {
	PublicID         int64     `gorm:"column:public_id"`
	ProblemID        int64     `gorm:"column:problem_id"`
	ProblemDisplayID string    `gorm:"column:problem_display_id"`
	ProblemName      string    `gorm:"column:problem_name"`
	ProblemLabel     string    `gorm:"column:problem_label"`
	UserID           int64     `gorm:"column:user_id"`
	UserUID          int64     `gorm:"column:user_uid"` // 对外用户号
	UserName         string    `gorm:"column:user_name"`
	Rating           int       `gorm:"column:rating"`
	Language         string    `gorm:"column:language"`
	Status           int       `gorm:"column:status"`
	TimeUsed         int64     `gorm:"column:time_used"`
	MemoryUsed       int64     `gorm:"column:memory_used"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

// GetContestSubmissionListWithInfo fetches display fields together with
// submissions, so the service does not query problem/contest_problem per row.
func (s *SubmissionRepo) GetContestSubmissionListWithInfo(ctx context.Context, q *SubmissionQuery) ([]ContestSubmissionListItem, error) {
	var list []ContestSubmissionListItem
	query := s.db.WithContext(ctx).Table("submissions AS s").
		Select(`s.public_id, s.problem_id, p.display_id AS problem_display_id, p.name AS problem_name, cp.label AS problem_label,
			s.user_id, u.uid AS user_uid, u.username AS user_name, u.rating AS rating, s.language, s.status, s.time_used,
			s.memory_used, s.created_at`).
		Joins("JOIN users AS u ON u.id = s.user_id").
		Joins("JOIN problems AS p ON p.id = s.problem_id").
		Joins("JOIN contest_problems AS cp ON cp.contest_id = s.contest_id AND cp.problem_id = s.problem_id").
		Where("s.contest_id = ?", q.ContestID)
	if q.UserID != nil {
		query = query.Where("s.user_id = ?", *q.UserID)
	}
	if q.Username != nil && *q.Username != "" {
		query = query.Where("u.username LIKE ?", "%"+*q.Username+"%")
	}
	if q.ProblemID != nil {
		query = query.Where("s.problem_id = ?", *q.ProblemID)
	}
	if q.Status != nil {
		query = query.Where("s.status = ?", *q.Status)
	}
	err := query.Order("s.created_at DESC").Scan(&list).Error
	return list, err
}

type DailyAcceptedCount struct {
	Date  time.Time `gorm:"column:date"`
	Count int       `gorm:"column:count"`
}

func (s *SubmissionRepo) GetUserRecentAcceptedSubmissions(ctx context.Context, userID int64, start time.Time) ([]model.Submission, error) {
	var result []model.Submission
	err := s.db.WithContext(ctx).Model(&model.Submission{}).
		Select("problem_id", "created_at").
		Where("user_id = ? AND status = ? AND created_at >= ?", userID, judge.Accepted, start).
		Order("created_at ASC").
		Find(&result).Error
	return result, err
}
