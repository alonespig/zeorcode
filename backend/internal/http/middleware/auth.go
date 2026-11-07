package middleware

import (
	"net/http"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/http/response"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/service"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// JWTAuth 校验登录态，成功后在 gin.Context 注入 userID / role。
func JWTAuth(ts *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := ts.VerifyToken(tokenString)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}
		// 以数据库中的实时角色为准，不信任 JWT 签发时的角色快照。
		role, err := ts.CurrentRole(c.Request.Context(), claims.UserID)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", role)
		c.Next()
	}
}

// JWTAuthOptional 可选登录：有 token 且校验通过才注入身份，否则放行。
func JWTAuthOptional(ts *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := ts.VerifyToken(tokenString)
		if err == nil {
			if role, roleErr := ts.CurrentRole(c.Request.Context(), claims.UserID); roleErr == nil {
				c.Set("userID", claims.UserID)
				c.Set("role", role)
			}
		}

		c.Next()
	}
}

// AdminRequired 必须在 JWTAuth 之后挂，依赖 c.Get("role")；不涉及 session，保持无状态。
// 只表示“至少是管理员”，细粒度权限由 RequirePermission 判断。
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		r, _ := role.(int)
		if r < model.RoleAdmin {
			response.Fail(c, errcode.ErrAdminRequired)
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractToken 从 Cookie 中提取 token
func extractToken(c *gin.Context) string {
	token, err := c.Cookie("token")
	if err != nil {
		return ""
	}
	return token
}

// SetTokenCookie 将 JWT 写入 HttpOnly Cookie
func SetTokenCookie(c *gin.Context, tokenString string, expire time.Duration) {
	expireSeconds := int(expire.Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", tokenString, expireSeconds, "/", "", false, true)
}

// ClearTokenCookie 清除 token cookie
func ClearTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", "", -1, "/", "", false, true)
}
