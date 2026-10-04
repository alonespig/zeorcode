package service

// RemoteProblem 远程拉题结果。
type RemoteProblem struct {
	OJ              string
	RemoteProblemID string
	Title           string
	TimeLimit       int
	MemoryLimit     int
	Description     string
	InputFormat     string
	OutputFormat    string
	Samples         []ProblemSample
	Hint            string
	Source          string
}

// RemoteAccount 远程账号（不含明文凭证，仅 HasSecret 标识是否已设置）。
type RemoteAccount struct {
	ID        int64
	OJ        string
	AuthType  string
	Username  string
	HasSecret bool
	Enabled   bool
	Valid     bool
	Busy      bool
	CreatedAt int64
}

// CreateRemoteAccountParams 新建远程账号的输入。
type CreateRemoteAccountParams struct {
	OJ       string
	AuthType string
	Username string
	Secret   string
	Enabled  bool
}

// UpdateRemoteAccountParams 编辑远程账号的输入；Secret 留空表示不修改凭证。
type UpdateRemoteAccountParams struct {
	AuthType string
	Username string
	Secret   string
	Enabled  bool
}
