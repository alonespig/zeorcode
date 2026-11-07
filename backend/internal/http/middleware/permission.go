package middleware

import (
	"github.com/alonespig/zeorcode/backend/internal/http/response"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/service"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// RequireSuperAdmin 只允许超级管理员访问，必须挂在 JWTAuth 之后。
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if r, _ := role.(int); r != model.RoleSuperAdmin {
			response.Fail(c, errcode.ErrPermissionDenied.WithMsg("需要超级管理员权限"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermission 校验当前用户是否拥有任意一个给定权限。
// 超级管理员直接放行；普通用户直接拒绝；admin 查询授权表。
func RequirePermission(ps *service.PermissionService, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, _ := c.Get("userID")
		userID, _ := userIDValue.(int64)
		roleValue, _ := c.Get("role")
		role, _ := roleValue.(int)

		allowed, err := ps.HasAny(c.Request.Context(), userID, role, permissions...)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}
		if !allowed {
			response.Fail(c, errcode.ErrPermissionDenied)
			c.Abort()
			return
		}
		c.Next()
	}
}
