package model

import "time"

// AdminAuditLog 记录管理员写操作的结果，不保存请求正文和敏感字段。
type AdminAuditLog struct {
	ID        int64     `gorm:"primaryKey"`
	ActorID   int64     `gorm:"not null;index"`
	Method    string    `gorm:"size:8;not null"`
	Path      string    `gorm:"size:255;not null;index"`
	Target    string    `gorm:"size:255;not null;default:''"`
	ClientIP  string    `gorm:"size:64;not null;default:''"`
	Success   bool      `gorm:"not null;default:false;index"`
	Code      int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}
