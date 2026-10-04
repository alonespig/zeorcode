package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Contest ====================

func (h *HttpServer) initContestRouter(r *gin.Engine) {
	// 公开浏览：可匿名（列表 / 简介 / 排名 / 倒计时）。登录后自动带上 isRegistered / isSelf
	contestPub := r.Group("/api/contest", middleware.JWTAuthOptional(h.auth))
	{
		contestPub.GET("", response.Wrap(h.contestController.List))
		contestPub.GET("/:id", response.Wrap(h.contestController.GetContestDetail))
		contestPub.GET("/:id/desc", response.Wrap(h.contestController.GetContestDesc))
		contestPub.GET("/:id/rank", response.Wrap(h.contestController.GetContestRank))
		contestPub.GET("/:id/myrank", response.Wrap(h.contestController.GetMyContestRank))
		contestPub.GET("/:id/sse", h.contestController.SSEventStream)
	}

	// 需登录：报名 / 做题 / 提交 / 我的提交
	contest := r.Group("/api/contest", middleware.JWTAuth(h.auth))
	{
		contest.POST("/join", response.Wrap(h.contestController.JoinContest))
		contest.GET("/:id/problem", response.Wrap(h.contestController.GetContestProblemList))
		contest.GET("/:id/problem/:problemID", response.Wrap(h.contestController.GetContestProblem))
		contest.POST("/submit", response.Wrap(h.contestController.Submit))
		contest.GET("/:id/submission", response.Wrap(h.contestController.GetContestSubmissions))
		contest.GET("/:id/submit-info", response.Wrap(h.contestController.GetContestSubmitInfo))
	}

	// 管理员接口
	contestAdmin := r.Group("/api/contest", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		contestAdmin.POST("", response.Wrap(h.contestController.Create))
		contestAdmin.GET("/:id/info", response.Wrap(h.contestController.EditContest))
		contestAdmin.PUT("/:id", response.Wrap(h.contestController.Update))
	}
}
