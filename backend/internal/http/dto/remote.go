package dto

// RemoteProblemResp 远程拉题返回（字段对齐 CreateProblemReq，前端可直接回填）
type RemoteProblemResp struct {
	OJ              string          `json:"oj"`
	RemoteProblemID string          `json:"remoteProblemId"`
	Title           string          `json:"title"`
	TimeLimit       int             `json:"timeLimit"`
	MemoryLimit     int             `json:"memoryLimit"`
	Description     string          `json:"description"`
	InputFormat     string          `json:"inputFormat"`
	OutputFormat    string          `json:"outputFormat"`
	Samples         []ProblemSample `json:"samples"`
	Hint            string          `json:"hint"`
	Source          string          `json:"source"`
}

// RemoteAccountItem 远程账号列表项（不含明文凭证，仅用 hasSecret 标识是否已设置）
type RemoteAccountItem struct {
	ID        int64  `json:"id"`
	OJ        string `json:"oj"`
	AuthType  string `json:"authType"`
	Username  string `json:"username"`
	HasSecret bool   `json:"hasSecret"`
	Enabled   bool   `json:"enabled"`
	Valid     bool   `json:"valid"`
	Busy      bool   `json:"busy"` // 账号池占用中(正在远程判题)
	CreatedAt int64  `json:"created_at"`
}

// CreateRemoteAccountReq 新建远程账号
type CreateRemoteAccountReq struct {
	OJ       string `json:"oj" binding:"required"`
	AuthType string `json:"authType" binding:"required,oneof=password cookie"`
	Username string `json:"username"`
	Secret   string `json:"secret"` // 密码或 cookie 明文，入库前 AES 加密
	Enabled  bool   `json:"enabled"`
}

// UpdateRemoteAccountReq 编辑远程账号；Secret 留空表示不修改凭证
type UpdateRemoteAccountReq struct {
	AuthType string `json:"authType" binding:"required,oneof=password cookie"`
	Username string `json:"username"`
	Secret   string `json:"secret"`
	Enabled  bool   `json:"enabled"`
}
