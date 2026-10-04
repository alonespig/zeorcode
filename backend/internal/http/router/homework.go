package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Homework ====================

func (h *HttpServer) initHomeworkRouter(r *gin.Engine) {
	// 作业读接口用可选登录：service 里非团队成员一律按「作业不存在」处理，
	// 不泄露作业的存在，因此这里不必强制登录
	pub := r.Group("/api")
	{
		pub.GET("/team/:id/homework", middleware.JWTAuthOptional(h.auth), response.Wrap(h.homeworkController.ListByTeam))
		pub.GET("/homework/:hid", middleware.JWTAuthOptional(h.auth), response.Wrap(h.homeworkController.Detail))
		pub.GET("/homework/:hid/problem/:problemID", middleware.JWTAuthOptional(h.auth), response.Wrap(h.homeworkController.Problem))
		pub.GET("/homework/:hid/rank", middleware.JWTAuthOptional(h.auth), response.Wrap(h.homeworkController.Rank))
		pub.GET("/homework/:hid/rank/export", middleware.JWTAuthOptional(h.auth), h.homeworkController.ExportRank)
		pub.GET("/homework/:hid/submission", middleware.JWTAuthOptional(h.auth), response.Wrap(h.homeworkController.Submissions))
		pub.GET("/homework/:hid/submission/:sid", middleware.JWTAuthOptional(h.auth), response.Wrap(h.homeworkController.SubmissionDetail))
	}

	auth := r.Group("/api", middleware.JWTAuth(h.auth))
	{
		auth.POST("/team/:id/homework", response.Wrap(h.homeworkController.Create))
		auth.PUT("/homework/:hid", response.Wrap(h.homeworkController.Update))
		auth.DELETE("/homework/:hid", response.Wrap(h.homeworkController.Delete))
		auth.POST("/homework/:hid/submit", response.Wrap(h.homeworkController.Submit))
	}
}
