package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Post(博客 / 讨论 / 题解) ====================

func (h *HttpServer) initPostRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/posts", middleware.JWTAuthOptional(h.auth), response.Wrap(h.postController.ListPosts))
		pub.GET("/posts/:id", middleware.JWTAuthOptional(h.auth), response.Wrap(h.postController.GetPost))
		pub.GET("/posts/:id/comments", middleware.JWTAuthOptional(h.auth), response.Wrap(h.postController.ListComments))
	}

	auth := r.Group("/api", middleware.JWTAuth(h.auth))
	{
		auth.POST("/posts", response.Wrap(h.postController.CreatePost))
		auth.PUT("/posts/:id", response.Wrap(h.postController.UpdatePost))
		auth.DELETE("/posts/:id", response.Wrap(h.postController.DeletePost))
		auth.POST("/posts/:id/like", response.Wrap(h.postController.ToggleLike))
		auth.POST("/posts/:id/comments", response.Wrap(h.postController.CreateComment))
		auth.POST("/comments/:cid/like", response.Wrap(h.postController.ToggleCommentLike))
		auth.DELETE("/comments/:cid", response.Wrap(h.postController.DeleteComment))
	}
}

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
		admin.POST("/notifications", response.Wrap(h.notificationController.Broadcast))
	}
}

// ==================== Upload ====================

func (h *HttpServer) initUploadRouter(r *gin.Engine) {
	// 注册头像允许未登录上传，但独立限流；普通图片上传仍要求登录。
	publicUpload := r.Group("/api/upload")
	{
		publicUpload.POST("/register-avatar",
			middleware.RateLimit(h.cache, "register-avatar", 20, time.Hour),
			response.Wrap(h.uploadController.UploadImage))
	}
	upload := r.Group("/api/upload", middleware.JWTAuth(h.auth))
	{
		upload.POST("/image", response.Wrap(h.uploadController.UploadImage))
	}
}
