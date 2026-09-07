package dto

import (
	"time"
	"zoj/internal/common/consts"
)

type CreateContestReq struct {
	Name        string                `json:"name" binding:"required"`
	ContestDate string                `json:"contestDate" binding:"required"`
	ContestTime string                `json:"contestTime" binding:"required"`
	Description string                `json:"description"`
	CoverURL    string                `json:"coverUrl"`                         // 封面图 URL（可空）
	Duration    int                   `json:"duration" binding:"required,gt=0"` // 比赛时长，单位：分钟
	Type        int                   `json:"ruleType" binding:"oneof=1 2 3 4"`
	Problems    []ContestProblemInput `json:"problems" binding:"required,dive"`
	// 邀请码：空字符串表示公开比赛
	InviteCode string `json:"inviteCode"`
	Rated      bool   `json:"rated"` // 是否计入 rating
}

// ContestProblemInput 创建比赛时每道题的信息(对外题号 + 气球颜色 + 满分)
type ContestProblemInput struct {
	ProblemID string `json:"problemId" binding:"required"` // 对外题号
	Color     string `json:"color"`
	Score     int    `json:"score"` // OI/IOI 满分；ACM 忽略
}

type UpdateContestReq struct {
	Name        string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	CoverURL    string                `json:"coverUrl"` // 封面图 URL（可空）
	Type        int                   `json:"ruleType"`
	StartTime   int64                 `json:"startTime" binding:"required"`
	EndTime     int64                 `json:"endTime" binding:"required"`
	Problems    []ContestProblemInput `json:"problems"`
	Rated       bool                  `json:"rated"`
}

type EasyInfo struct {
	ID    string `json:"id"` // 对外题号（编辑比赛回显 + 重新提交）
	Name  string `json:"name"`
	Color string `json:"color"`
	Score int    `json:"score"`
}

type EditContestResp struct {
	Name        string     `json:"title"`
	Description string     `json:"description"`
	CoverURL    string     `json:"coverUrl"` // 封面图 URL（编辑回显）
	Type        int        `json:"ruleType"`
	StartTime   int64      `json:"startTime"`
	EndTime     int64      `json:"endTime"`
	ProblemList []EasyInfo `json:"problems"`
	Rated       bool       `json:"rated"`
}

type CreateContestResp struct {
	ID int64 `json:"id"`
}

type ContestListReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
type ContestItem struct {
	ID             int64                `json:"id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	CoverURL       string               `json:"coverUrl"` // 封面图 URL（空则前端回退默认图）
	StartTime      time.Time            `json:"startTime"`
	EndTime        time.Time            `json:"endTime"`
	Type           string               `json:"type"`
	Status         consts.ContestStatus `json:"status"`
	IsRegistered   bool                 `json:"isRegistered"`
	Rated          bool                 `json:"rated"`
	NeedInviteCode bool                 `json:"needInviteCode"` // 是否需要邀请码（不返回真实 code）
	Duration       int                  `json:"duration"`
	Participants   int                  `json:"participants"`
}

// ContestListForm 比赛列表查询：分页 + 搜索/筛选
type ContestListForm struct {
	PageForm
	Keyword string `form:"keyword"`
	Type    int    `form:"type"`   // 0=全部，1 ACM / 2 OI / 3 IOI
	Status  *int   `form:"status"` // nil=全部，0 未开始 / 1 进行中 / 2 已结束
}

type ContestListResp struct {
	Total int64         `json:"total"`
	List  []ContestItem `json:"list"`
}

type JoinContestReq struct {
	ContestID  int64  `json:"id" binding:"required"`
	InviteCode string `json:"inviteCode"`
}

type ContestProblemItem struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type ContestProblem struct {
	ProblemID     string `json:"problemID"` // 对外题号
	Name          string `json:"name"`
	Label         string `json:"label"`
	Color         string `json:"color"`
	Status        int    `json:"status"`
	TimeLimit     int    `json:"timeLimit"`
	MemoryLimit   int    `json:"memoryLimit"`
	TotalCount    int64  `json:"totalCount"`
	AcceptedCount int64  `json:"acceptedCount"`
}

type ContestSubmitStatus struct {
	Label  string `json:"label"`
	Status int    `json:"status"`
}

type ContestProblemListResp struct {
	ProblemList []ContestProblem `json:"problemList"`
}

type ContestProblemResp struct {
	Name         string          `json:"name"`
	TimeLimit    int             `json:"timeLimitMs"`
	MemoryLimit  int             `json:"memoryLimitMb"`
	Description  string          `json:"description"`
	InputFormat  string          `json:"inputFormat"`
	OutputFormat string          `json:"outputFormat"`
	Hint         string          `json:"hint"`
	Samples      []ProblemSample `json:"samples"`
	HasTestData  bool            `json:"hasTestData"`
}

type ContestSubmitResp struct {
	ProblemList []ContestProblemItem `json:"problemList"`
}

type ContestDetailResp struct {
	Name           string `json:"name"`
	StartTime      int64  `json:"startTime"`
	EndTime        int64  `json:"endTime"`
	IsRegistered   bool   `json:"isRegistered"`
	NeedInviteCode bool   `json:"needInviteCode"`
	Rule           string `json:"rule"` // ACM / OI / IOI
	Rated          bool   `json:"rated"`
	Settled        bool   `json:"settled"`
}

type ContestDesc struct {
	Name        string    `json:"name"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Description string    `json:"description"`
}

