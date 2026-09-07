package model

import "time"

const (
	SubmissionOutboxPending uint8 = iota
	SubmissionOutboxProcessing
	SubmissionOutboxPublished
	SubmissionOutboxJudging
	SubmissionOutboxCompleted
)

type Submission struct {
	ID        int64 `gorm:"primaryKey" json:"-"`
	PublicID  int64 `gorm:"column:public_id;uniqueIndex;not null" json:"id,omitempty"`
	ProblemID int64 `gorm:"index:idx_submission_problem_status,priority:1;index:idx_submission_contest_problem_status_created,priority:2;index:idx_submission_homework_problem_status,priority:2" json:"problemID,omitempty"`
	UserID    int64 `gorm:"index:idx_submission_user_status_created,priority:1;index:idx_submission_contest_user_created,priority:2;index:idx_submission_homework_user,priority:2" json:"userID,omitempty"`
	ContestID int64 `gorm:"index;not null;default:0;index:idx_submission_contest_created,priority:1;index:idx_submission_contest_problem_status_created,priority:1;index:idx_submission_contest_user_created,priority:1" json:"contestID,omitempty"` // 可选，0 表示非比赛提交
	// HomeworkID 团队作业提交，0 表示非作业提交。与 ContestID 对称、互斥：
	// 作业不是比赛，塞进 contest 会污染 rating/封榜/赛制那一整套逻辑，所以单开一列。
	HomeworkID    int64     `gorm:"not null;default:0;index:idx_submission_homework_user,priority:1;index:idx_submission_homework_problem_status,priority:1" json:"homeworkID,omitempty"`
	Code          string    `gorm:"not null" json:"code,omitempty"`
	Language      string    `gorm:"not null" json:"language,omitempty"`
	Status        int       `gorm:"not null;index:idx_submission_problem_status,priority:2;index:idx_submission_contest_problem_status_created,priority:3;index:idx_submission_user_status_created,priority:2;index:idx_submission_homework_problem_status,priority:3" json:"status,omitempty"`
	TimeUsed      int64     `json:"timeUsed,omitempty"`                             // ms（所有测试点的最大值）
	MemoryUsed    int64     `json:"memoryUsed,omitempty"`                           // KB（最大值）
	CompileOutput string    `gorm:"type:mediumtext" json:"compileOutput,omitempty"` // 编译失败时沙箱返回的原始输出
	Score         int       `gorm:"not null;default:0" json:"score,omitempty"`      // OI/IOI 单次提交得分（测试点均分）
	Version       int       `gorm:"not null;default:0" json:"-"`                    // 重判代数（乐观锁）：每次重判 +1；worker 写回按此校验，version 不符则丢弃过时结果
	CreatedAt     time.Time `gorm:"autoCreateTime;index:idx_submission_contest_created,priority:2;index:idx_submission_contest_problem_status_created,priority:4;index:idx_submission_contest_user_created,priority:3;index:idx_submission_user_status_created,priority:3;index:idx_submission_homework_user,priority:3" json:"createdAt"`
}

// SubmissionOutbox 记录需要投递到判题队列的 Submission 版本。
// Submission 与 Outbox 在同一数据库事务中写入，Relay 负责可靠投递。
type SubmissionOutbox struct {
	ID                int64      `gorm:"primaryKey"`
	SubmissionID      int64      `gorm:"not null;uniqueIndex:idx_submission_outbox_version,priority:1"`
	SubmissionVersion int        `gorm:"not null;uniqueIndex:idx_submission_outbox_version,priority:2"`
	Status            uint8      `gorm:"not null;default:0;index:idx_submission_outbox_dispatch,priority:1"`
	Attempts          int        `gorm:"not null;default:0"`
	AvailableAt       time.Time  `gorm:"not null;index:idx_submission_outbox_dispatch,priority:2"`
	LockedBy          string     `gorm:"size:128;not null;default:''"`
	LockedUntil       *time.Time `gorm:"default:null"`
	PublishedAt       *time.Time `gorm:"default:null"`
	LastError         string     `gorm:"size:1024;not null;default:''"`
	CreatedAt         time.Time  `gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime"`
}

type JudgeResult struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	SubmissionID int64  `gorm:"index" json:"submissionID"`
	CaseID       int64  `gorm:"index" json:"case_id"` // 或 TestDataID
	Status       int    `json:"status"`               // AC / WA / TLE / RE ...
	TimeUsed     int64  `json:"timeUsed"`             // ms
	MemoryUsed   int64  `json:"memoryUsed"`           // KB
	OutputDiff   string `json:"outputDiff"`           // 可选：WA 时的 diff / 提示
}
