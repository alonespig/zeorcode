package model

import "time"

const (
	LanguageDisabled = 0
	LanguageEnabled  = 1
)

// Language 管理前台可选的编程语言。具体编译和运行方式仍由 pkg/judge 的安全预设提供。
type Language struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:32;not null;uniqueIndex" json:"name"`
	Status    int       `gorm:"not null;default:1;index:idx_languages_listing,priority:1" json:"status"`
	Sort      int       `gorm:"not null;default:0;index:idx_languages_listing,priority:2" json:"sort"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
