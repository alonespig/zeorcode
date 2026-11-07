package router

import (
	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/http/middleware"
	"github.com/alonespig/zeorcode/backend/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Language ====================

func (h *HttpServer) initLanguageRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/languages", response.Wrap(h.languageController.List))
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		admin.GET("/languages", middleware.RequirePermission(h.permissions, authz.JudgeLanguageRead), response.Wrap(h.languageController.AdminList))
		admin.POST("/languages", middleware.RequirePermission(h.permissions, authz.JudgeLanguageManage), response.Wrap(h.languageController.Create))
		admin.PUT("/languages/:id", middleware.RequirePermission(h.permissions, authz.JudgeLanguageManage), response.Wrap(h.languageController.Update))
		admin.DELETE("/languages/:id", middleware.RequirePermission(h.permissions, authz.JudgeLanguageManage), response.Wrap(h.languageController.Delete))
	}
}
