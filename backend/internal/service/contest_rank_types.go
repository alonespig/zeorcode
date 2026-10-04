package service

// 排名类型上的 JSON tag 与旧 dto 保持一致，确保已落缓存的排行榜 JSON 仍能反序列化。

// ContestRankInfo 排行榜里的比赛信息。
type ContestRankInfo struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Rule   string `json:"rule"`
	Frozen bool   `json:"frozen,omitempty"`
}

// ContestProblemLabel 排行榜列头（题目 + 气球颜色 + 满分）。
type ContestProblemLabel struct {
	Label    string `json:"label"`
	Color    string `json:"color"`
	MaxScore int    `json:"maxScore,omitempty"`
}

// ContestRankUser 排行榜里的用户信息。
type ContestRankUser struct {
	ID     int64  `json:"id"`
	Name   string `json:"nickname"`
	School string `json:"school"`
	Avatar string `json:"avatar"`
	Rating int    `json:"rating"`
}

// ContestProblemStatus 排行榜里某人在某题的成绩。
type ContestProblemStatus struct {
	Label      string  `json:"label"`
	Status     int     `json:"status"`
	AcTime     *string `json:"acTime,omitempty"`
	Tries      int     `json:"tries"`
	FirstBlood int     `json:"first,omitempty"`
	Score      *int    `json:"score,omitempty"`
}

// ContestRankItem 排行榜一行。
type ContestRankItem struct {
	Rank        int                    `json:"rank"`
	User        ContestRankUser        `json:"user"`
	PassCount   int64                  `json:"acCount"`
	Penalty     int                    `json:"penalty"`
	TotalScore  int                    `json:"totalScore"`
	Problems    []ContestProblemStatus `json:"problems"`
	IsSelf      bool                   `json:"isSelf"`
	RatingDelta *int                   `json:"ratingDelta,omitempty"`
	NewRating   *int                   `json:"newRating,omitempty"`
}

// ContestRank 排行榜结果。
type ContestRank struct {
	Contest  ContestRankInfo       `json:"contest"`
	RankList []ContestRankItem     `json:"rankList"`
	Problems []ContestProblemLabel `json:"problems"`
}

// MyContestRank 当前用户在某场比赛的名次。
type MyContestRank struct {
	Found      bool   `json:"found"`
	Rank       int    `json:"rank"`
	PassCount  int64  `json:"acCount"`
	Penalty    int    `json:"penalty"`
	TotalScore int    `json:"totalScore"`
	Rule       string `json:"rule"`
}
