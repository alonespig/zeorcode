package errcode

import "fmt"

// AppErr 业务错误：code + 给前端看的 msg，cause 不导出、不参与 JSON
type AppErr struct {
	Code  Code   `json:"code"`
	Msg   string `json:"msg"`
	cause error  // 底层错误，仅服务端日志可见
}

// Error 实现 error 接口：[code] msg: cause
func (e *AppErr) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Msg, e.cause.Error())
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// Unwrap 支持 errors.Is / errors.As 穿透
func (e *AppErr) Unwrap() error { return e.cause }

// New 直接构造一个 AppErr（无底层 cause）
func New(code Code, msg string) *AppErr {
	return &AppErr{Code: code, Msg: msg}
}

// WithMsg 覆盖 msg，原实例不污染
func (e *AppErr) WithMsg(msg string) *AppErr {
	return &AppErr{Code: e.Code, Msg: msg, cause: e.cause}
}

// Wrap 附加底层错误链；可与 errors.Is/As 配合
func (e *AppErr) Wrap(err error) *AppErr {
	return &AppErr{Code: e.Code, Msg: e.Msg, cause: err}
}

// Cause 取底层 error（如果有）
func (e *AppErr) Cause() error { return e.cause }

// ========================
// 通用 / 参数
// ========================

var (
	ErrInvalidParams   = &AppErr{Code: InvalidParams, Msg: "参数错误"}
	ErrBadRequest      = &AppErr{Code: BadRequest, Msg: "请求格式错误"}
	ErrTooManyRequests = &AppErr{Code: TooManyRequests, Msg: "请求过于频繁"}
)

// ========================
// 鉴权
// ========================

var (
	ErrUnauthorized = &AppErr{Code: Unauthorized, Msg: "未登录"}
	ErrTokenExpired = &AppErr{Code: TokenExpired, Msg: "登录已过期"}
	ErrInvalidToken = &AppErr{Code: InvalidToken, Msg: "登录凭证无效"}
)

// ========================
// 用户
// ========================

var (
	ErrUserNotFound            = &AppErr{Code: UserNotFound, Msg: "用户不存在"}
	ErrWrongPassword           = &AppErr{Code: WrongPassword, Msg: "密码错误"}
	ErrUserExists              = &AppErr{Code: UserExists, Msg: "用户已存在"}
	ErrUserDisabled            = &AppErr{Code: UserDisabled, Msg: "用户被禁用"}
	ErrEmailExists             = &AppErr{Code: EmailExists, Msg: "邮箱已被注册"}
	ErrVerificationCodeInvalid = &AppErr{Code: VerificationCodeInvalid, Msg: "验证码错误"}
	ErrVerificationCodeExpired = &AppErr{Code: VerificationCodeExpired, Msg: "验证码已过期，请重新获取"}
	ErrCaptchaInvalid          = &AppErr{Code: CaptchaInvalid, Msg: "图形验证码错误或已过期"}
)

// ========================
// 题目
// ========================

var (
	ErrProblemNotFound          = &AppErr{Code: ProblemNotFound, Msg: "题目不存在"}
	ErrProblemHidden            = &AppErr{Code: ProblemHidden, Msg: "题目未公开"}
	ErrContestNotFound          = &AppErr{Code: ContestNotFound, Msg: "比赛不存在"}
	ErrContestNotStart          = &AppErr{Code: ContestNotStart, Msg: "比赛未开始"}
	ErrProblemTestDataMissing   = &AppErr{Code: ProblemTestDataMissing, Msg: "题目暂无测试数据，暂时无法提交"}
	ErrProblemDisplayIDExists   = &AppErr{Code: ProblemDisplayIDExists, Msg: "题号已被占用"}
	ErrContestFinished          = &AppErr{Code: ContestFinished, Msg: "比赛已结束"}
	ErrContestAlreadyJoined     = &AppErr{Code: ContestAlreadyJoined, Msg: "已报名该比赛"}
	ErrContestInviteCodeInvalid = &AppErr{Code: ContestInviteCodeInvalid, Msg: "邀请码错误"}
	ErrContestNotRegistered     = &AppErr{Code: ContestNotRegistered, Msg: "尚未报名该比赛"}
	ErrContestProblemNotFound   = &AppErr{Code: ContestProblemNotFound, Msg: "比赛题目不存在"}
	ErrProblemSetNotFound       = &AppErr{Code: ProblemSetNotFound, Msg: "题单不存在"}
	ErrProblemSetLocked         = &AppErr{Code: ProblemSetLocked, Msg: "该题单需要邀请码"}
	ErrTeamNotFound             = &AppErr{Code: TeamNotFound, Msg: "团队不存在"}
	ErrHomeworkNotFound         = &AppErr{Code: HomeworkNotFound, Msg: "作业不存在"}
	ErrHomeworkNotStarted       = &AppErr{Code: HomeworkNotStarted, Msg: "作业尚未开始"}
)

