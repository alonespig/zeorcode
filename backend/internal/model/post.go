package model

// 帖子审核状态（review_status）
const (
	PostReviewPending  = 0 // 待审核
	PostReviewApproved = 1 // 通过
	PostReviewRejected = 2 // 拒绝
)

// 注意：时间字段统一用 int64 时间戳（autoCreateTime/autoUpdateTime 会写 Unix 秒），
// 以匹配共用库 oj_db 中已存在的 posts/post_comments 表（BIGINT 时间列）。
type Post struct {
	ID           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int64  `json:"userID" gorm:"index;not null"`
	Category     string `json:"category" gorm:"size:16;not null;index;index:idx_post_category_problem,priority:1"`
	ProblemID    int64  `json:"problemID" gorm:"not null;default:0;index:idx_post_category_problem,priority:2"`
	Title        string `json:"title" gorm:"size:120;not null"`
	Content      string `json:"content" gorm:"type:longtext;not null"`
	Summary      string `json:"summary" gorm:"size:255;not null;default:''"`
	LikeCount    int64  `json:"likeCount" gorm:"not null;default:0"`
	ViewCount    int64  `json:"viewCount" gorm:"not null;default:0"`
	CommentCount int64  `json:"commentCount" gorm:"not null;default:0"`
	Status       int    `json:"status" gorm:"not null;default:0;index"` // 软删除：0 正常 / 1 已删除
	// ReviewStatus 审核状态：0 待审核 / 1 通过 / 2 拒绝（与 Status 独立）
	ReviewStatus int    `json:"reviewStatus" gorm:"not null;default:0;index"`
	RejectReason string `json:"rejectReason" gorm:"size:255;not null;default:''"` // 拒绝理由（拒绝时填）
	CreatedAt    int64  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    int64  `json:"updatedAt" gorm:"autoUpdateTime"`
}

// PostComment 两级评论表。
// RootID=0 为一级评论；否则为所属一级评论的 id（同一楼的回复都挂在它下面）。
// ReplyUserID 为被回复人 id（用来 @）；0=直接回复楼主，不显示 @。
type PostComment struct {
	ID          int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	PostID      int64  `json:"postID" gorm:"not null;index:idx_post_root,priority:1"`
	UserID      int64  `json:"userID" gorm:"index;not null"`
	RootID      int64  `json:"rootID" gorm:"not null;default:0;index:idx_post_root,priority:2"` // 0=一级评论；否则=所属一级评论 id
	ReplyUserID int64  `json:"replyUserID" gorm:"not null;default:0"`                           // 被回复人 id（@）；0=回复楼主
	Content     string `json:"content" gorm:"type:text;not null"`
	LikeCount   int64  `json:"likeCount" gorm:"not null;default:0"`
	Status      int    `json:"status" gorm:"not null;default:0;index"`
	CreatedAt   int64  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   int64  `json:"updatedAt" gorm:"autoUpdateTime"`
}

type PostLike struct {
	ID        int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	PostID    int64 `json:"postID" gorm:"uniqueIndex:idx_post_like_user;not null"`
	UserID    int64 `json:"userID" gorm:"uniqueIndex:idx_post_like_user;not null"`
	CreatedAt int64 `json:"createdAt" gorm:"autoCreateTime"`
}

type PostCommentLike struct {
	ID        int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	CommentID int64 `json:"commentID" gorm:"uniqueIndex:idx_comment_like_user;not null"`
	UserID    int64 `json:"userID" gorm:"uniqueIndex:idx_comment_like_user;not null"`
	CreatedAt int64 `json:"createdAt" gorm:"autoCreateTime"`
}
