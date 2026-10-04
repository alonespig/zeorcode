package service

import "time"

// ===== 查询结果 =====

// UserRankItem 解题榜 / rating 榜一行。
type UserRankItem struct {
	ID          int64
	Index       int
	Username    string
	Avatar      string
	PassCount   int64
	SubmitCount int64
	Signature   string
	Rating      int
	MaxRating   int
}

// UserRankList 解题榜 / rating 榜。
type UserRankList struct {
	Total int64
	List  []UserRankItem
}

// RatingHistoryItem 一次 rating 变化。
type RatingHistoryItem struct {
	ContestID   int64
	ContestName string
	Rank        int
	OldRating   int
	NewRating   int
	Delta       int
	Time        time.Time
}

// RatingHistory 用户 rating 历史。
type RatingHistory struct {
	Rating    int
	MaxRating int
	History   []RatingHistoryItem
}

// ContestHistoryItem 一场比赛记录。
type ContestHistoryItem struct {
	ContestID   int64
	ContestName string
	Type        int
	Time        time.Time
	Status      string
	Rank        int
	Total       int
	OldRating   int
	NewRating   int
	Delta       int
}

// ContestHistory 比赛记录列表。
type ContestHistory struct {
	List []ContestHistoryItem
}

// UserProblem 已解决/未解决的题目简要信息。
type UserProblem struct {
	ID   string
	Name string
}

// UserInfo 个人主页信息。
type UserInfo struct {
	ID          int64
	Name        string
	Avatar      string
	Email       string
	Gender      int
	School      string
	Signature   *string
	CreatedAt   time.Time
	SolveItem   []UserProblem
	UnsolveItem []UserProblem
}

// ProfileUser 个人主页的用户信息。
type ProfileUser struct {
	ID        int64
	Name      string
	Avatar    string
	Gender    int
	School    string
	Signature *string
	CreatedAt time.Time
}

// Profile 个人主页。
type Profile struct {
	User        ProfileUser
	SolveItem   []UserProblem
	UnsolveItem []UserProblem
}
