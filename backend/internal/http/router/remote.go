package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Remote(远程 OJ) ====================

func (h *HttpServer) initRemoteRouter(r *gin.Engine) {
	remoteAdmin := r.Group("/api/remote", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		remoteAdmin.GET("/oj", response.Wrap(h.remoteController.ListOJ))
		remoteAdmin.GET("/problem", response.Wrap(h.remoteController.GetRemoteProblem))
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		admin.GET("/remote-account", response.Wrap(h.remoteController.ListAccounts))
		admin.POST("/remote-account", response.Wrap(h.remoteController.CreateAccount))
		admin.PUT("/remote-account/:id", response.Wrap(h.remoteController.UpdateAccount))
		admin.DELETE("/remote-account/:id", response.Wrap(h.remoteController.DeleteAccount))
	}
}
