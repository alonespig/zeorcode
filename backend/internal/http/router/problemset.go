package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== ProblemSet ====================

func (h *HttpServer) initProblemSetRouter(r *gin.Engine) {
	// 前台：列表和详情都用可选登录 —— 未登录也能看，只是不返回做题进度
	pub := r.Group("/api")
	{
		pub.GET("/problemset", middleware.JWTAuthOptional(h.auth), response.Wrap(h.problemSetController.List))
		pub.GET("/problemset/:id", middleware.JWTAuthOptional(h.auth), response.Wrap(h.problemSetController.Detail))
		// 解锁要记到当前用户名下，必须登录
		pub.POST("/problemset/:id/unlock", middleware.JWTAuth(h.auth),
			middleware.RateLimit(h.cache, "problemset-unlock", 10, time.Minute),
			response.Wrap(h.problemSetController.Unlock))
	}

	// 后台：仅管理员，列表含草稿
	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/problemset", response.Wrap(h.problemSetController.AdminList))
		admin.POST("/problemset", response.Wrap(h.problemSetController.AdminCreate))
		admin.PUT("/problemset/:id", response.Wrap(h.problemSetController.AdminUpdate))
		admin.DELETE("/problemset/:id", response.Wrap(h.problemSetController.AdminDelete))
	}
}
