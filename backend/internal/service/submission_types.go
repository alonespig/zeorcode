package service

import "time"

// ===== 命令输入 =====

// SubmitCodeParams 提交代码的输入。
type SubmitCodeParams struct {
	ProblemID string
	Code      string
	Language  int
}

// ===== 通用值对象 =====

// SubmissionCaseResult 测试点结果（通用，Homework 也复用）。
type SubmissionCaseResult struct {
	ID         int
	Status     int
	TimeUsed   int64
	MemoryUsed int64
}

// ===== 结果 =====

// SubmissionAuthor 提交详情中的提交者信息。
type SubmissionAuthor struct {
	ID     int64
	Name   string
	Avatar string
}

// SubmissionProblem 提交详情中的题目信息。
type SubmissionProblem struct {
	ID          string
	Name        string
	Description string
	OJ          string
}

// SubmissionInfo 提交详情中的提交信息。
type SubmissionInfo struct {
	ID            int64
	Language      string
	Code          string
	CanViewCode   bool
	Status        int
	TimeUsed      int64
	MemoryUsed    int64
	CompileOutput string
	CreatedAt     time.Time
}

// SubmissionDetail 提交详情（用户 + 题目 + 提交 + 测试点）。
type SubmissionDetail struct {
	User        SubmissionAuthor
	Problem     SubmissionProblem
	Submission  SubmissionInfo
	CaseResults []SubmissionCaseResult
}

// SubmissionItem 提交列表项。
type SubmissionItem struct {
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
	CreatedAt   time.Time
}

type SubmissionList struct {
	Total int64
	List  []SubmissionItem
}

// DailyAcceptedCount 每日通过题数统计。
type DailyAcceptedCount struct {
	Date  []string
	Count []int
}
