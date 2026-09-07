package dto

type CreateUserReq struct {
	Username  string `json:"username" binding:"required,min=2,max=20"`
	StudentNo string `json:"studentNo" binding:"omitempty,max=32"`
	RealName  string `json:"realName" binding:"omitempty,max=64"`
	Password  string `json:"password" binding:"required,min=6,max=72"`
	Email     string `json:"email" binding:"required,email"`
	Code      string `json:"code" binding:"required"` // 邮箱验证码
	Avatar    string `json:"avatar" binding:"omitempty,max=255"`
}

type LoginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId" binding:"required,max=64"`
	CaptchaCode string `json:"captchaCode" binding:"required,len=5,numeric"`
}

type CaptchaResp struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}

// SendCodeReq 发送邮箱验证码：scene = register | reset | bind
type SendCodeReq struct {
	Email string `json:"email" binding:"required,email"`
	Scene string `json:"scene" binding:"required,oneof=register reset bind"`
}

// ResetPasswordReq 通过邮箱验证码重置密码
type ResetPasswordReq struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// ChangePasswordReq 登录用户通过当前密码修改密码。
type ChangePasswordReq struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6,max=72"`
}

// BindEmailReq 换绑/绑定邮箱（需登录）
type BindEmailReq struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// UpdateProfileReq 编辑个人资料
type UpdateProfileReq struct {
	Signature string `json:"signature" binding:"max=25"`
	School    string `json:"school" binding:"max=50"`
	Gender    int    `json:"gender" binding:"oneof=1 2"`
	Avatar    string `json:"avatar"`
}

type UserInfo struct {
	ID          int64  `json:"id"`
	Index       int    `json:"index"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	PassCount   int64  `json:"count,omitempty"`
	SubmitCount int64  `json:"submitCount,omitempty"`
	Signature   string `json:"signature,omitempty"`
	Role        int    `json:"role,omitempty"`
	Rating      int    `json:"rating"`
	MaxRating   int    `json:"maxRating,omitempty"`
}

type LoginResp struct {
	User   UserInfo `json:"user"`
	Avatar string   `json:"avatar"`
}

// SessionResp 是前端启动时用于同步 HttpOnly Cookie 登录态的响应。
// 未登录时只返回 authenticated=false，避免前端把 localStorage 当作身份依据。
type SessionResp struct {
	Authenticated bool      `json:"authenticated"`
	User          *UserInfo `json:"user,omitempty"`
	Avatar        string    `json:"avatar,omitempty"`
}

type UserRank struct {
	Total    int64      `json:"total"`
	UserRank []UserInfo `json:"list"`
}

type RatingHistoryItem struct {
	ContestID   int64  `json:"contestId"`
	ContestName string `json:"contestName"`
	Rank        int    `json:"rank"`
	OldRating   int    `json:"oldRating"`
	NewRating   int    `json:"newRating"`
	Delta       int    `json:"delta"`
	Time        int64  `json:"time"`
}

type RatingHistoryResp struct {
	Rating    int                 `json:"rating"`
	MaxRating int                 `json:"maxRating"`
	History   []RatingHistoryItem `json:"history"`
}

// ContestHistoryItem 个人页「比赛记录」一行
type ContestHistoryItem struct {
	ContestID   int64  `json:"contestId"`
	ContestName string `json:"contestName"`
	Type        int    `json:"type"`
	Time        int64  `json:"time"`   // 比赛开始时间（秒）
	Status      string `json:"status"` // settled | calculating | running | upcoming | unrated
	Rank        int    `json:"rank"`
	Total       int    `json:"total"` // 名次分母（该场结算人数）
	OldRating   int    `json:"oldRating"`
	NewRating   int    `json:"newRating"`
	Delta       int    `json:"delta"`
}

type ContestHistoryResp struct {
	List []ContestHistoryItem `json:"list"`
}

type ProblemSimple struct {
	ID   string `json:"id"` // 对外题号
	Name string `json:"name"`
}

type UserInfoResp struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Avatar      string          `json:"avatar"`
	Email       string          `json:"email,omitempty"`
	Gender      int             `json:"gender,omitempty"`
	School      string          `json:"school,omitempty"`
	Signature   *string         `json:"signature,omitempty"`
	CreatedAt   int64           `json:"createdAt"`
	SolveItem   []ProblemSimple `json:"solveItem,omitempty"`
	UnsolveItem []ProblemSimple `json:"unsolveItem,omitempty"`
}

type UserProfile struct {
	User        UserDetail      `json:"user"`
	SolveItem   []ProblemSimple `json:"solveItem"`
	UnsolveItem []ProblemSimple `json:"unsolveItem"`
}

// ===== 管理员后台 =====

// AdminUserItem 管理员视角下的用户简要信息
type AdminUserItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	StudentNo string `json:"studentNo"`
	RealName  string `json:"realName"`
	Email     string `json:"email"`
	Role      int    `json:"role"`
	Status    int    `json:"status"` // 0 正常 / 1 封禁
	Signature string `json:"signature"`
	CreatedAt int64  `json:"createdAt"`
}

// SetUserStatusReq 管理员封禁/解封用户
type SetUserStatusReq struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// SetUserRoleReq 管理员修改用户角色（0 普通用户 / 1 管理员）
type SetUserRoleReq struct {
	Role int `json:"role" binding:"oneof=0 1"`
}

type AdminUserListResp struct {
	Total int64           `json:"total"`
	List  []AdminUserItem `json:"list"`
}

// BatchUserItem 批量创建时单个用户的录入项
type BatchUserItem struct {
	Username  string `json:"username" binding:"required,min=2,max=20"`
	StudentNo string `json:"studentNo" binding:"omitempty,max=32"`
	RealName  string `json:"realName" binding:"required,max=64"`
	Email     string `json:"email" binding:"omitempty,email"`
	Password  string `json:"password" binding:"required"`
	Role      int    `json:"role" binding:"oneof=0 1"`
}

type BatchCreateUsersReq struct {
	Users []BatchUserItem `json:"users" binding:"required,min=1,dive"`
}

type BatchCreateUsersResp struct {
	Created int `json:"created"`
}

type UserRankReq struct {
	PageForm
	UserName *string `form:"q"`
}