// ========================
// 提交 / 评测
// ========================

var (
	ErrSubmissionFailed    = &AppErr{Code: SubmissionFailed, Msg: "提交失败"}
	ErrJudgingInProgress   = &AppErr{Code: JudgingInProgress, Msg: "正在判题"}
	ErrCompilationError    = &AppErr{Code: CompilationError, Msg: "编译错误"}
	ErrJudgeUnavailable    = &AppErr{Code: JudgeUnavailable, Msg: "判题机不可用"}
	ErrSubmissionNotFound  = &AppErr{Code: SubmissionNotFound, Msg: "提交记录不存在"}
	ErrUnsupportedLanguage = &AppErr{Code: UnsupportedLanguage, Msg: "不支持的编程语言"}
	ErrRejudgeTargetEmpty  = &AppErr{Code: RejudgeTargetEmpty, Msg: "没有可重判的提交"}
)

// ========================
// 权限
// ========================

var (
	ErrPermissionDenied = &AppErr{Code: PermissionDenied, Msg: "权限不足"}
	ErrAdminRequired    = &AppErr{Code: AdminRequired, Msg: "需要管理员权限"}
	ErrOperationDenied  = &AppErr{Code: OperationDenied, Msg: "非本人操作"}
)

// ========================
// 讨论
// ========================

var (
	ErrPostNotFound    = &AppErr{Code: PostNotFound, Msg: "文章不存在"}
	ErrCommentNotFound = &AppErr{Code: CommentNotFound, Msg: "评论不存在"}
)

// ========================
// 远程 OJ
// ========================

var (
	ErrRemoteOJUnsupported      = &AppErr{Code: RemoteOJUnsupported, Msg: "不支持的远程 OJ"}
	ErrRemoteProblemUnavailable = &AppErr{Code: RemoteProblemUnavailable, Msg: "远程题目暂时无法获取"}
	ErrRemoteAccountNotFound    = &AppErr{Code: RemoteAccountNotFound, Msg: "远程账号不存在"}
)

// ========================
// 系统
// ========================

var (
	ErrUnknown                   = &AppErr{Code: UnknownError, Msg: "未知错误"}
	ErrInternal                  = &AppErr{Code: InternalServerError, Msg: "服务器内部错误"}
	ErrDatabase                  = &AppErr{Code: DatabaseError, Msg: "数据库错误"}
	ErrRedis                     = &AppErr{Code: RedisError, Msg: "Redis 错误"}
	ErrFileUpload                = &AppErr{Code: FileUploadFailed, Msg: "文件上传失败"}
	ErrAgentDisabled             = &AppErr{Code: AgentDisabled, Msg: "AI 助手当前未启用"}
	ErrAgentConversationNotFound = &AppErr{Code: AgentConversationNotFound, Msg: "对话不存在"}
	ErrAgentConversationBusy     = &AppErr{Code: AgentConversationBusy, Msg: "当前对话正在处理上一条消息"}
	ErrAgentInteractionExpired   = &AppErr{Code: AgentInteractionExpired, Msg: "该交互已经失效，请重新操作"}
	ErrAgentActionConflict       = &AppErr{Code: AgentActionConflict, Msg: "操作内容已经变化，请重新确认"}
)
