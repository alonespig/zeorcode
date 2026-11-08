package handler

import (
	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// hasAnyPermission 判断当前请求是否拥有任意一个权限点。
// 支持可选登录：未登录时直接返回 false；超级管理员始终通过。
func hasAnyPermission(c *gin.Context, permissions *service.PermissionService, required ...string) bool {
	if permissions == nil {
		return false
	}
	v, ok := c.Get("userID")
	if !ok {
		return false
	}
	userID, _ := v.(int64)
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(int)
	if role == model.RoleSuperAdmin {
		return true
	}
	if role != model.RoleAdmin {
		return false
	}
	allowed, err := permissions.HasAny(c.Request.Context(), userID, role, required...)
	return err == nil && allowed
}

// isSiteTeamManager 表示可以跨团队管理（查看/编辑任意团队及其作业）。
// 超级管理员直接放行；admin 需要 team.manage 权限；普通用户只能走团队成员角色。
func isSiteTeamManager(c *gin.Context, permissions *service.PermissionService) bool {
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return false
	}
	switch role {
	case model.RoleSuperAdmin:
		return true
	case model.RoleAdmin:
		allowed, err := permissions.HasAny(
			c.Request.Context(), userID, role, authz.TeamManage,
		)
		return err == nil && allowed
	default:
		return false
	}
}
