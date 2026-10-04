package service

import "time"

// ===== 查询 / 命令输入 =====

// ProblemListParams 题目列表查询。
type ProblemListParams struct {
	Page       int
	PageSize   int
	Keyword    string
	Order      string
	Difficulty *int
	Tags       string
}

// CreateProblemParams 创建题目的输入。
type CreateProblemParams struct {
	DisplayID       string
	Name            string
	Difficulty      int
	TimeLimit       int
	MemoryLimit     int
	Description     string
	InputFormat     string
	OutputFormat    string
	Hint            string
	Samples         []ProblemSample
	TagsID          []int64
	OJ              string
	RemoteProblemID string
	Hidden          bool
}

// UpdateProblemParams 更新题目的输入。
type UpdateProblemParams struct {
	ID           int64
	DisplayID    string
	Name         string
	Difficulty   int
	TimeLimit    int
	MemoryLimit  int
	Description  string
	InputFormat  string
	OutputFormat string
	Hint         string
	Samples      []ProblemSample
	Tags         []int64
	Hidden       bool
}

// ===== 通用值对象（可被 Homework / ProblemSet 等复用）=====

// TagItem 题目标签。
type TagItem struct {
	ID   int64
	Name string
}

// ProblemSample 题目样例。
type ProblemSample struct {
	Input   string
	Output  string
	Explain string
}

// ===== 结果 =====

// ProblemItem 题目列表项。
type ProblemItem struct {
	ID            string
	Name          string
	Status        *int
	Difficulty    int
	Tags          []TagItem
	AcceptedCount int
	SubmitCount   int
	CreatedAt     time.Time
	Hidden        bool
	OJ            string
}

type ProblemList struct {
	Total int64
	List  []ProblemItem
}

// ProblemDetail 题目详情。
type ProblemDetail struct {
	ID              string
	Name            string
	Difficulty      int
	TimeLimit       int
	MemoryLimit     int
	Description     string
	InputFormat     string
	OutputFormat    string
	SubmitCount     int
	AcceptedCount   int
	Hint            string
	Samples         []ProblemSample
	Tags            []TagItem
	OJ              string
	RemoteProblemID string
	Hidden          bool
	HasTestData     bool
}

// ProblemForEdit 题目编辑回显。
type ProblemForEdit struct {
	ID           string
	Name         string
	Difficulty   int
	TimeLimit    int
	MemoryLimit  int
	Description  string
	InputFormat  string
	OutputFormat string
	Hint         string
	Samples      []ProblemSample
	Tags         []TagItem
	Hidden       bool
}

// TagList 标签列表。
type TagList struct {
	Tags []TagItem
}

// AdminTagItem 管理端标签项。
type AdminTagItem struct {
	ID              int64
	Name            string
	ProblemCount    int64
	ProblemSetCount int64
}
