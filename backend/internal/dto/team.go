package dto

// ===== 团队 =====

// TeamListReq 团队列表查询。Mine=1 只返回我加入的，Visibility 不传时返回全部权限类型。
type TeamListReq struct {
	Page       int    `form:"page" binding:"required"`
	PageSize   int    `form:"pageSize" binding:"required"`
	Keyword    string `form:"q"`
	Mine       int    `form:"mine"`
	Visibility *int   `form:"visibility"`
}

// TeamItemResp 列表项。列表不返回简介，简介只在详情页展示。
type TeamItemResp struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	CoverURL    string `json:"coverUrl"`
	Visibility  int    `json:"visibility"` // 0 公开 / 1 需邀请码
	MemberCount int    `json:"memberCount"`
	Owner       string `json:"owner"`
	OwnerAvatar string `json:"ownerAvatar"`
	// MyRole 当前用户在该团队中的角色；未登录或非成员为 nil，前端据此显示「加入 / 进入」
	MyRole    *int   `json:"myRole,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type TeamListResp struct {
	Total int            `json:"total"`
	List  []TeamItemResp `json:"list"`
}

// TeamDetailResp 团队详情。
// 非公开团队的非成员：Joined=false 且看不到作业/成员/文件，但名称、简介、
// 成员数照常返回，这样别人知道团队存在、也知道它是干什么的。
type TeamDetailResp struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	CoverURL      string `json:"coverUrl"`
	Description   string `json:"description"`
	Visibility    int    `json:"visibility"`
	MemberCount   int    `json:"memberCount"`
	HomeworkCount int    `json:"homeworkCount"`
	Owner         string `json:"owner"`
	OwnerUID      int64  `json:"ownerUid"`
	MyRole        *int   `json:"myRole,omitempty"`
	CanManage     bool   `json:"canManage"`            // 能否改团队信息、管成员、布置作业
	InviteCode    string `json:"inviteCode,omitempty"` // 仅向所有者/管理员返回，用于邀请成员
	CreatedAt     string `json:"createdAt"`
}

// SaveTeamReq 新建/编辑团队
type SaveTeamReq struct {
	Name        string `json:"name" binding:"required,max=60"`
	CoverURL    string `json:"coverUrl" binding:"max=255"`
	Description string `json:"description"`
	Visibility  int    `json:"visibility" binding:"oneof=0 1"`
	InviteCode  string `json:"inviteCode" binding:"max=32"`
}

// JoinTeamReq 加入团队；非公开团队必须带邀请码
type JoinTeamReq struct {
	InviteCode string `json:"inviteCode"`
}

// ===== 成员 =====

// TeamMemberItem 成员项。UID 是对外用户号，不暴露内部主键。
type TeamMemberItem struct {
	UID       int64  `json:"uid"`
	Username  string `json:"username"`
	StudentNo string `json:"studentNo"`
	RealName  string `json:"realName"`
	Avatar    string `json:"avatar"`
	Role      int    `json:"role"` // 0 成员 / 1 管理员 / 2 所有者
	JoinedAt  string `json:"joinedAt"`
}

type TeamMemberListResp struct {
	Total int              `json:"total"`
	List  []TeamMemberItem `json:"list"`
}

// TeamStudentImportResp 学生名单导入结果。CreatedUsers 是本次新建账号数，
// AddedMembers 是本次实际加入团队的人数，SkippedMembers 是原本已在团队的人数。
type TeamStudentImportResp struct {
	Total          int `json:"total"`
	CreatedUsers   int `json:"createdUsers"`
	AddedMembers   int `json:"addedMembers"`
	SkippedMembers int `json:"skippedMembers"`
}

type TeamStudentManualImportReq struct {
	Text string `json:"text" binding:"required"`
}

// SetMemberRoleReq 设置/取消团队管理员（仅所有者可用）
type SetMemberRoleReq struct {
	Role int `json:"role" binding:"oneof=0 1"`
}
