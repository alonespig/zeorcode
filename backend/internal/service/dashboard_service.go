package service

import (
	"context"
	"time"

	"zoj/internal/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepo
}

func NewDashboardService(repo *repository.DashboardRepo) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Overview(ctx context.Context) (*DashboardOverview, error) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	trendStart := dayStart.AddDate(0, 0, -6)

	counts, err := s.repo.Counts(ctx, dayStart, now)
	if err != nil {
		return nil, err
	}
	trendRows, err := s.repo.SubmissionTrend(ctx, trendStart)
	if err != nil {
		return nil, err
	}
	recentRows, err := s.repo.RecentSubmissions(ctx, 8)
	if err != nil {
		return nil, err
	}

	trendByDate := make(map[string]int64, len(trendRows))
	for _, item := range trendRows {
		trendByDate[item.Date] = item.Count
	}
	trend := make([]DashboardTrendItem, 0, 7)
	for offset := 0; offset < 7; offset++ {
		date := trendStart.AddDate(0, 0, offset).Format("2006-01-02")
		trend = append(trend, DashboardTrendItem{Date: date, Count: trendByDate[date]})
	}

	recent := make([]DashboardSubmission, 0, len(recentRows))
	for _, item := range recentRows {
		recent = append(recent, DashboardSubmission{
			ID:          item.ID,
			UserID:      item.UserID,
			Username:    item.Username,
			ProblemID:   item.ProblemID,
			ProblemName: item.ProblemName,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt,
		})
	}

	return &DashboardOverview{
		Summary: DashboardSummary{
			Users:            counts.Users,
			Problems:         counts.Problems,
			TodaySubmissions: counts.TodaySubmissions,
			PendingPosts:     counts.PendingPosts,
			RunningContests:  counts.RunningContests,
			PendingJudging:   counts.PendingJudging,
		},
		SubmissionTrend:   trend,
		RecentSubmissions: recent,
	}, nil
}
