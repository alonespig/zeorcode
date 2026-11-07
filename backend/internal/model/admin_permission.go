package model

import "time"

// AdminPermission 是一条“某个管理员拥有某个权限”的授权记录。
// 权限点字符串对应 internal/authz 中的常量；超级管理员不写这张表。
type AdminPermission struct {
	UserID     int64     `gorm:"primaryKey;column:user_id"`
	Permission string    `gorm:"primaryKey;size:64"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (AdminPermission) TableName() string {
	return "admin_permissions"
}
