package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Submission ====================

func (h *HttpServer) initSubmissionRouter(r *gin.Engine) {
	api := r.Group("/api/submission")
	{
		api.POST("", middleware.JWTAuth(h.auth),
			response.Wrap(h.submissionController.Create))

		api.GET("", middleware.JWTAuthOptional(h.auth),
			response.Wrap(h.submissionController.GetSubmissionList))
		api.GET("/:id", middleware.JWTAuthOptional(h.auth),
			response.Wrap(h.submissionController.GetSubmissionByID))
		api.GET("/:id/stream", middleware.JWTAuthOptional(h.auth),
			h.submissionController.GetSubmissionStream)
	}
}
