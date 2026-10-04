package service

import "time"

// ===== 查询 / 命令输入 =====

// PostListParams 帖子列表查询。Page/PageSize 由 normalizePostPage 兜底。
type PostListParams struct {
	Category  string
	ProblemID string
	UserID    int64
	Keyword   string
	Sort      string
	Page      int
	PageSize  int
}

type CreatePostParams struct {
	Category  string
	ProblemID string
	Title     string
	Content   string
	Summary   string
}

type UpdatePostParams struct {
	Category  string
	ProblemID string
	Title     string
	Content   string
	Summary   string
}

type CreateCommentParams struct {
	Content  string
	ParentID int64
}

// ===== 通用值对象 =====

// PostUser 帖子/评论的作者信息。
type PostUser struct {
	ID       int64
	Username string
	Avatar   string
	Rating   int
}

// PostProblem 帖子关联的题目信息。
type PostProblem struct {
	ID   string
	Name string
}

// ===== 结果 =====

type PostItem struct {
	ID           int64
	Category     string
	ReviewStatus int
	RejectReason string
	ProblemID    string
	Problem      *PostProblem
	Title        string
	Summary      string
	User         PostUser
	LikeCount    int64
	ViewCount    int64
	CommentCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PostList struct {
	Total int64
	List  []PostItem
}

type PostDetail struct {
	PostItem
	Content string
	IsLiked bool
}

type CreatePostResult struct {
	ID           int64
	ReviewStatus int
}

type ToggleLikeResult struct {
	Liked bool
}

type PostComment struct {
	ID        int64
	PostID    int64
	Content   string
	User      PostUser
	ReplyTo   *PostUser
	LikeCount int64
	IsLiked   bool
	CreatedAt time.Time
	Replies   []PostComment
}

type PostCommentList struct {
	Total int64
	List  []PostComment
}
