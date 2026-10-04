package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== User ====================

func (h *HttpServer) initUserRouter(r *gin.Engine) {
	user := r.Group("/api")
	{
		user.POST("/user", response.Wrap(h.userController.CreateUser))
		user.GET("/captcha", middleware.RateLimit(h.cache, "captcha", 30, time.Minute), response.Wrap(h.userController.GenerateCaptcha))
		user.POST("/login", middleware.RateLimit(h.cache, "login", 10, time.Minute), h.userController.Login)
		user.GET("/session", middleware.JWTAuthOptional(h.auth), response.Wrap(h.userController.Session))
		// 即使 token 已失效也允许退出，以确保浏览器中的 HttpOnly Cookie 能被清除。
		user.POST("/logout", middleware.JWTAuthOptional(h.auth), h.userController.Logout)
		// 邮箱验证码：发码 + 找回密码按 IP 限流；换绑需登录
		user.POST("/verify-code", middleware.RateLimit(h.cache, "verifycode", 20, time.Hour), response.Wrap(h.userController.SendVerifyCode))
		user.POST("/reset-password", middleware.RateLimit(h.cache, "resetpwd", 20, time.Hour), response.Wrap(h.userController.ResetPassword))
		user.PUT("/user/email", middleware.JWTAuth(h.auth), response.Wrap(h.userController.BindEmail))
		user.PUT("/user/password", middleware.JWTAuth(h.auth), middleware.RateLimit(h.cache, "change-password", 10, time.Hour), h.userController.ChangePassword)
		user.GET("/user/rank", response.Wrap(h.userController.Rank))
		user.GET("/user/rating-rank", response.Wrap(h.userController.RatingRank))
		user.GET("/user/:id/profile", response.Wrap(h.userController.UserProfile))
		user.GET("/user/info", middleware.JWTAuth(h.auth), response.Wrap(h.userController.GetUserInfo))
		user.PUT("/user/info", middleware.JWTAuth(h.auth), response.Wrap(h.userController.UpdateUserInfo))
		user.GET("/user/:id/ac-stats", response.Wrap(h.userController.GetUserRecent7DaysPassCount))
		user.GET("/user/:id/rating", response.Wrap(h.userController.RatingHistory))
		user.GET("/user/:id/contest-history", response.Wrap(h.userController.ContestHistory))
	}
}
