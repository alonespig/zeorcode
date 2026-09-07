package model

import (
	"time"
	"zoj/internal/common/consts"
)

type Contest struct {
	ID          int64 `gorm:"primaryKey"`
	PublicID    int64 `gorm:"column:public_id;uniqueIndex;not null"`
	Name        string
	Description string
	// 封面图 URL：空字符串表示未设置，前端回退默认 ICPC 图
	CoverURL  string `gorm:"column:cover_url;size:255;not null;default:''"`
	Type      consts.ContestType
	StartTime time.Time
	EndTime   time.Time
	Duration  int // 比赛时长，单位：分钟
	// 邀请码：空字符串表示公开比赛，非空则报名需要校验
	InviteCode string    `gorm:"size:32;not null;default:''"`
	Rated      bool      `gorm:"default:false"` // 是否计入 rating
	Settled    bool      `gorm:"default:false"` // rating 是否已结算（惰性结算的幂等标记）
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// ContestProblem 比赛题目关联表
type ContestProblem struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ContestID int64  `gorm:"index;index:idx_contest_problem_pair,priority:1;index:idx_contest_problem_label,priority:1"`
	ProblemID int64  `gorm:"index;index:idx_contest_problem_pair,priority:2"`
	Label     string `gorm:"size:10;index:idx_contest_problem_label,priority:2"`
	Color     string `gorm:"size:26"`
	Score     int    `gorm:"default:100"` // OI/IOI 每题满分；ACM 不使用
}

// ContestUser 比赛用户关联表
type ContestUser struct {
	ID        int64     `gorm:"primaryKey"`
	ContestID int64     `gorm:"uniqueIndex:idx_contest_user"`
	UserID    int64     `gorm:"uniqueIndex:idx_contest_user"`
	JoinTime  time.Time `gorm:"autoCreateTime"`
}

// RatingChange 一名用户在一场 rated 比赛的 rating 变化（历史 + 结算幂等来源）
type RatingChange struct {
	ID        int64 `gorm:"primaryKey"`
	ContestID int64 `gorm:"uniqueIndex:idx_rc_contest_user,priority:1"`
	UserID    int64 `gorm:"uniqueIndex:idx_rc_contest_user,priority:2;index:idx_rc_user"`
	Rank      int
	OldRating int
	NewRating int
	Delta     int
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type UserContestProblem struct {
	ID         int64 `gorm:"primaryKey"`
	ContestID  int64 `gorm:"uniqueIndex:idx_ucp_lookup,priority:1;index:idx_ucp_contest_problem_status,priority:1"`
	UserID     int64 `gorm:"uniqueIndex:idx_ucp_lookup,priority:2"`
	ProblemID  int64 `gorm:"uniqueIndex:idx_ucp_lookup,priority:3;index:idx_ucp_contest_problem_status,priority:2"`
	Status     int   `gorm:"index:idx_ucp_contest_problem_status,priority:3"`
	AcCount    int
	Score      int        // OI/IOI 该题得分（IOI 取历史最高，OI 取最后一次）
	UnAcCount  int        `gorm:"column:un_ac_count"`
	SubCount   int        `gorm:"column:sub_count"`
	AcTime     *time.Time `gorm:"column:ac_time;default:NULL"`
	CreateTime time.Time  `gorm:"autoCreateTime"`
}

func (u *UserContestProblem) TableName() string {
	return "user_contest_problem"
}
