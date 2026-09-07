package model

import "time"

// Notification 站内通知（统一表）：互动(comment/reply/like) + 系统(system) + rating。
// 系统通知靠广播时给每个用户扇出一行实现"按人投递 + 已读"。
type Notification struct {
	ID          int64     `gorm:"primaryKey"`
	RecipientID int64     `gorm:"index:idx_notif_recipient,priority:1"` // 收件人
	IsRead      bool      `gorm:"index:idx_notif_recipient,priority:2;not null;default:false"`
	Type        string    `gorm:"size:16"` // comment / reply / like / system / rating
	ActorID     int64     // 触发者；系统/rating 为 0
	Title       string    `gorm:"size:255"` // 系统通知标题；互动可空
	Content     string    `gorm:"size:255"` // 摘要
	Link        string    `gorm:"size:255"` // 站内相对路径，如 /blog/123
	SourceType  string    `gorm:"size:32"`  // post / comment / contest …
	SourceID    int64     //
	CreatedAt   time.Time `gorm:"autoCreateTime;index"`
}
