package handler

import (
	"zoj/internal/http/dto"
	"zoj/internal/service"
)

func toDashboardOverviewResp(result *service.DashboardOverview) *dto.DashboardOverviewResp {
	trend := make([]dto.DashboardTrendItemResp, 0, len(result.SubmissionTrend))
	for _, item := range result.SubmissionTrend {
		trend = append(trend, dto.DashboardTrendItemResp{Date: item.Date, Count: item.Count})
	}
	recent := make([]dto.DashboardSubmissionResp, 0, len(result.RecentSubmissions))
	for _, item := range result.RecentSubmissions {
		recent = append(recent, dto.DashboardSubmissionResp{
			ID:          item.ID,
			UserID:      item.UserID,
			Username:    item.Username,
			ProblemID:   item.ProblemID,
			ProblemName: item.ProblemName,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &dto.DashboardOverviewResp{
		Summary: dto.DashboardSummaryResp{
			Users:            result.Summary.Users,
			Problems:         result.Summary.Problems,
			TodaySubmissions: result.Summary.TodaySubmissions,
			PendingPosts:     result.Summary.PendingPosts,
			RunningContests:  result.Summary.RunningContests,
			PendingJudging:   result.Summary.PendingJudging,
		},
		SubmissionTrend:   trend,
		RecentSubmissions: recent,
	}
}
