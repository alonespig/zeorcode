package handler

import (
	"context"
	"strconv"

	"zoj/internal/common/errcode"
	"zoj/internal/common/response"
	"zoj/internal/dto"
	"zoj/internal/middleware"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userSrv    *service.UserService
	submitSrv  *service.SubmissionService
	verifySrv  *service.VerifyService
	captchaSrv *service.CaptchaService
	auth       *middleware.Auth
}

func NewUserController(userSrv *service.UserService,
	submitSrv *service.SubmissionService,
	verifySrv *service.VerifyService,
	captchaSrv *service.CaptchaService,
	auth *middleware.Auth) *UserController {
	return &UserController{
		userSrv:    userSrv,
		submitSrv:  submitSrv,
		verifySrv:  verifySrv,
		captchaSrv: captchaSrv,
		auth:       auth,
	}
}

// GenerateCaptcha GET /api/captcha 生成一次性登录图形验证码。
func (u *UserController) GenerateCaptcha(c *gin.Context) (any, error) {
	return u.captchaSrv.Generate(c.Request.Context())
}

// resolveUserPK 把 URL 里的对外用户号(:id)解析成内部主键；找不到返回用户不存在。
func (u *UserController) resolveUserPK(c *gin.Context) (int64, error) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	return u.userSrv.ResolveUID(c.Request.Context(), uid)
}

// RatingHistory GET /api/user/:id/rating 某用户 rating + 变化历史（公开）
func (u *UserController) RatingHistory(c *gin.Context) (any, error) {
	id, err := u.resolveUserPK(c)
	if err != nil {
		return nil, err
	}
	return u.userSrv.GetRatingHistory(c.Request.Context(), id)
}

// ContestHistory GET /api/user/:id/contest-history 用户参赛记录（公开）
func (u *UserController) ContestHistory(c *gin.Context) (any, error) {
	id, err := u.resolveUserPK(c)
	if err != nil {
		return nil, err
	}
	return u.userSrv.GetContestHistory(c.Request.Context(), id)
}

// RatingRank GET /api/user/rating-rank 按 rating 排名（公开）
func (u *UserController) RatingRank(c *gin.Context) (any, error) {
	var req dto.UserRankReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 15
	}
	return u.userSrv.RatingRankList(c.Request.Context(), req.Page, req.PageSize, req.UserName)
}

func (u *UserController) CreateUser(c *gin.Context) (any, error) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if err := u.userSrv.CreateUser(c.Request.Context(), &req); err != nil {
		return nil, err
	}
	return nil, nil
}

// SendVerifyCode POST /api/verify-code 发送邮箱验证码（注册/找回密码/换绑）
func (u *UserController) SendVerifyCode(c *gin.Context) (any, error) {
	var req dto.SendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if err := u.verifySrv.SendCode(c.Request.Context(), req.Scene, req.Email); err != nil {
		return nil, err
	}
	return nil, nil
}

// ResetPassword POST /api/reset-password 通过邮箱验证码重置密码
func (u *UserController) ResetPassword(c *gin.Context) (any, error) {
	var req dto.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if err := u.userSrv.ResetPassword(c.Request.Context(), req.Email, req.Code, req.Password); err != nil {
		return nil, err
	}
	return nil, nil
}

// ChangePassword PUT /api/user/password 校验当前密码后修改密码，并退出当前登录态。
func (u *UserController) ChangePassword(c *gin.Context) {
	v, _ := c.Get("userID")
	userID, _ := v.(int64)
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.Wrap(err))
		return
	}
	if err := u.userSrv.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	middleware.ClearTokenCookie(c)
	response.SuccessEmpty(c)
}

// BindEmail PUT /api/user/email 换绑/绑定邮箱（需登录）
func (u *UserController) BindEmail(c *gin.Context) (any, error) {
	v, _ := c.Get("userID")
	userID, _ := v.(int64)
	var req dto.BindEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if err := u.userSrv.BindEmail(c.Request.Context(), userID, req.Email, req.Code); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *UserController) Rank(c *gin.Context) (any, error) {
	var req dto.UserRankReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return u.userSrv.UserRankList(c.Request.Context(), req.Page, req.PageSize, req.UserName)
}

func (u *UserController) GetUserInfo(c *gin.Context) (any, error) {
	v, _ := c.Get("userID")
	userID, _ := v.(int64)
	return u.userSrv.UserInfo(c.Request.Context(), userID)
}

// UpdateUserInfo PUT /api/user/info 更新当前登录用户的资料
func (u *UserController) UpdateUserInfo(c *gin.Context) (any, error) {
	v, _ := c.Get("userID")
	userID, _ := v.(int64)
	var req dto.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return nil, u.userSrv.UpdateProfile(c.Request.Context(), userID, &req)
}

// Login 登录：验证用户名密码，成功后将 JWT 写入 HttpOnly Cookie
func (u *UserController) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.Wrap(err))
		return
	}
	if err := u.captchaSrv.Verify(c.Request.Context(), req.CaptchaID, req.CaptchaCode); err != nil {
		response.Fail(c, err)
		return
	}
	resp, token, err := u.userSrv.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	middleware.SetTokenCookie(c, token)
	response.Success(c, resp)
}

// Session GET /api/session 同步服务端 Cookie 登录态。
// 路由使用 JWTAuthOptional，因此游客也会得到明确的 authenticated=false。
func (u *UserController) Session(c *gin.Context) (any, error) {
	v, exists := c.Get("userID")
	userID, ok := v.(int64)
	if !exists || !ok || userID <= 0 {
		return &dto.SessionResp{Authenticated: false}, nil
	}

	loginResp, err := u.userSrv.Session(c.Request.Context(), userID)
	if err != nil {
		// token 所指用户已不可用时，让浏览器和 Redis 会话一起失效。
		middleware.ClearTokenCookie(c)
		_ = u.auth.Revoke(context.Background(), userID)
		return nil, err
	}
	return &dto.SessionResp{
		Authenticated: true,
		User:          &loginResp.User,
		Avatar:        loginResp.Avatar,
	}, nil
}

// Logout 退出登录：清除 Cookie + 删除 Redis 中的 token
func (u *UserController) Logout(c *gin.Context) {
	v, _ := c.Get("userID")
	userID, _ := v.(int64)
	middleware.ClearTokenCookie(c)
	if userID > 0 {
		_ = u.auth.Revoke(context.Background(), userID)
	}
	response.SuccessEmpty(c)
}

func (u *UserController) GetUserRecent7DaysPassCount(c *gin.Context) (any, error) {
	userID, err := u.resolveUserPK(c)
	if err != nil {
		return nil, err
	}
	return u.submitSrv.GetUserRecent7DaysPassCount(c.Request.Context(), userID)
}

func (u *UserController) UserProfile(c *gin.Context) (any, error) {
	userID, err := u.resolveUserPK(c)
	if err != nil {
		return nil, err
	}
	return u.userSrv.Profile(c.Request.Context(), userID)
}
