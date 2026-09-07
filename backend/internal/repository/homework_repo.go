package repository

import (
	"context"
	"time"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type HomeworkRepo struct {
	db *gorm.DB
}

func NewHomeworkRepo(db *gorm.DB) *HomeworkRepo {
	return &HomeworkRepo{db: db}
}

// HomeworkProblemRef 作业-题目引用（按 sort 升序）
type HomeworkProblemRef struct {
	HomeworkID int64 `gorm:"column:homework_id"`
	ProblemID  int64 `gorm:"column:problem_id"`
	Sort       int   `gorm:"column:sort"`
}

// HomeworkBestScore 某人某题在时间窗内的最高分及其达成时刻。
// 排行榜同分排序要用 AchievedAt：同分时达成早者靠前。
type HomeworkBestScore struct {
	UserID     int64     `gorm:"column:user_id"`
	ProblemID  int64     `gorm:"column:problem_id"`
	Best       int       `gorm:"column:best"`
	AchievedAt time.Time `gorm:"column:achieved_at"`
}

func (r *HomeworkRepo) ListByTeam(ctx context.Context, teamID int64, page, pageSize int) ([]model.Homework, int64, error) {
	var list []model.Homework
	var total int64
	db := r.db.WithContext(ctx).Model(&model.Homework{}).Where("team_id = ?", teamID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("start_time DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *HomeworkRepo) GetByID(ctx context.Context, id int64) (*model.Homework, error) {
	var hw model.Homework
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&hw).Error
	return &hw, err
}

func (r *HomeworkRepo) PublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Homework{}).
		Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *HomeworkRepo) ResolveIDByPublicID(ctx context.Context, publicID int64) (int64, error) {
	var homework model.Homework
	err := r.db.WithContext(ctx).Select("id").
		Where("public_id = ?", publicID).First(&homework).Error
	return homework.ID, err
}

func (r *HomeworkRepo) GetBySourceActionID(ctx context.Context, actionID int64) (*model.Homework, error) {
	var homework model.Homework
	err := r.db.WithContext(ctx).Where("source_action_id = ?", actionID).First(&homework).Error
	return &homework, err
}

// ProblemRefsByHomeworkIDs 批量取题目引用，按 (作业, sort) 升序，列表页算题数用。
func (r *HomeworkRepo) ProblemRefsByHomeworkIDs(ctx context.Context, ids []int64) ([]HomeworkProblemRef, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var refs []HomeworkProblemRef
	err := r.db.WithContext(ctx).Table("homework_problems").
		Select("homework_id, problem_id, sort").
		Where("homework_id IN ?", ids).
		Order("homework_id ASC, sort ASC").
		Find(&refs).Error
	return refs, err
}

// CreateWithProblems 原子创建作业及其题目关联。
func (r *HomeworkRepo) CreateWithProblems(ctx context.Context, hw *model.Homework, problems []model.HomeworkProblem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(hw).Error; err != nil {
			return err
		}
		return writeHomeworkProblems(tx, hw.ID, problems)
	})
}

// UpdateWithProblems 原子更新作业：主记录 Updates + 题目关联「先删后插」全量替换，
// 与 ContestRepo.UpdateContestWithProblems 同一套做法。
func (r *HomeworkRepo) UpdateWithProblems(ctx context.Context, hw *model.Homework, problems []model.HomeworkProblem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Homework{}).Where("id = ?", hw.ID).
			Select("title", "description", "start_time", "end_time").
			Updates(hw).Error
		if err != nil {
			return err
		}
		if err := tx.Where("homework_id = ?", hw.ID).Delete(&model.HomeworkProblem{}).Error; err != nil {
			return err
		}
		return writeHomeworkProblems(tx, hw.ID, problems)
	})
}

func writeHomeworkProblems(tx *gorm.DB, hwID int64, problems []model.HomeworkProblem) error {
	if len(problems) == 0 {
		return nil
	}
	for i := range problems {
		problems[i].HomeworkID = hwID
	}
	return tx.Create(&problems).Error
}

func (r *HomeworkRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("homework_id = ?", id).Delete(&model.HomeworkProblem{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Homework{}).Error
	})
}

