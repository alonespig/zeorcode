package dto

type SubmitCodeReq struct {
	ProblemID string `json:"problemID" binding:"required"` // 对外题号
	Code      string `json:"code" binding:"required"`
	Language  int    `json:"language" binding:"required"`
}

type SubmissionItemResp struct {
	ID          int64  `json:"id"`        // 对外提交编号
	ProblemID   string `json:"problemID"` // 对外题号
	ProblemName string `json:"problemName"`
	UserID      int64  `json:"userID"`
	UserName    string `json:"userName"`
	Rating      int    `json:"rating"`
	Language    string `json:"language"`
	Result      int    `json:"result"`
	TimeUsed    int64  `json:"timeUsed"`
	MemoryUsed  int64  `json:"memoryUsed"`
	CreatedAt   string `json:"createdAt"`
}

type SubmitListResp struct {
	Total int64                `json:"total"`
	List  []SubmissionItemResp `json:"list"`
}

// SubmissionListForm 评测列表的查询参数：分页 + 可选的状态/用户名筛选。
// Status 用指针，未传时为 nil（不筛选），可正确区分“未传”与“状态码 0(Pending)”。
type SubmissionListForm struct {
	PageForm
	Status    *int   `form:"status"`
	Username  string `form:"username"`
	UserID    int64  `form:"userID"`    // >0 时只看该用户的提交（个人主页"提交记录"tab，按 id 精确匹配）
	ProblemID string `form:"problemID"` // 非空时只看该题的提交（题目页"本题提交记录"用；传对外题号）
}

type CaseResultResp struct {
	ID         int   `json:"id"`
	Result     int   `json:"status"`
	TimeUsed   int64 `json:"time"`
	MemoryUsed int64 `json:"memory"`
}

type SubmissionDetailResp struct {
	ID          int64            `json:"id"`
	ProblemID   int64            `json:"problemID"`
	ProblemName string           `json:"problemName"`
	Description string           `json:"description"`
	UserID      int64            `json:"userID"`
	UserName    string           `json:"userName"`
	Language    string           `json:"language"`
	Code        string           `json:"code"`
	Result      int              `json:"status"`
	TimeUsed    int64            `json:"time"`
	MemoryUsed  int64            `json:"memory"`
	CaseResults []CaseResultResp `json:"caseResults"`
	CreatedAt   string           `json:"createdAt"`
}

type UserDetail struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Avatar    string  `json:"avatar"`
	Gender    int     `json:"gender"`
	Email     string  `json:"email,omitempty"`
	Signature *string `json:"signature,omitempty"`
	School    string  `json:"school,omitempty"`
	CreatedAt string  `json:"createdAt,omitempty"`
}

type SubmissionProblem struct {
	ID          string `json:"id"` // 对外题号
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Status      int    `json:"status"`
	OJ          string `json:"oj"` // 非空=远程题(前端据此隐藏测试点)
}

type SubmissionInfo struct {
	ID            int64  `json:"id"` // 对外提交编号
	Language      string `json:"language"`
	Code          string `json:"code"`
	CanViewCode   bool   `json:"canViewCode"`
	Status        int    `json:"status"`
	TimeUsed      int64  `json:"time"`
	MemoryUsed    int64  `json:"memory"`
	CompileOutput string `json:"compileOutput"`
	CreatedAt     string `json:"createdAt"`
}

// SubmissionEventInfo 是判题进度事件中的提交快照。
// 实时推送端点只传递状态，不携带源代码，避免绕过提交详情接口的权限控制。
type SubmissionEventInfo struct {
	ID         int64  `json:"id"` // 对外提交编号
	Language   string `json:"language"`
	Status     int    `json:"status"`
	TimeUsed   int64  `json:"time"`
	MemoryUsed int64  `json:"memory"`
	CreatedAt  string `json:"createdAt"`
}

type SubmissionCaseResult struct {
	ID         int   `json:"id"`
	Status     int   `json:"status"`
	TimeUsed   int64 `json:"time"`
	MemoryUsed int64 `json:"memory"`
}

type SubmissionResp struct {
	User        UserDetail             `json:"user"`
	Problem     SubmissionProblem      `json:"problem"`
	Submission  SubmissionInfo         `json:"submission"`
	CaseResults []SubmissionCaseResult `json:"caseResults"`
}

type DailyAcceptedCountResp struct {
	Date  []string `json:"dates"`
	Count []int    `json:"counts"`
}
