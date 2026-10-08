package service

import "time"

type DashboardSummary struct {
	Users            int64
	Problems         int64
	TodaySubmissions int64
	PendingPosts     int64
	RunningContests  int64
	PendingJudging   int64
}

type DashboardTrendItem struct {
	Date  string
	Count int64
}

type DashboardSubmission struct {
	ID          int64
	UserID      int64
	Username    string
	ProblemID   string
	ProblemName string
	Status      int
	CreatedAt   time.Time
}

type DashboardOverview struct {
	Summary           DashboardSummary
	SubmissionTrend   []DashboardTrendItem
	RecentSubmissions []DashboardSubmission
}
