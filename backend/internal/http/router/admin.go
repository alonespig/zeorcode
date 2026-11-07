package router

import (
	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/http/middleware"
	"github.com/alonespig/zeorcode/backend/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Admin ====================

func (h *HttpServer) initAdminRouter(r *gin.Engine) {
	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		admin.GET("/dashboard", middleware.RequirePermission(h.permissions, authz.DashboardRead), response.Wrap(h.adminController.Dashboard))
		admin.GET("/audit-logs", middleware.RequirePermission(h.permissions, authz.AuditRead), response.Wrap(h.adminController.AuditLogs))
		admin.GET("/users", middleware.RequirePermission(h.permissions, authz.UserRead), response.Wrap(h.adminController.ListUsers))
		admin.GET("/users/import/template", middleware.RequirePermission(h.permissions, authz.UserImport), h.adminController.DownloadUserImportTemplate)
		admin.POST("/users/import", middleware.RequirePermission(h.permissions, authz.UserImport), response.Wrap(h.adminController.ImportUsers))
		admin.PUT("/users/:id/status", middleware.RequirePermission(h.permissions, authz.UserManage), response.Wrap(h.adminController.SetUserStatus))
		admin.PUT("/users/:id/role", middleware.RequireSuperAdmin(), response.Wrap(h.adminController.SetUserRole))
		admin.GET("/permissions/catalog", middleware.RequireSuperAdmin(), response.Wrap(h.adminController.PermissionCatalog))
		admin.GET("/users/:id/permissions", middleware.RequireSuperAdmin(), response.Wrap(h.adminController.GetUserPermissions))
		admin.PUT("/users/:id/permissions", middleware.RequireSuperAdmin(), response.Wrap(h.adminController.SetUserPermissions))
		admin.GET("/judge/status", middleware.RequirePermission(h.permissions, authz.JudgeNodeRead), response.Wrap(h.adminController.JudgeStatus))
	}
}
