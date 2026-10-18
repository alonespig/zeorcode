package router

import (
	"time"

	"github.com/alonespig/zeorcode/backend/internal/http/middleware"
	"github.com/alonespig/zeorcode/backend/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Submission ====================

func (h *HttpServer) initSubmissionRouter(r *gin.Engine) {
	api := r.Group("/api/submission")
	{
		api.POST("", middleware.JWTAuth(h.auth),
			response.Wrap(h.submissionController.Create))
		api.POST("/run",
			middleware.RateLimit(h.cache, "code-run", 120, time.Minute),
			middleware.JWTAuth(h.auth),
			response.Wrap(h.submissionController.Run))

		api.GET("", middleware.JWTAuthOptional(h.auth),
			response.Wrap(h.submissionController.GetSubmissionList))
		api.GET("/:id", middleware.JWTAuthOptional(h.auth),
			response.Wrap(h.submissionController.GetSubmissionByID))
		api.GET("/:id/stream", middleware.JWTAuthOptional(h.auth),
			h.submissionController.GetSubmissionStream)
	}

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired(), middleware.AdminAudit(h.audit))
	{
		// 重判（超管）：单条 / 整场 / 某比赛某题
		admin.POST("/submission/:id/rejudge", response.Wrap(h.submissionController.Rejudge))
		admin.POST("/contest/:id/rejudge", response.Wrap(h.submissionController.RejudgeContest))
		admin.POST("/contest/:id/problem/:pid/rejudge", response.Wrap(h.submissionController.RejudgeContestProblem))
	}
}
