package handler

import (
	"zoj/internal/http/dto"
	"zoj/internal/service"
)

const postTimeLayout = "2006-01-02 15:04:05"

func toPostUser(u service.PostUser) dto.PostUser {
	return dto.PostUser{ID: u.ID, Username: u.Username, Avatar: u.Avatar, Rating: u.Rating}
}

func toPostProblem(p *service.PostProblem) *dto.ProblemSimple {
	if p == nil {
		return nil
	}
	return &dto.ProblemSimple{ID: p.ID, Name: p.Name}
}

func toPostItem(item service.PostItem) dto.PostItem {
	return dto.PostItem{
		ID:           item.ID,
		Category:     item.Category,
		ReviewStatus: item.ReviewStatus,
		RejectReason: item.RejectReason,
		ProblemID:    item.ProblemID,
		Problem:      toPostProblem(item.Problem),
		Title:        item.Title,
		Summary:      item.Summary,
		User:         toPostUser(item.User),
		LikeCount:    item.LikeCount,
		ViewCount:    item.ViewCount,
		CommentCount: item.CommentCount,
		CreatedAt:    item.CreatedAt.Format(postTimeLayout),
		UpdatedAt:    item.UpdatedAt.Format(postTimeLayout),
	}
}

func toPostListResp(r *service.PostList) *dto.PostListResp {
	items := make([]dto.PostItem, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, toPostItem(it))
	}
	return &dto.PostListResp{Total: r.Total, List: items}
}

func toPostDetailResp(r *service.PostDetail) *dto.PostDetailResp {
	return &dto.PostDetailResp{
		PostItem: toPostItem(r.PostItem),
		Content:  r.Content,
		IsLiked:  r.IsLiked,
	}
}

func toCreatePostResult(r *service.CreatePostResult) *dto.CreatePostResp {
	return &dto.CreatePostResp{ID: r.ID, ReviewStatus: r.ReviewStatus}
}

func toToggleLikeResult(r *service.ToggleLikeResult) *dto.TogglePostLikeResp {
	return &dto.TogglePostLikeResp{Liked: r.Liked}
}

func toPostComment(c service.PostComment) dto.PostCommentItem {
	var replyTo *dto.PostUser
	if c.ReplyTo != nil {
		u := toPostUser(*c.ReplyTo)
		replyTo = &u
	}
	var replies []dto.PostCommentItem
	if c.Replies != nil {
		replies = make([]dto.PostCommentItem, 0, len(c.Replies))
		for _, reply := range c.Replies {
			replies = append(replies, toPostComment(reply))
		}
	}
	return dto.PostCommentItem{
		ID:        c.ID,
		PostID:    c.PostID,
		Content:   c.Content,
		User:      toPostUser(c.User),
		ReplyTo:   replyTo,
		LikeCount: c.LikeCount,
		IsLiked:   c.IsLiked,
		CreatedAt: c.CreatedAt.Unix(),
		Replies:   replies,
	}
}

func toPostCommentListResp(r *service.PostCommentList) *dto.PostCommentListResp {
	items := make([]dto.PostCommentItem, 0, len(r.List))
	for _, c := range r.List {
		items = append(items, toPostComment(c))
	}
	return &dto.PostCommentListResp{Total: r.Total, List: items}
}

func toPostListParams(req dto.PostListReq) service.PostListParams {
	return service.PostListParams{
		Category:  req.Category,
		ProblemID: req.ProblemID,
		UserID:    req.UserID,
		Keyword:   req.Keyword,
		Sort:      req.Sort,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
}

func toCreatePostParams(req dto.CreatePostReq) service.CreatePostParams {
	return service.CreatePostParams{
		Category:  req.Category,
		ProblemID: req.ProblemID,
		Title:     req.Title,
		Content:   req.Content,
		Summary:   req.Summary,
	}
}

func toUpdatePostParams(req dto.UpdatePostReq) service.UpdatePostParams {
	return service.UpdatePostParams{
		Category:  req.Category,
		ProblemID: req.ProblemID,
		Title:     req.Title,
		Content:   req.Content,
		Summary:   req.Summary,
	}
}

func toCreateCommentParams(req dto.CreatePostCommentReq) service.CreateCommentParams {
	return service.CreateCommentParams{
		Content:  req.Content,
		ParentID: req.ParentID,
	}
}
