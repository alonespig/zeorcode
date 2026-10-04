package dto

type ProblemSample struct {
	Input   string `json:"input"`
	Output  string `json:"output"`
	Explain string `json:"explain"`
}

type CreateProblemReq struct {
	DisplayID       string          `json:"displayId" binding:"required"` // 对外题号（管理员指定的字符串，唯一）
	Name            string          `json:"name" binding:"required"`
	Difficulty      int             `json:"difficulty" binding:"required"`
	TimeLimit       int             `json:"timeLimit" binding:"required"`
	MemoryLimit     int             `json:"memoryLimit" binding:"required"`
	Description     string          `json:"description" binding:"required"`
	InputFormat     string          `json:"inputFormat" binding:"required"`
	OutputFormat    string          `json:"outputFormat" binding:"required"`
	Hint            string          `json:"hint"`
	Samples         []ProblemSample `json:"samples" binding:"required"`
	TagsID          []int64         `json:"tagsID"`
	OJ              string          `json:"oj"` // 非空=远程题
	RemoteProblemID string          `json:"remoteProblemId"`
	Hidden          bool            `json:"hidden"` // 隐藏题
}

type CreateProblemResp struct {
	ID string `json:"id"`
}

type ProblemListReq struct {
	Page       int    `form:"page" binding:"required"`
	PageSize   int    `form:"pageSize" binding:"required"`
	Keyword    string `form:"q"`
	Order      string `form:"order"`      // "latest" => 按 id 倒序
	Difficulty *int   `form:"difficulty"` // 1 简单 / 2 中等 / 3 困难
	Tags       string `form:"tags"`       // 逗号分隔标签 id，如 "1,2,3"（AND 交集）
}

type TagsListResp struct {
	Tags []TagItem `form:"tags"`
}

type TagItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type SaveTagReq struct {
	Name string `json:"name" binding:"required,max=64"`
}

type AdminTagItem struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	ProblemCount    int64  `json:"problemCount"`
	ProblemSetCount int64  `json:"problemSetCount"`
}

type ProblemItemResp struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Status        *int      `json:"status,omitempty"` // 0: 未通过, 1: 已通过, nil: 未尝试
	Difficulty    int       `json:"difficulty"`
	Tags          []TagItem `json:"tags"`
	AcceptedCount int       `json:"acceptedCount"`
	SubmitCount   int       `json:"submitCount"`
	CreatedAt     string    `json:"createdAt"`
	Hidden        bool      `json:"hidden"` // 隐藏题（仅管理员能在列表看到）
	OJ            string    `json:"oj"`     // 空=本地题；非空=远程题（如 "LibreOJ"）
}

type ProblemListResp struct {
	Total int               `json:"total"`
	List  []ProblemItemResp `json:"list"`
}

type ProblemDetailResp struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Difficulty      int             `json:"difficulty"`
	TimeLimit       int             `json:"timeLimitMs"`
	MemoryLimit     int             `json:"memoryLimitMb"`
	Description     string          `json:"description"`
	InputFormat     string          `json:"inputFormat"`
	OutputFormat    string          `json:"outputFormat"`
	SubmitCount     int             `json:"submitCount"`
	AcceptedCount   int             `json:"acceptedCount"`
	Hint            string          `json:"hint"`
	Samples         []ProblemSample `json:"samples"`
	Tags            []TagItem       `json:"tags"`
	OJ              string          `json:"oj"`
	RemoteProblemID string          `json:"remoteProblemId"`
	Hidden          bool            `json:"hidden"`
	HasTestData     bool            `json:"hasTestData"`
}

type ProblemModel struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Difficulty   int             `json:"difficulty"`
	TimeLimit    int             `json:"timeLimit"`
	MemoryLimit  int             `json:"memoryLimit"`
	Description  string          `json:"description"`
	InputFormat  string          `json:"inputFormat"`
	OutputFormat string          `json:"outputFormat"`
	Hint         string          `json:"hint"`
	Samples      []ProblemSample `json:"samples"`
	Tags         []TagItem       `json:"tagsID"`
	Hidden       bool            `json:"hidden"`
}

type UpdateProblemReq struct {
	ID           int64           `json:"-"`                            // 内部主键（handler 由 URL 题号解析后填入，不来自请求体）
	DisplayID    string          `json:"displayId" binding:"required"` // 对外题号（可改，唯一）
	Name         string          `json:"name"`
	Difficulty   int             `json:"difficulty"`
	TimeLimit    int             `json:"timeLimit"`
	MemoryLimit  int             `json:"memoryLimit"`
	Description  string          `json:"description"`
	InputFormat  string          `json:"inputFormat"`
	OutputFormat string          `json:"outputFormat"`
	Hint         string          `json:"hint"`
	Samples      []ProblemSample `json:"samples"`
	Tags         []int64         `json:"tagsID"`
	Hidden       bool            `json:"hidden"`
}
