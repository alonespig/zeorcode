package judge

// 评测状态码：值与数据库 submission.status 字段直接对应，不能随意变更顺序
const (
	Pending             = 0
	Accepted            = 1
	MemoryLimitExceeded = 2
	TimeLimitExceeded   = 3
	RuntimeError        = 4
	WrongAnswer         = 5
	CompileError        = 6
	UnknownError        = 7
	PresentationError   = 8 // 内容对、仅空白排版不同（与前端 JUDGE_STATUS.PRESENTATION_ERROR 对齐）
)
