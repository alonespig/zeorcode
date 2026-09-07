package remoteoj

// ProblemSample 远程题样例
type ProblemSample struct {
	Input   string `json:"input"`
	Output  string `json:"output"`
	Explain string `json:"explain"`
}

// RemoteProblem 从远程 OJ 爬取到的题目
type RemoteProblem struct {
	OJ              string
	RemoteProblemID string
	Title           string
	TimeLimit       int // ms
	MemoryLimit     int // MB
	Description     string
	InputFormat     string
	OutputFormat    string
	Samples         []ProblemSample
	Hint            string
	Source          string // 原题链接
}

// RemoteAccount 远程账号。
// cookie 方案(牛客/AtCoder)用 Cookie；账号密码方案(HDU/CF)用 Username+Password。
type RemoteAccount struct {
	ID       int64
	OJ       string
	Username string
	Password string
	Cookie   string
	Valid    bool
}

// SubmitReq 远程提交请求
type SubmitReq struct {
	RemoteProblemID string
	Language        string // 本系统语言名，如 "cpp"
	Code            string
}

// Verdict 远程判题结果（已映射成本系统状态码）
type Verdict struct {
	Status   int
	TimeMs   int64
	MemoryKB int64
	Msg      string
}
