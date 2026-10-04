package service

import "time"

// ===== 查询 / 命令输入 =====

// ProblemSetListParams 题单列表查询。Visibility 用指针区分「不筛选」和「筛公开」。
type ProblemSetListParams struct {
	Page       int
	PageSize   int
	Keyword    string
	Tags       string
	Visibility *int
}

// SaveProblemSetParams 新建/编辑题单。Problems 的数组顺序即展示顺序。
type SaveProblemSetParams struct {
	Title       string
	Description string
	Published   int
	Visibility  int
	InviteCode  string
	TagIDs      []int64
	Problems    []string
}

// ===== 结果 =====

// ProblemSetItem 题单列表项。列表不返回描述，描述只在详情页展示。
type ProblemSetItem struct {
	ID           int64
	Title        string
	Tags         []TagItem
	ProblemCount int
	SolvedCount  *int
	Visibility   int
	Published    int
	Author       string
	UpdatedAt    time.Time
}

type ProblemSetList struct {
	Total int64
	List  []ProblemSetItem
}

// ProblemSetProblem 题单内的单题。
type ProblemSetProblem struct {
	ID            string
	Name          string
	Status        *int
	Difficulty    int
	Tags          []TagItem
	AcceptedCount int
	SubmitCount   int
}

// ProblemSetDetail 题单详情。
type ProblemSetDetail struct {
	ID           int64
	Title        string
	Description  string
	Tags         []TagItem
	Visibility   int
	Published    int
	ProblemCount int
	SolvedCount  *int
	Locked       bool
	Author       string
	UpdatedAt    time.Time
	Problems     []ProblemSetProblem
}
