package service

import (
	"time"

	"zoj/internal/model"
)

// ===== 命令 / 查询输入 =====

// ContestProblemInput 创建/编辑比赛时每道题的信息（对外题号 + 气球颜色 + 满分）。
type ContestProblemInput struct {
	ProblemID string
	Color     string
	Score     int
}

type CreateContestParams struct {
	Name        string
	ContestDate string
	ContestTime string
	Description string
	CoverURL    string
	Duration    int
	Type        int
	Problems    []ContestProblemInput
	InviteCode  string
	Rated       bool
}

type UpdateContestParams struct {
	Name        string
	Description string
	CoverURL    string
	Type        int
	StartTime   int64
	EndTime     int64
	Problems    []ContestProblemInput
	Rated       bool
}

type ContestSubmitParams struct {
	ContestID int64
	Label     string
	Code      string
	Language  int
}

type ContestSubmissionQueryParams struct {
	Username  string
	ProblemID string
	Status    *int
}

// ===== 结果 =====

type ContestItem struct {
	ID             int64
	Name           string
	Description    string
	CoverURL       string
	StartTime      time.Time
	EndTime        time.Time
	Type           string
	Status         model.ContestStatus
	IsRegistered   bool
	Rated          bool
	NeedInviteCode bool
	Duration       int
	Participants   int
	Archived       bool
}

type ContestList struct {
	Total int64
	List  []ContestItem
}

type ContestDesc struct {
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Description string
}

type ContestDetail struct {
	Name           string
	StartTime      time.Time
	EndTime        time.Time
	IsRegistered   bool
	NeedInviteCode bool
	Rule           string
	Rated          bool
	Settled        bool
}

type ContestProblem struct {
	ProblemID     string
	Name          string
	Label         string
	Color         string
	Status        int
	TimeLimit     int
	MemoryLimit   int
	TotalCount    int64
	AcceptedCount int64
}

type ContestProblemList struct {
	ProblemList []ContestProblem
}

type ContestProblemDetail struct {
	Name         string
	TimeLimit    int
	MemoryLimit  int
	Description  string
	InputFormat  string
	OutputFormat string
	Hint         string
	Samples      []ProblemSample
	HasTestData  bool
}

type ContestProblemItem struct {
	Name  string
	Label string
}

type ContestSubmitInfo struct {
	ProblemList []ContestProblemItem
}

type ContestEasyProblem struct {
	ID    string
	Name  string
	Color string
	Score int
}

type ContestEditInfo struct {
	Name        string
	Description string
	CoverURL    string
	Type        int
	StartTime   time.Time
	EndTime     time.Time
	ProblemList []ContestEasyProblem
	Rated       bool
}

type ContestSubmissionItem struct {
	ID          int64
	ProblemID   string
	ProblemName string
	UserID      int64
	UserName    string
	Rating      int
	Language    string
	Result      int
	TimeUsed    int64
	MemoryUsed  int64
	CreatedAt   string
}

type ContestSubmissionList struct {
	Total       int64
	Submissions []ContestSubmissionItem
}

type CreateContestResult struct {
	ID int64
}
