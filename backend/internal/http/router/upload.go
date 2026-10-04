package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

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
