package service

import (
	"time"

	"zoj/internal/dto"
)

// ===== 命令 / 查询输入 =====

// SaveHomeworkParams 布置/编辑作业的输入。StartTime/EndTime 为字符串，
// 由 parseHomeworkWindow 解析并校验。
type SaveHomeworkParams struct {
	Title       string
	Description string
	StartTime   string
	EndTime     string
	Problems    []string
}

// HomeworkSubmitParams 作业内提交的输入。
type HomeworkSubmitParams struct {
	ProblemID string
	Code      string
	Language  int
}

// HomeworkSubmissionQueryParams 作业提交列表筛选。
type HomeworkSubmissionQueryParams struct {
	Page      int
	PageSize  int
	UID       int64
	ProblemID string
	Status    *int
}

// ===== 结果 =====

// HomeworkItem 作业列表项。
type HomeworkItem struct {
	ID           int64
	Title        string
	Status       int
	ProblemCount int
	SolvedCount  *int
	StartTime    time.Time
	EndTime      time.Time
	Creator      string
	CanEdit      bool
	CanDelete    bool
}

type HomeworkList struct {
	Total int64
	List  []HomeworkItem
}

// HomeworkProblem 作业内的单题。
type HomeworkProblem struct {
	ID            string
	Name          string
	Status        *int
	Difficulty    int
	Tags          []TagItem // 题目标签，复用 problem 模块的 Service 类型
	MyScore       *int
	AcceptedCount int
	SubmitCount   int
}

type HomeworkDetail struct {
	ID            int64
	TeamID        int64
	TeamName      string
	Title         string
	Description   string
	Status        int
	StartTime     time.Time
	EndTime       time.Time
	Creator       string
	CreatorAvatar string
	ProblemCount  int
	SolvedCount   *int
	MyScore       *int
	TotalScore    int
	Locked        bool
	CanEdit       bool
	CanSeeAll     bool
	Problems      []HomeworkProblem
}

// RankCell 排行榜里某人某题的成绩。
type RankCell struct {
	ProblemID string
	Score     int
	Solved    bool
}

// HomeworkRankRow 排行榜一行。
type HomeworkRankRow struct {
	Rank        int
	UID         int64
	Username    string
	StudentNo   string
	RealName    string
	Avatar      string
	TotalScore  int
	SolvedCount int
	Cells       []RankCell
}

type HomeworkRank struct {
	StartTime  time.Time
	EndTime    time.Time
	ProblemIDs []string
	TotalScore int
	List       []HomeworkRankRow
}

// HomeworkSubmissionItem 作业提交列表项。
type HomeworkSubmissionItem struct {
	ID          int64
	UID         int64
	Username    string
	StudentNo   string
	RealName    string
	Avatar      string
	ProblemID   string
	ProblemName string
	Status      int
	Score       int
	Language    string
	TimeUsed    int64
	MemoryUsed  int64
	InWindow    bool
	CreatedAt   time.Time
}

type HomeworkSubmissionList struct {
	Total     int64
	CanSeeAll bool
	List      []HomeworkSubmissionItem
}

// HomeworkSubmissionDetail 作业提交弹窗详情。
type HomeworkSubmissionDetail struct {
	ID            int64
	UID           int64
	Username      string
	StudentNo     string
	RealName      string
	Avatar        string
	ProblemID     string
	ProblemName   string
	OJ            string
	Status        int
	Score         int
	Language      string
	Code          string
	CompileOutput string
	TimeUsed      int64
	MemoryUsed    int64
	CaseResults   []dto.SubmissionCaseResult // 评测用例结果，暂引用 submission 模块的 HTTP DTO
	InWindow      bool
	CreatedAt     time.Time
}
