package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Notification(站内通知) ====================

func (h *HttpServer) initNotificationRouter(r *gin.Engine) {
	auth := r.Group("/api", middleware.JWTAuth(h.auth))
	{
		auth.GET("/notifications", response.Wrap(h.notificationController.List))
		auth.GET("/notifications/unread-count", response.Wrap(h.notificationController.UnreadCount))
		auth.POST("/notifications/read", response.Wrap(h.notificationController.MarkRead))
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/notifications", response.Wrap(h.notificationController.BroadcastHistory))
		admin.POST("/notifications", response.Wrap(h.notificationController.Broadcast))
	}
}
