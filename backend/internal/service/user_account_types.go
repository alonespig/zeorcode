package service

// ===== 账户与登录态 =====

// CreateUserParams 注册新用户的输入。
type CreateUserParams struct {
	Username  string
	StudentNo string
	RealName  string
	Password  string
	Email     string
	Code      string
	Avatar    string
}

// UpdateProfileParams 更新个人资料的输入。
type UpdateProfileParams struct {
	Signature string
	School    string
	Gender    int
	Avatar    string
}

// AuthenticatedUser 登录态中的用户信息。
type AuthenticatedUser struct {
	ID       int64
	Username string
	Role     int
}

// LoginResult 登录 / 会话结果。
type LoginResult struct {
	User   AuthenticatedUser
	Avatar string
}

// CaptchaChallenge 登录验证码。
type CaptchaChallenge struct {
	ID    string
	Image string
}
