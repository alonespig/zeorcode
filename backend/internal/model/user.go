package model

import "time"

// 用户状态
const (
	UserStatusNormal = 0 // 正常
	UserStatusBanned = 1 // 封禁
)

// 用户角色
const (
	RoleNormal = 0 // 普通用户
	RoleAdmin  = 1 // 管理员
)

type User struct {
	ID int64 `gorm:"primaryKey"`
	// UID 对外用户号（8 位随机数，注册时生成）。公开接口/URL 一律用它，隐藏自增主键防枚举；
	// 内部 FK（提交/比赛/rating 等）与 JWT 仍用 ID。0 表示尚未回填（迁移前的老数据）。
	UID       int64   `gorm:"uniqueIndex"`
	Role      int     `gorm:"default:0"` // 0 普通用户 / 1 管理员（RoleNormal / RoleAdmin）
	Username  string  `gorm:"unique;not null"`
	StudentNo *string `gorm:"type:varchar(32);uniqueIndex"` // 学号；教师/管理员可为空
	RealName  string  `gorm:"type:varchar(64);not null"`    // 校内真实姓名，由注册或管理员维护
	Password  string  `gorm:"not null"`
	// Email 为空时保存为 NULL，使唯一索引既能约束已绑定邮箱，又允许多个未绑定邮箱的账号。
	Email     *string `gorm:"type:varchar(191);uniqueIndex"`
	Gender    int     `gorm:"gender;default:1"`
	Signature string  `gorm:"type:varchar(25);default:''"`
	School    string  `gorm:"type:varchar(50);default:''"`
	Avatar    string  `gorm:"type:varchar(255);default:''"`
	Rating    int     `gorm:"default:0"`          // 竞赛 rating（CF 风）；0=未参加过 rated 比赛（未定级），结算时按 1200 基线起算
	MaxRating int     `gorm:"default:0"`          // 历史最高 rating；0=未定级
	Status    int     `gorm:"not null;default:0"` // 0 正常 / 1 封禁（UserStatusNormal / UserStatusBanned）
	CreatedAt time.Time
	UpdatedAt time.Time
}
