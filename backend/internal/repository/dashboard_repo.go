package repository

import (
	"context"
	"time"

	"zoj/internal/model"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

type DashboardRepo struct {
	db *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

type DashboardCounts struct {
	Users            int64
	Problems         int64
	TodaySubmissions int64
	PendingPosts     int64
	RunningContests  int64
	PendingJudging   int64
}

type DashboardDailyCount struct {
	Date  string
	Count int64
}

type DashboardRecentSubmission struct {
	ID          int64
	UserID      int64
	Username    string
	ProblemID   string
	ProblemName string
	Status      int
	CreatedAt   time.Time
}

func (r *DashboardRepo) Counts(ctx context.Context, dayStart, now time.Time) (*DashboardCounts, error) {
	db := r.db.WithContext(ctx)
	result := &DashboardCounts{}
	queries := []struct {
		model any
		where string
		args  []any
		dest  *int64
	}{
		{model: &model.User{}, dest: &result.Users},
		{model: &model.Problem{}, dest: &result.Problems},
		{model: &model.Submission{}, where: "created_at >= ?", args: []any{dayStart}, dest: &result.TodaySubmissions},
		{model: &model.Post{}, where: "status = ? AND review_status = ?", args: []any{0, model.PostReviewPending}, dest: &result.PendingPosts},
		{model: &model.Contest{}, where: "archived = ? AND start_time <= ? AND end_time > ?", args: []any{false, now, now}, dest: &result.RunningContests},
		{model: &model.Submission{}, where: "status = ?", args: []any{judge.Pending}, dest: &result.PendingJudging},
	}
	for _, query := range queries {
		statement := db.Model(query.model)
		if query.where != "" {
			statement = statement.Where(query.where, query.args...)
		}
		if err := statement.Count(query.dest).Error; err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *DashboardRepo) SubmissionTrend(ctx context.Context, since time.Time) ([]DashboardDailyCount, error) {
	var rows []DashboardDailyCount
	err := r.db.WithContext(ctx).Model(&model.Submission{}).
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("DATE(created_at)").
		Scan(&rows).Error
	return rows, err
}

func (r *DashboardRepo) RecentSubmissions(ctx context.Context, limit int) ([]DashboardRecentSubmission, error) {
	var rows []DashboardRecentSubmission
	err := r.db.WithContext(ctx).Table("submissions AS s").
		Select(`s.public_id AS id, u.uid AS user_id, u.username,
			p.display_id AS problem_id, p.name AS problem_name, s.status, s.created_at`).
		Joins("JOIN users AS u ON u.id = s.user_id").
		Joins("JOIN problems AS p ON p.id = s.problem_id").
		Order("s.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}
