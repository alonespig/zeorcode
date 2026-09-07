package dto

// ===== 作业 =====

// HomeworkItemResp 作业列表项
type HomeworkItemResp struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Status       int    `json:"status"` // 0 未开始 / 1 进行中 / 2 已截止
	ProblemCount int    `json:"problemCount"`
	// SolvedCount 当前用户在本作业内已通过的题数；未登录/非成员为 nil
	SolvedCount *int   `json:"solvedCount,omitempty"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Creator     string `json:"creator"`
	CanEdit     bool   `json:"canEdit"`   // 方案 B：仅布置者本人（站点超管例外）
	CanDelete   bool   `json:"canDelete"` // 布置者、团队所有者或站点超管
}

type HomeworkListResp struct {
	Total int                `json:"total"`
	List  []HomeworkItemResp `json:"list"`
}

// HomeworkProblemResp 作业内的单题。Status 语义同题库列表；
// MyScore 是 IOI 计分下该题在时间窗内的最高得分。
type HomeworkProblemResp struct {
	ID            string    `json:"id"` // 对外题号
	Name          string    `json:"name"`
	Status        *int      `json:"status,omitempty"`
	Difficulty    int       `json:"difficulty"`
	Tags          []TagItem `json:"tags"`
	MyScore       *int      `json:"myScore,omitempty"`
	AcceptedCount int       `json:"acceptedCount"`
	SubmitCount   int       `json:"submitCount"`
}

// HomeworkDetailResp 作业详情。
// 未开始时普通成员看不到题目列表（防提前刷题），Problems 为空且 Locked=true；
// 布置者和团队管理员不受限。
type HomeworkDetailResp struct {
	ID            int64                 `json:"id"`
	TeamID        int64                 `json:"teamId"`
	TeamName      string                `json:"teamName"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Status        int                   `json:"status"`
	StartTime     string                `json:"startTime"`
	EndTime       string                `json:"endTime"`
	Creator       string                `json:"creator"`
	CreatorAvatar string                `json:"creatorAvatar"`
	ProblemCount  int                   `json:"problemCount"`
	SolvedCount   *int                  `json:"solvedCount,omitempty"`
	MyScore       *int                  `json:"myScore,omitempty"`
	TotalScore    int                   `json:"totalScore"` // 满分合计 = 题数 × 100
	Locked        bool                  `json:"locked"`     // 未开始且无权提前查看
	CanEdit       bool                  `json:"canEdit"`
	CanSeeAll     bool                  `json:"canSeeAll"` // 能否看全部成员的提交
	Problems      []HomeworkProblemResp `json:"problems"`
}

// SaveHomeworkReq 布置/编辑作业。Problems 的数组顺序即展示顺序。
type SaveHomeworkReq struct {
	Title       string   `json:"title" binding:"required,max=100"`
	Description string   `json:"description"`
	StartTime   string   `json:"startTime" binding:"required"` // "2006-01-02 15:04:05"
	EndTime     string   `json:"endTime" binding:"required"`
	Problems    []string `json:"problems"` // 对外题号列表，顺序敏感
}

// HomeworkSubmitReq 作业内提交
type HomeworkSubmitReq struct {
	ProblemID string `json:"problemID" binding:"required"` // 对外题号
	Code      string `json:"code" binding:"required"`
	Language  int    `json:"language" binding:"required"`
}

// ===== 排行榜 =====

// RankCell 排行榜里某人某题的成绩
type RankCell struct {
	ProblemID string `json:"problemId"` // 对外题号
	Score     int    `json:"score"`
	Solved    bool   `json:"solved"` // 是否拿到满分
}

// HomeworkRankRow 排行榜一行
type HomeworkRankRow struct {
	Rank        int        `json:"rank"`
	UID         int64      `json:"uid"`
	Username    string     `json:"username"`
	StudentNo   string     `json:"studentNo"`
	RealName    string     `json:"realName"`
	Avatar      string     `json:"avatar"`
	TotalScore  int        `json:"totalScore"`
	SolvedCount int        `json:"solvedCount"`
	Cells       []RankCell `json:"cells"`
}

// HomeworkRankResp 排行榜。ProblemIDs 是列头顺序（与作业题目顺序一致）。
type HomeworkRankResp struct {
	StartTime  string            `json:"startTime"`
	EndTime    string            `json:"endTime"`
	ProblemIDs []string          `json:"problemIds"`
	TotalScore int               `json:"totalScore"`
	List       []HomeworkRankRow `json:"list"`
}

// ===== 作业提交列表 =====

// HomeworkSubmissionQuery 作业提交列表筛选。
// UID 只有能看全部提交的人（团队管理员及以上）才生效。
type HomeworkSubmissionQuery struct {
	Page      int    `form:"page" binding:"required"`
	PageSize  int    `form:"pageSize" binding:"required"`
	UID       int64  `form:"uid"`
	ProblemID string `form:"problemId"`
	Status    *int   `form:"status"`
}

type HomeworkSubmissionItem struct {
	ID          int64  `json:"id"` // 对外提交编号
	UID         int64  `json:"uid"`
	Username    string `json:"username"`
	StudentNo   string `json:"studentNo"`
	RealName    string `json:"realName"`
	Avatar      string `json:"avatar"`
	ProblemID   string `json:"problemId"` // 对外题号
	ProblemName string `json:"problemName"`
	Status      int    `json:"status"`
	Score       int    `json:"score"`
	Language    string `json:"language"`
	TimeUsed    int64  `json:"timeUsed"`
	MemoryUsed  int64  `json:"memoryUsed"`
	InWindow    bool   `json:"inWindow"` // 是否落在计分时间窗内（窗外不计入排行榜）
	CreatedAt   string `json:"createdAt"`
}

type HomeworkSubmissionListResp struct {
	Total     int                      `json:"total"`
	CanSeeAll bool                     `json:"canSeeAll"`
	List      []HomeworkSubmissionItem `json:"list"`
}

// HomeworkSubmissionDetailResp 作业提交弹窗详情。
// 源代码只通过作业专用接口返回，权限为提交本人或团队管理者。
type HomeworkSubmissionDetailResp struct {
	ID          int64                  `json:"id"` // 对外提交编号
	UID         int64                  `json:"uid"`
	Username    string                 `json:"username"`
	StudentNo   string                 `json:"studentNo"`
	RealName    string                 `json:"realName"`
	Avatar      string                 `json:"avatar"`
	ProblemID   string                 `json:"problemId"`
	ProblemName string                 `json:"problemName"`
	OJ          string                 `json:"oj"`
	Status      int                    `json:"status"`
	Score       int                    `json:"score"`
	Language    string                 `json:"language"`
	Code        string                 `json:"code"`
	TimeUsed    int64                  `json:"timeUsed"`
	MemoryUsed  int64                  `json:"memoryUsed"`
	CaseResults []SubmissionCaseResult `json:"caseResults"`
	InWindow    bool                   `json:"inWindow"`
	CreatedAt   string                 `json:"createdAt"`
}
