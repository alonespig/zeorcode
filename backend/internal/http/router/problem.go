package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Problem ====================

func (h *HttpServer) initLanguageRouter(r *gin.Engine) {
	pub := r.Group("/api")
	{
		pub.GET("/languages", response.Wrap(h.languageController.List))
	}
}

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

func (h *HttpServer) initProblemRouter(r *gin.Engine) {
	// 公开接口
	pub := r.Group("/api")
	{
		pub.GET("/tags", response.Wrap(h.problemController.GetTagList))
		pub.GET("/problems", middleware.JWTAuthOptional(h.auth), response.Wrap(h.problemController.List))
		pub.GET("/problems/:id", middleware.JWTAuthOptional(h.auth), response.Wrap(h.problemController.GetProblemDetail))
	}

	// 管理员接口
	admin := r.Group("/api", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.POST("/problems", response.Wrap(h.problemController.CreateProblem))
		admin.PUT("/problems/:id", response.Wrap(h.problemController.UpdateProblem))
		admin.GET("/problems/:id/edit", response.Wrap(h.problemController.GetProblemForEdit))
		admin.GET("/problems/:id/testdata", response.Wrap(h.problemController.ListFiles))
		admin.GET("/problems/:id/testdata/content", response.Wrap(h.problemController.GetFileContent))
		admin.POST("/problems/:id/testdata", response.Wrap(h.problemController.UploadFile))
		admin.POST("/problems/:id/testdata/zip", response.Wrap(h.problemController.UploadTestcaseZip))
		admin.DELETE("/problems/:id/testdata", response.Wrap(h.problemController.DeleteFiles))
		admin.GET("/problems/:id/testdata/download", h.problemController.DownloadFiles)
	}

	tagAdmin := r.Group("/api/admin/tags", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		tagAdmin.GET("", response.Wrap(h.problemController.GetAdminTagList))
		tagAdmin.POST("", response.Wrap(h.problemController.CreateTag))
		tagAdmin.PUT("/:id", response.Wrap(h.problemController.UpdateTag))
		tagAdmin.DELETE("/:id", response.Wrap(h.problemController.DeleteTag))
	}
}
