package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Admin ====================

func (h *HttpServer) initAdminRouter(r *gin.Engine) {
	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/dashboard", response.Wrap(h.adminController.Dashboard))
		admin.GET("/users", response.Wrap(h.adminController.ListUsers))
		admin.GET("/users/import/template", h.adminController.DownloadUserImportTemplate)
		admin.POST("/users/import", response.Wrap(h.adminController.ImportUsers))
		admin.PUT("/users/:id/status", response.Wrap(h.adminController.SetUserStatus))
		admin.PUT("/users/:id/role", response.Wrap(h.adminController.SetUserRole))
		admin.GET("/judge/status", response.Wrap(h.adminController.JudgeStatus))
	}
}
