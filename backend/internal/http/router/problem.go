package router

import (
	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/http/middleware"
	"github.com/alonespig/zeorcode/backend/internal/http/response"

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
	admin := r.Group("/api", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		archive := middleware.RequirePermission(h.permissions, authz.ProblemArchive)
		write := middleware.RequirePermission(h.permissions, authz.ProblemWrite)

		admin.GET("/admin/problem-archive", archive, h.problemController.DownloadProblemArchive)
		admin.POST("/admin/problem-archive", archive, response.Wrap(h.problemController.ImportProblemArchive))
		admin.POST("/problems", write, response.Wrap(h.problemController.CreateProblem))
		admin.PUT("/problems/:id", write, response.Wrap(h.problemController.UpdateProblem))
		admin.GET("/problems/:id/edit", write, response.Wrap(h.problemController.GetProblemForEdit))
		admin.GET("/problems/:id/testdata", write, response.Wrap(h.problemController.ListFiles))
		admin.GET("/problems/:id/testdata/content", write, response.Wrap(h.problemController.GetFileContent))
		admin.POST("/problems/:id/testdata", write, response.Wrap(h.problemController.UploadFile))
		admin.POST("/problems/:id/testdata/zip", write, response.Wrap(h.problemController.UploadTestcaseZip))
		admin.DELETE("/problems/:id/testdata", write, response.Wrap(h.problemController.DeleteFiles))
		admin.GET("/problems/:id/testdata/download", write, h.problemController.DownloadFiles)
	}

	tagRead := middleware.RequirePermission(h.permissions, authz.TagRead)
	tagWrite := middleware.RequirePermission(h.permissions, authz.TagWrite)
	tagAdmin := r.Group("/api/admin/tags", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		tagAdmin.GET("", tagRead, response.Wrap(h.problemController.GetAdminTagList))
		tagAdmin.POST("", tagWrite, response.Wrap(h.problemController.CreateTag))
		tagAdmin.PUT("/:id", tagWrite, response.Wrap(h.problemController.UpdateTag))
		tagAdmin.DELETE("/:id", tagWrite, response.Wrap(h.problemController.DeleteTag))
	}
}
