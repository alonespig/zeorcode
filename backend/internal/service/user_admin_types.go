package service

import "time"

// ===== 管理员与批量导入 =====

// AdminUser 管理员视角下的用户信息。
type AdminUser struct {
	ID        int64
	Username  string
	StudentNo string
	RealName  string
	Email     string
	Role      int
	Status    int
	Signature string
	CreatedAt time.Time
}

// AdminUserList 管理员用户列表。
type AdminUserList struct {
	Total int64
	List  []AdminUser
}

type AdminUserListParams struct {
	Page     int
	PageSize int
	Keyword  string
	Role     *int
	Status   *int
}

// BatchUserParams 批量创建时单个用户的录入项。
type BatchUserParams struct {
	Username  string
	StudentNo string
	RealName  string
	Email     string
	Password  string
	Role      int
}

// BatchCreateUsersParams 批量创建用户的输入。
type BatchCreateUsersParams struct {
	Users []BatchUserParams
}

// BatchCreateUsersResult 批量创建结果。
type BatchCreateUsersResult struct {
	Created int
}
