package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Problem ====================

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