type ContestSubmitReq struct {
	ContestID int64  `json:"contestId" binding:"required"`
	Label     string `json:"label" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Language  int    `json:"language" binding:"required"`
}

type ContestSubmissionItem struct {
	ID          int64  `json:"id"`        // 对外提交编号
	ProblemID   string `json:"problemID"` // 对外题号
	ProblemName string `json:"problemName"`
	UserID      int64  `json:"userID"`
	UserName    string `json:"userName"`
	Rating      int    `json:"rating"`
	Language    string `json:"language"`
	Result      int    `json:"result"`
	TimeUsed    int64  `json:"timeUsed"`
	MemoryUsed  int64  `json:"memoryUsed"`
	CreatedAt   string `json:"createdAt"`
}

type ContestSubmissionListResp struct {
	Total       int64                   `json:"total"`
	Submissions []ContestSubmissionItem `json:"list"`
}

// ContestSubmissionQuery 比赛提交列表筛选（管理员全场视图用；普通用户忽略，只看自己）
type ContestSubmissionQuery struct {
	Username  string `form:"username"`
	ProblemID string `form:"problemID"` // 对外题号
	Status    *int   `form:"status"`
}

type ContestProblemStatus struct {
	Label      string  `json:"label"`
	Status     int     `json:"status"`
	AcTime     *string `json:"acTime,omitempty"`
	Tries      int     `json:"tries"`
	FirstBlood int     `json:"first,omitempty"`
	Score      *int    `json:"score,omitempty"` // OI/IOI 该题得分；ACM 为 nil
}

type User struct {
	ID     int64  `json:"id"`
	Name   string `json:"nickname"`
	School string `json:"school"`
	Avatar string `json:"avatar"`
	Rating int    `json:"rating"`
}

type ContestRankItem struct {
	Rank        int                    `json:"rank"`
	User        User                   `json:"user"`
	PassCount   int64                  `json:"acCount"`
	Penalty     int                    `json:"penalty"`
	TotalScore  int                    `json:"totalScore"` // OI/IOI 总分
	Problems    []ContestProblemStatus `json:"problems"`
	IsSelf      bool                   `json:"isSelf"`                // 是否当前登录用户(前端高亮自己那一行)
	RatingDelta *int                   `json:"ratingDelta,omitempty"` // 结算后本场 rating 变化
	NewRating   *int                   `json:"newRating,omitempty"`   // 结算后新 rating
}

// MyContestRankResp 当前用户在某场比赛的名次
type MyContestRankResp struct {
	Found      bool   `json:"found"` // 是否参与了该比赛
	Rank       int    `json:"rank"`
	PassCount  int64  `json:"acCount"`
	Penalty    int    `json:"penalty"`
	TotalScore int    `json:"totalScore"` // OI/IOI 总分
	Rule       string `json:"rule"`
}

type Contest struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Rule   string `json:"rule"`             // ACM / OI / IOI，前端据此选榜单组件
	Frozen bool   `json:"frozen,omitempty"` // OI 赛中封榜
}

type ProblemLabel struct {
	Label    string `json:"label"`
	Color    string `json:"color"`
	MaxScore int    `json:"maxScore,omitempty"` // OI/IOI 该题满分
}

type ContestRankResp struct {
	Contest  Contest           `json:"contest"`
	RankList []ContestRankItem `json:"rankList"`
	Problems []ProblemLabel    `json:"problems"`
}

type PageForm struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}
