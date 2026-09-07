package dto

// ===== 题单：前台 =====

// ProblemSetListReq 题单列表查询。Visibility 用指针以便区分「不筛选」和「筛公开」。
type ProblemSetListReq struct {
	Page       int    `form:"page" binding:"required"`
	PageSize   int    `form:"pageSize" binding:"required"`
	Keyword    string `form:"q"`
	Tags       string `form:"tags"`       // 逗号分隔标签 id，如 "1,2,3"（AND 交集）
	Visibility *int   `form:"visibility"` // 0 公开 / 1 需邀请码；不传=全部
}

// ProblemSetItemResp 列表项。列表不返回描述，描述只在详情页展示。
type ProblemSetItemResp struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Tags         []TagItem `json:"tags"`
	ProblemCount int       `json:"problemCount"`
	// SolvedCount 当前用户在本题单内已通过的题数；未登录时为 nil，前端不渲染进度
	SolvedCount *int   `json:"solvedCount,omitempty"`
	Visibility  int    `json:"visibility"` // 0 公开 / 1 需邀请码
	Published   int    `json:"published"`  // 0 草稿 / 1 已发布（仅后台列表会出现 0）
	Author      string `json:"author"`
	UpdatedAt   string `json:"updatedAt"`
}

type ProblemSetListResp struct {
	Total int                  `json:"total"`
	List  []ProblemSetItemResp `json:"list"`
}

// ProblemSetProblemResp 题单内的单题。Status 语义与题库列表一致：
// nil 未尝试 / 1 已通过 / 其他值为最近一次非 AC 判题状态码。
type ProblemSetProblemResp struct {
	ID            string    `json:"id"` // 对外题号
	Name          string    `json:"name"`
	Status        *int      `json:"status,omitempty"`
	Difficulty    int       `json:"difficulty"`
	Tags          []TagItem `json:"tags"`
	AcceptedCount int       `json:"acceptedCount"`
	SubmitCount   int       `json:"submitCount"`
}

// ProblemSetDetailResp 详情。Locked=true 时 Problems 为空，
// 前端展示「需要邀请码」的输入块；标题、描述、标签、题数仍可见。
type ProblemSetDetailResp struct {
	ID           int64                   `json:"id"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Tags         []TagItem               `json:"tags"`
	Visibility   int                     `json:"visibility"`
	Published    int                     `json:"published"`
	ProblemCount int                     `json:"problemCount"`
	SolvedCount  *int                    `json:"solvedCount,omitempty"`
	Locked       bool                    `json:"locked"`
	Author       string                  `json:"author"`
	UpdatedAt    string                  `json:"updatedAt"`
	Problems     []ProblemSetProblemResp `json:"problems"`
}

// UnlockProblemSetReq 提交邀请码解锁
type UnlockProblemSetReq struct {
	InviteCode string `json:"inviteCode" binding:"required"`
}

// ===== 题单：后台 =====

// SaveProblemSetReq 新建/编辑题单。Problems 的数组顺序即展示顺序。
type SaveProblemSetReq struct {
	Title       string   `json:"title" binding:"required,max=100"`
	Description string   `json:"description"`
	Published   int      `json:"published" binding:"oneof=0 1"`
	Visibility  int      `json:"visibility" binding:"oneof=0 1"`
	InviteCode  string   `json:"inviteCode" binding:"max=32"`
	TagIDs      []int64  `json:"tagIds"`
	Problems    []string `json:"problems"` // 对外题号列表，顺序敏感
}