// BestScores 取某作业在时间窗内每人每题的最高分及其最早达成时刻。
// 排行榜实时聚合，不建聚合表：团队规模是几十到几百人 × 几到几十题，一条
// GROUP BY 完全够；维护聚合表要动判题 worker 的写回路径，风险和复杂度高得多。
//
// achieved_at 要的是「取到最高分的最早时间」，不是「最早提交时间」，
// 所以用 MIN(CASE WHEN score = 该组最高分 THEN created_at END)，
// 借子查询把每组最高分先算出来。
func (r *HomeworkRepo) BestScores(ctx context.Context, homeworkID int64, start, end time.Time) ([]HomeworkBestScore, error) {
	return r.bestScores(ctx, homeworkID, 0, start, end)
}

// BestScoresByUser 只查询指定用户的作业最佳成绩，供个人进度展示使用。
func (r *HomeworkRepo) BestScoresByUser(ctx context.Context, homeworkID, userID int64, start, end time.Time) ([]HomeworkBestScore, error) {
	return r.bestScores(ctx, homeworkID, userID, start, end)
}

func (r *HomeworkRepo) bestScores(ctx context.Context, homeworkID, userID int64, start, end time.Time) ([]HomeworkBestScore, error) {
	var rows []HomeworkBestScore
	sub := r.db.WithContext(ctx).Table("submissions").
		Select("user_id, problem_id, MAX(score) AS best").
		Where("homework_id = ? AND created_at >= ? AND created_at <= ?", homeworkID, start, end).
		Group("user_id, problem_id")
	if userID != 0 {
		sub = sub.Where("user_id = ?", userID)
	}

	query := r.db.WithContext(ctx).Table("submissions AS s").
		Select("s.user_id, s.problem_id, b.best AS best, MIN(s.created_at) AS achieved_at").
		Joins("JOIN (?) AS b ON b.user_id = s.user_id AND b.problem_id = s.problem_id AND b.best = s.score", sub).
		Where("s.homework_id = ? AND s.created_at >= ? AND s.created_at <= ?", homeworkID, start, end)
	if userID != 0 {
		query = query.Where("s.user_id = ?", userID)
	}
	err := query.
		Group("s.user_id, s.problem_id, b.best").
		Find(&rows).Error
	return rows, err
}

// SolvedCountsByUser 各成员在该作业时间窗内的满分题数（用于列表页「我的进度」）。
func (r *HomeworkRepo) SolvedCountsByUser(ctx context.Context, homeworkIDs []int64, userID int64, full int) (map[int64]int, error) {
	if userID == 0 || len(homeworkIDs) == 0 {
		return map[int64]int{}, nil
	}
	type row struct {
		HomeworkID int64 `gorm:"column:homework_id"`
		Cnt        int   `gorm:"column:cnt"`
	}
	var rows []row
	// 只数「在各自作业时间窗内拿到满分」的题：先按 (作业,题) 取最高分，再筛满分
	sub := r.db.WithContext(ctx).Table("submissions AS s").
		Select("s.homework_id, s.problem_id, MAX(s.score) AS best").
		Joins("JOIN homeworks h ON h.id = s.homework_id").
		Where("s.homework_id IN ? AND s.user_id = ?", homeworkIDs, userID).
		Where("s.created_at >= h.start_time AND s.created_at <= h.end_time").
		Group("s.homework_id, s.problem_id")

	err := r.db.WithContext(ctx).Table("(?) AS t", sub).
		Select("t.homework_id, COUNT(*) AS cnt").
		Where("t.best >= ?", full).
		Group("t.homework_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int, len(rows))
	for _, r := range rows {
		out[r.HomeworkID] = r.Cnt
	}
	return out, nil
}

// HomeworkSubmissionFilter 作业提交列表筛选
type HomeworkSubmissionFilter struct {
	HomeworkID int64
	UserID     int64 // 0 表示不限（仅有权看全部的人可用）
	ProblemID  int64
	Status     *int
	Page       int
	PageSize   int
}

func (r *HomeworkRepo) ListSubmissions(ctx context.Context, f *HomeworkSubmissionFilter) ([]model.Submission, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.Submission{}).
		Where("homework_id = ?", f.HomeworkID)
	if f.UserID != 0 {
		db = db.Where("user_id = ?", f.UserID)
	}
	if f.ProblemID != 0 {
		db = db.Where("problem_id = ?", f.ProblemID)
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Submission
	err := db.Order("id DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
