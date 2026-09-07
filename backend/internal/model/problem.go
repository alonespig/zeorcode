package model

import "time"

type Problem struct {
	ID int64 `gorm:"primaryKey"`
	// DisplayID 对外题号（管理员创建时指定的字符串，如 "A"/"1001"/"P1234"）。
	// 公开接口/URL 一律用它，隐藏自增主键防枚举；内部 FK（提交/比赛/标签等）仍用 ID。
	DisplayID    string `gorm:"size:32;index"`
	Name         string
	Difficulty   int
	TimeLimit    int
	MemoryLimit  int
	Description  string
	InputFormat  string
	OutputFormat string
	Hint         string
	// 空字符串=本地题；非空=远程题(如 "LibreOJ")，配合 RemoteProblemID 标记
	OJ              string `gorm:"size:20;not null;default:''"`
	RemoteProblemID string `gorm:"size:50;not null;default:''"`
	// 隐藏题：仅管理员可见/可提交；普通用户列表看不到、直接访问 404、提交被拒、提交列表也过滤
	Hidden    bool      `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type ProblemSamples struct {
	ID        int64 `gorm:"primaryKey"`
	ProblemID int64
	Input     string
	Output    string
	Explain   string
}

func (problem *Problem) TableName() string {
	return "problems"
}

type Tag struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(64);not null;uniqueIndex"`
}

type ProblemTag struct {
	ProblemID int64 `gorm:"index:idx_problem_tag,priority:1"`
	TagID     int64 `gorm:"index:idx_problem_tag,priority:2"`
}

// 用户题目状态（UserProblem.Status）。
// 注意语义不是简单的 0/1/2：已通过时写 judge.Accepted(=1)，
// 从未通过时写的是首次提交的判题结果码（2=WA、3=TLE… 见 pkg/judge/status.go），
// 所以「交过但没过」不是单一取值。判断时用 != 0 && != UserProblemSolved，别写 == 2。
const (
	UserProblemUntried = 0 // 无记录 / 未尝试
	UserProblemSolved  = 1 // 已通过，与 judge.Accepted 同值
)

type UserProblem struct {
	ID          int64     `gorm:"primaryKey"`
	UserID      int64     `gorm:"index:idx_user_problem;index:idx_user_problem_status,priority:1"`
	ProblemID   int64     `gorm:"index:idx_user_problem;index:idx_problem_user_status,priority:1"`
	Status      int       `gorm:"default:0;index:idx_user_problem_status,priority:2;index:idx_problem_user_status,priority:2"` // 0: 未尝试, 1: 已通过, 2: 尝试中
	AcCount     int       `gorm:"default:0"`
	SubmitCount int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
