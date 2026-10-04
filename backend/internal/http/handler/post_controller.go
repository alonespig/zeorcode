package handler

import (
	"strconv"

	"zoj/internal/http/dto"
	"zoj/internal/service"
	"zoj/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postSrv *service.PostService
}

func NewPostController(postSrv *service.PostService) *PostController {
	return &PostController{postSrv: postSrv}
}

func (p *PostController) ListPosts(c *gin.Context) (any, error) {
	var req dto.PostListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	// 经 JWTAuthOptional 注入：查看自己的帖子时能带出待审/被拒的
	params := toPostListParams(req)
	resp, err := p.postSrv.ListPosts(c.Request.Context(), &params, optionalCurrentUserID(c))
	if err != nil {
		return nil, err
	}
	return toPostListResp(resp), nil
}

// ReviewListPending GET /api/admin/posts/pending 审核队列（超管）
func (p *PostController) ReviewListPending(c *gin.Context) (any, error) {
	var req dto.PageForm
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	resp, err := p.postSrv.AdminListPending(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return toPostListResp(resp), nil
}

// ReviewPost PUT /api/admin/posts/:id/review 通过/拒绝帖子（超管）
func (p *PostController) ReviewPost(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	var req dto.ReviewPostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return nil, p.postSrv.ReviewPost(c.Request.Context(), id, req.Status, req.Reason)
}

func (p *PostController) CreatePost(c *gin.Context) (any, error) {
	var req dto.CreatePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	resp, err := p.postSrv.CreatePost(c.Request.Context(), toCreatePostParams(req), userID, role)
	if err != nil {
		return nil, err
	}
	return toCreatePostResult(resp), nil
}

func (p *PostController) GetPost(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(int)
	resp, err := p.postSrv.GetPost(c.Request.Context(), id, optionalCurrentUserID(c), role)
	if err != nil {
		return nil, err
	}
	return toPostDetailResp(resp), nil
}

func (p *PostController) UpdatePost(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	var req dto.UpdatePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return nil, p.postSrv.UpdatePost(c.Request.Context(), id, toUpdatePostParams(req), userID, role)
}

func (p *PostController) DeletePost(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return nil, p.postSrv.DeletePost(c.Request.Context(), id, userID, role)
}

func (p *PostController) ToggleLike(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	resp, err := p.postSrv.ToggleLike(c.Request.Context(), id, userID)
	if err != nil {
		return nil, err
	}
	return toToggleLikeResult(resp), nil
}

func (p *PostController) ListComments(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	v, _ := c.Get("userID") // 可选：登录了才知道 isLiked
	uid, _ := v.(int64)
	resp, err := p.postSrv.ListComments(c.Request.Context(), id, uid)
	if err != nil {
		return nil, err
	}
	return toPostCommentListResp(resp), nil
}

// ToggleCommentLike POST /api/comments/:cid/like
func (p *PostController) ToggleCommentLike(c *gin.Context) (any, error) {
	cid, err := parseInt64Param(c, "cid")
	if err != nil {
		return nil, err
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	resp, err := p.postSrv.ToggleCommentLike(c.Request.Context(), cid, userID)
	if err != nil {
		return nil, err
	}
	return toToggleLikeResult(resp), nil
}

func (p *PostController) CreateComment(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return nil, err
	}
	var req dto.CreatePostCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return nil, p.postSrv.CreateComment(c.Request.Context(), id, userID, toCreateCommentParams(req))
}

func (p *PostController) DeleteComment(c *gin.Context) (any, error) {
	id, err := parseInt64Param(c, "cid")
	if err != nil {
		return nil, err
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return nil, p.postSrv.DeleteComment(c.Request.Context(), id, userID, role)
}

func parseInt64Param(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, errcode.ErrInvalidParams.Wrap(err)
	}
	return id, nil
}

func optionalCurrentUserID(c *gin.Context) int64 {
	v, ok := c.Get("userID")
	if !ok {
		return 0
	}
	userID, _ := v.(int64)
	return userID
}

func requireCurrentUser(c *gin.Context) (int64, int, error) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, 0, errcode.ErrUnauthorized
	}
	userID, ok := v.(int64)
	if !ok {
		return 0, 0, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(int)
	return userID, role, nil
}
