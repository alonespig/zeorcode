package router

import (
	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Contest ====================

func (h *HttpServer) initContestRouter(r *gin.Engine) {
	// 公开浏览：可匿名。题目接口仍由 Service 按比赛状态、公开性和报名状态鉴权。
	contestPub := r.Group("/api/contest", middleware.JWTAuthOptional(h.auth))
	{
		contestPub.GET("", response.Wrap(h.contestController.List))
		contestPub.GET("/:id", response.Wrap(h.contestController.GetContestDetail))
		contestPub.GET("/:id/desc", response.Wrap(h.contestController.GetContestDesc))
		contestPub.GET("/:id/problem", response.Wrap(h.contestController.GetContestProblemList))
		contestPub.GET("/:id/problem/:problemID", response.Wrap(h.contestController.GetContestProblem))
		contestPub.GET("/:id/rank", response.Wrap(h.contestController.GetContestRank))
		contestPub.GET("/:id/myrank", response.Wrap(h.contestController.GetMyContestRank))
		contestPub.GET("/:id/sse", h.contestController.SSEventStream)
	}

	// 需登录：报名 / 提交 / 我的提交
	contest := r.Group("/api/contest", middleware.JWTAuth(h.auth))
	{
		contest.POST("/join", response.Wrap(h.contestController.JoinContest))
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

	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		// 比赛榜从 submissions 明细整场重算（榜单漂移时手动重建；路①下重判已不需要它）
		admin.POST("/contest/:id/recompute", response.Wrap(h.contestController.RecomputeContest))
	}
}
