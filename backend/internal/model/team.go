package model

import "time"

// 团队成员角色
const (
	TeamRoleMember = 0 // 普通成员
	TeamRoleAdmin  = 1 // 团队管理员
	TeamRoleOwner  = 2 // 所有者（创建者）
)

// 团队公开度
const (
	TeamPublic     = 0 // 公开，任何人可直接加入
	TeamInviteOnly = 1 // 非公开，需邀请码
)

// HomeworkProblemFullScore 作业每题满分固定 100，不做逐题配分。
// 与比赛不同：比赛的每题满分存在 ContestProblem.Score，作业不需要这个灵活度。
const HomeworkProblemFullScore = 100

// Team 团队：老师/队长建团队，往里布置作业，成员刷题并按 IOI 计分排名。
type Team struct {
	ID          int64  `gorm:"primaryKey"`
	PublicID    int64  `gorm:"column:public_id;uniqueIndex;not null"`
	Name        string `gorm:"size:60;not null"`
	CoverURL    string `gorm:"column:cover_url;size:255;not null;default:''"`
	Description string `gorm:"type:text"` // Markdown，显示在概览页
	Visibility  int    `gorm:"not null;default:0"`
	InviteCode  string `gorm:"size:32;not null;default:''"`
	OwnerID     int64  `gorm:"index"` // 所有者内部主键
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TeamMember 团队成员。学校身份统一存放在 User，关系表只维护团队内角色。
type TeamMember struct {
	ID       int64 `gorm:"primaryKey"`
	TeamID   int64 `gorm:"uniqueIndex:idx_team_member,priority:1"`
	UserID   int64 `gorm:"uniqueIndex:idx_team_member,priority:2"`
	Role     int   `gorm:"not null;default:0"`
	JoinedAt time.Time
}

// Homework 团队作业：有起止时间窗，窗内提交才计入排行榜。
type Homework struct {
	ID          int64  `gorm:"primaryKey"`
	PublicID    int64  `gorm:"column:public_id;uniqueIndex;not null"`
	TeamID      int64  `gorm:"index:idx_homework_team_start,priority:1"`
	Title       string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"` // Markdown
	StartTime   time.Time
	EndTime     time.Time
	// CreatedBy 决定谁能改：只有本人能改，团队管理员也不行（用户明确选的方案 B）。
	// 删除额外给团队所有者兜底，否则布置人离队后作业没人能清理。
	CreatedBy int64 `gorm:"index"`
	// SourceActionID makes an Agent write idempotent. A retry after the homework
	// transaction commits returns the existing homework instead of creating one again.
	SourceActionID *int64 `gorm:"column:source_action_id;uniqueIndex"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// HomeworkProblem 作业题目关联，Sort 决定展示顺序（从 0 递增）。
// 不带 Score：每题满分统一 HomeworkProblemFullScore。
type HomeworkProblem struct {
	ID         int64 `gorm:"primaryKey"`
	HomeworkID int64 `gorm:"index;uniqueIndex:idx_homework_problem,priority:1"`
	ProblemID  int64 `gorm:"uniqueIndex:idx_homework_problem,priority:2"`
	Sort       int   `gorm:"not null;default:0"`
}

// 作业状态（按当前时间与起止时间算出来的，不落库）
const (
	HomeworkNotStarted = 0
	HomeworkRunning    = 1
	HomeworkEnded      = 2
)

// Status 按给定时刻算作业状态。
func (h *Homework) Status(now time.Time) int {
	switch {
	case now.Before(h.StartTime):
		return HomeworkNotStarted
	case now.After(h.EndTime):
		return HomeworkEnded
	default:
		return HomeworkRunning
	}
}

// InWindow 判断某个时刻是否落在计分时间窗内（闭区间）。
// 截止后仍可提交，但不计入排行榜，靠这个方法过滤。
func (h *Homework) InWindow(t time.Time) bool {
	return !t.Before(h.StartTime) && !t.After(h.EndTime)
}
