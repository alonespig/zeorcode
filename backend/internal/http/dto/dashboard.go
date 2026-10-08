package dto

type DashboardSummaryResp struct {
	Users            int64 `json:"users"`
	Problems         int64 `json:"problems"`
	TodaySubmissions int64 `json:"todaySubmissions"`
	PendingPosts     int64 `json:"pendingPosts"`
	RunningContests  int64 `json:"runningContests"`
	PendingJudging   int64 `json:"pendingJudging"`
}

type DashboardTrendItemResp struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type DashboardSubmissionResp struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userID"`
	Username    string `json:"username"`
	ProblemID   string `json:"problemID"`
	ProblemName string `json:"problemName"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type DashboardOverviewResp struct {
	Summary           DashboardSummaryResp      `json:"summary"`
	SubmissionTrend   []DashboardTrendItemResp  `json:"submissionTrend"`
	RecentSubmissions []DashboardSubmissionResp `json:"recentSubmissions"`
}
