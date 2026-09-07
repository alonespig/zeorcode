package repository

import (
	"context"
	"errors"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type PostRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) List(ctx context.Context, category string, problemID, userID int64, keyword string, page, pageSize int, isHot bool, reviewStatus *int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Post{}).Where("status = ?", 0)
	if reviewStatus != nil {
		db = db.Where("review_status = ?", *reviewStatus)
	}
	if category != "all" {
		db = db.Where("category = ?", category)
	}
	if problemID > 0 {
		db = db.Where("problem_id = ?", problemID)
	}
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}
	if isHot {
		db = db.Order("view_count DESC")
	} else {
		db = db.Order("id DESC")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepo) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, 0).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepo) Update(ctx context.Context, id int64, values map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).
		Where("id = ? AND status = ?", id, 0).
		Updates(values).Error
}

func (r *PostRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).
		Where("id = ? AND status = ?", id, 0).
		Update("status", 1).Error
}

func (r *PostRepo) IncrViewCount(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).
		Where("id = ? AND status = ?", id, 0).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

func (r *PostRepo) HasLiked(ctx context.Context, postID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PostLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *PostRepo) ToggleLike(ctx context.Context, postID, userID int64) (bool, error) {
	liked := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var like model.PostLike
		err := tx.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error
		if err == nil {
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			return tx.Model(&model.Post{}).
				Where("id = ? AND status = ?", postID, 0).
				UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", 1)).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(&model.PostLike{PostID: postID, UserID: userID}).Error; err != nil {
			return err
		}
		liked = true
		return tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", postID, 0).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
	})
	return liked, err
}

// ToggleCommentLike 点赞/取消点赞一条评论，返回点赞后状态。
func (r *PostRepo) ToggleCommentLike(ctx context.Context, commentID, userID int64) (bool, error) {
	liked := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var like model.PostCommentLike
		err := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&like).Error
		if err == nil {
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			return tx.Model(&model.PostComment{}).
				Where("id = ? AND status = ?", commentID, 0).
				UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", 1)).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(&model.PostCommentLike{CommentID: commentID, UserID: userID}).Error; err != nil {
			return err
		}
		liked = true
		return tx.Model(&model.PostComment{}).
			Where("id = ? AND status = ?", commentID, 0).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
	})
	return liked, err
}

// LikedCommentIDs 某用户在某帖下点过赞的评论 id 集合（列表 isLiked 用）。
func (r *PostRepo) LikedCommentIDs(ctx context.Context, postID, userID int64) (map[int64]bool, error) {
	set := map[int64]bool{}
	if userID <= 0 {
		return set, nil
	}
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.PostCommentLike{}).
		Joins("JOIN post_comments c ON c.id = post_comment_likes.comment_id").
		Where("c.post_id = ? AND post_comment_likes.user_id = ?", postID, userID).
		Pluck("post_comment_likes.comment_id", &ids).Error
	for _, id := range ids {
		set[id] = true
	}
	return set, err
}

// ListComments 取某帖的全部有效评论（顶层+回复），按 id 升序，由 service 组装成楼中楼。
func (r *PostRepo) ListComments(ctx context.Context, postID int64) ([]model.PostComment, error) {
	var comments []model.PostComment
	err := r.db.WithContext(ctx).Where("post_id = ? AND status = ?", postID, 0).
		Order("id ASC").
		Find(&comments).Error
	return comments, err
}

func (r *PostRepo) CreateComment(ctx context.Context, comment *model.PostComment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", comment.PostID, 0).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	})
}

func (r *PostRepo) GetComment(ctx context.Context, id int64) (*model.PostComment, error) {
	var comment model.PostComment
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, 0).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *PostRepo) DeleteComment(ctx context.Context, id, postID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.PostComment{}).
			Where("id = ? AND status = ?", id, 0).
			Update("status", 1).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", postID, 0).
			UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", 1)).Error
	})
}
