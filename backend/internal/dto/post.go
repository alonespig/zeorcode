package dto

type PostListReq struct {
	Category  string `form:"category" binding:"oneof=all blog help solution announcement"`
	ProblemID string `form:"problemID"` // 对外题号
	UserID    int64  `form:"userID"` // >0 时只看该用户发布的帖子（个人主页"帖子"tab）
	Keyword   string `form:"keyword"`
	Sort      string `form:"sort"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

type CreatePostReq struct {
	Category  string `json:"category" binding:"required,oneof=blog help solution announcement"`
	ProblemID string `json:"problemID"` // 对外题号
	Title     string `json:"title" binding:"required,max=120"`
	Content   string `json:"content" binding:"required"`
	Summary   string `json:"summary" binding:"max=255"`
}

type UpdatePostReq struct {
	Category  string `json:"category" binding:"required,oneof=blog help solution announcement"`
	ProblemID string `json:"problemID"` // 对外题号
	Title     string `json:"title" binding:"required,max=120"`
	Content   string `json:"content" binding:"required"`
	Summary   string `json:"summary" binding:"max=255"`
}

type PostUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Rating   int    `json:"rating"`
}

type PostItem struct {
	ID           int64          `json:"id"`
	Category     string         `json:"category"`
	ReviewStatus int            `json:"reviewStatus"` // 0 待审核 / 1 通过 / 2 拒绝
	RejectReason string         `json:"rejectReason,omitempty"`
	ProblemID    string         `json:"problemID"` // 对外题号
	Problem      *ProblemSimple `json:"problem,omitempty"`
	Title        string         `json:"title"`
	Summary      string         `json:"summary"`
	User         PostUser       `json:"user"`
	LikeCount    int64          `json:"likeCount"`
	ViewCount    int64          `json:"viewCount"`
	CommentCount int64          `json:"commentCount"`
	CreatedAt    string         `json:"createdAt"`
	UpdatedAt    string         `json:"updatedAt"`
}

type PostListResp struct {
	Total int64      `json:"total"`
	List  []PostItem `json:"list"`
}

type PostDetailResp struct {
	PostItem
	Content string `json:"content"`
	IsLiked bool   `json:"isLiked"`
}

type CreatePostResp struct {
	ID           int64 `json:"id"`
	ReviewStatus int   `json:"reviewStatus"` // 0 待审核(前端提示"等待审核") / 1 直接通过
}

// ReviewPostReq 管理员审核帖子：status = 1 通过 / 2 拒绝；拒绝可带 reason
type ReviewPostReq struct {
	Status int    `json:"status" binding:"oneof=1 2"`
	Reason string `json:"reason"`
}

type TogglePostLikeResp struct {
	Liked bool `json:"liked"`
}

type CreatePostCommentReq struct {
	Content  string `json:"content" binding:"required"`
	ParentID int64  `json:"parentID"` // 0=顶层评论；否则=被回复评论的 id
}

type PostCommentItem struct {
	ID        int64             `json:"id"`
	PostID    int64             `json:"postID"`
	Content   string            `json:"content"`
	User      PostUser          `json:"user"`
	ReplyTo   *PostUser         `json:"replyTo,omitempty"` // 被回复者；nil=直接回复楼主
	LikeCount int64             `json:"likeCount"`
	IsLiked   bool              `json:"isLiked"`
	CreatedAt int64             `json:"createdAt"`
	Replies   []PostCommentItem `json:"replies,omitempty"` // 仅顶层评论带楼内回复
}

type PostCommentListResp struct {
	Total int64             `json:"total"`
	List  []PostCommentItem `json:"list"`
}
