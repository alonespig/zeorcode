package errcode

type Code int

const (
	Success Code = 200

	// 1xxxx 通用 / 参数 / 请求
	InvalidParams   Code = 10001
	BadRequest      Code = 10002
	TooManyRequests Code = 10003

	// 2xxxx 鉴权
	Unauthorized Code = 20001 // 未登录 / 缺少 token
	TokenExpired Code = 20002 // token 过期
	InvalidToken Code = 20003 // token 非法

	// 3xxxx 用户
	UserNotFound            Code = 30001
	WrongPassword           Code = 30002
	UserExists              Code = 30003
	UserDisabled            Code = 30004
	EmailExists             Code = 30005
	VerificationCodeInvalid Code = 30006
	VerificationCodeExpired Code = 30007
	CaptchaInvalid          Code = 30008

	// 4xxxx 题目
	ProblemNotFound          Code = 40001
	ProblemHidden            Code = 40002
	ContestNotFound          Code = 40003
	ContestNotStart          Code = 40004
	ProblemTestDataMissing   Code = 40005
	ProblemDisplayIDExists   Code = 40006
	ContestFinished          Code = 40007
	ContestAlreadyJoined     Code = 40008
	ContestInviteCodeInvalid Code = 40009
	ContestNotRegistered     Code = 40010
	ContestProblemNotFound   Code = 40011
	ProblemSetNotFound       Code = 40012
	ProblemSetLocked         Code = 40013 // 邀请码题单未解锁
	TeamNotFound             Code = 40014
	HomeworkNotFound         Code = 40015
	HomeworkNotStarted       Code = 40016 // 作业未开始，成员看不到题目

	// 5xxxx 提交 / 评测
	SubmissionFailed    Code = 50001
	JudgingInProgress   Code = 50002
	CompilationError    Code = 50003
	JudgeUnavailable    Code = 50004
	SubmissionNotFound  Code = 50005
	UnsupportedLanguage Code = 50006
	RejudgeTargetEmpty  Code = 50007

	// 6xxxx 权限
	PermissionDenied Code = 60001
	AdminRequired    Code = 60002
	OperationDenied  Code = 60003

	// 7xxxx 讨论
	PostNotFound    Code = 70001
	CommentNotFound Code = 70002

	// 8xxxx 远程 OJ
	RemoteOJUnsupported      Code = 80001
	RemoteProblemUnavailable Code = 80002
	RemoteAccountNotFound    Code = 80003

	// 9xxxx 系统
	UnknownError              Code = 90000
	InternalServerError       Code = 90001
	DatabaseError             Code = 90002
	RedisError                Code = 90003
	FileUploadFailed          Code = 90004
	AgentDisabled             Code = 91001
	AgentConversationNotFound Code = 91002
	AgentConversationBusy     Code = 91003
	AgentInteractionExpired   Code = 91004
	AgentActionConflict       Code = 91005
)
