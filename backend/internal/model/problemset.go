package model

import "time"

// 题单上架状态
const (
	ProblemSetDraft     = 0 // 草稿，前台完全不可见
	ProblemSetPublished = 1 // 已发布
)

// 题单可见性
const (
	ProblemSetPublic     = 0 // 公开，任何人可看题目列表
	ProblemSetInviteOnly = 1 // 需邀请码，解锁后才返回题目列表
)

// ProblemSet 题单：管理员挑一组题配上说明，用户按单刷题并看到整体进度。
type ProblemSet struct {
	ID          int64  `gorm:"primaryKey"`
	PublicID    int64  `gorm:"column:public_id;uniqueIndex;not null"`
	Title       string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"` // Markdown
	// Published 与 Visibility 是两个正交的轴，不要合并成一个枚举：
	// 前者管「编好之前不给看」，后者管「给谁看」。
	Published  int    `gorm:"not null;default:0"`
	Visibility int    `gorm:"not null;default:0"`
	InviteCode string `gorm:"size:32;not null;default:''"`
	CreatedBy  int64  `gorm:"index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ProblemSetProblem 题单题目关联，Sort 决定展示顺序（从 0 递增）。
type ProblemSetProblem struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	ProblemSetID int64 `gorm:"index;uniqueIndex:idx_set_problem,priority:1"`
	ProblemID    int64 `gorm:"uniqueIndex:idx_set_problem,priority:2"`
	Sort         int   `gorm:"not null;default:0"`
}

// ProblemSetTag 题单标签关联，复用题目那套 Tag 表。
type ProblemSetTag struct {
	ProblemSetID int64 `gorm:"index:idx_set_tag,priority:1"`
	TagID        int64 `gorm:"index:idx_set_tag,priority:2"`
}

// ProblemSetUnlock 邀请码题单的解锁记录，输对一次后长期有效，不必每次重输。
type ProblemSetUnlock struct {
	ID           int64 `gorm:"primaryKey"`
	UserID       int64 `gorm:"uniqueIndex:idx_set_unlock,priority:1"`
	ProblemSetID int64 `gorm:"uniqueIndex:idx_set_unlock,priority:2"`
	CreatedAt    time.Time
}
