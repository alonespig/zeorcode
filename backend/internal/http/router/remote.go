package router

import (
	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/http/middleware"
	"github.com/alonespig/zeorcode/backend/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Remote(远程 OJ) ====================

func (h *HttpServer) initRemoteRouter(r *gin.Engine) {
	remoteWrite := middleware.RequirePermission(h.permissions, authz.ProblemWrite)
	remoteAdmin := r.Group("/api/remote", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		remoteAdmin.GET("/oj", remoteWrite, response.Wrap(h.remoteController.ListOJ))
		remoteAdmin.GET("/problem", remoteWrite, response.Wrap(h.remoteController.GetRemoteProblem))
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		admin.GET("/remote-account", middleware.RequirePermission(h.permissions, authz.JudgeRemoteRead), response.Wrap(h.remoteController.ListAccounts))
		admin.POST("/remote-account", middleware.RequirePermission(h.permissions, authz.JudgeRemoteManage), response.Wrap(h.remoteController.CreateAccount))
		admin.PUT("/remote-account/:id", middleware.RequirePermission(h.permissions, authz.JudgeRemoteManage), response.Wrap(h.remoteController.UpdateAccount))
		admin.DELETE("/remote-account/:id", middleware.RequirePermission(h.permissions, authz.JudgeRemoteManage), response.Wrap(h.remoteController.DeleteAccount))
	}
}
