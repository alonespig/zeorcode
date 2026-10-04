package router

import (
	"time"

	"zoj/internal/http/middleware"
	"zoj/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Admin ====================

func (h *HttpServer) initAdminRouter(r *gin.Engine) {
	admin := r.Group("/api/admin", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/users", response.Wrap(h.adminController.ListUsers))
		admin.GET("/users/import/template", h.adminController.DownloadUserImportTemplate)
		admin.POST("/users/import", response.Wrap(h.adminController.ImportUsers))
		admin.PUT("/users/:id/status", response.Wrap(h.adminController.SetUserStatus))
		admin.PUT("/users/:id/role", response.Wrap(h.adminController.SetUserRole))
		admin.GET("/judge/status", response.Wrap(h.adminController.JudgeStatus))
		admin.GET("/languages", response.Wrap(h.languageController.AdminList))
		admin.POST("/languages", response.Wrap(h.languageController.Create))
		admin.PUT("/languages/:id", response.Wrap(h.languageController.Update))
		admin.DELETE("/languages/:id", response.Wrap(h.languageController.Delete))
		admin.GET("/posts/pending", response.Wrap(h.postController.ReviewListPending))
		admin.PUT("/posts/:id/review", response.Wrap(h.postController.ReviewPost))

		// 比赛榜从 submissions 明细整场重算（榜单漂移时手动重建；路①下重判已不需要它）
		admin.POST("/contest/:id/recompute", response.Wrap(h.contestController.RecomputeContest))

		// 重判（超管）：单条 / 整场 / 某比赛某题
		admin.POST("/submission/:id/rejudge", response.Wrap(h.submissionController.Rejudge))
		admin.POST("/contest/:id/rejudge", response.Wrap(h.submissionController.RejudgeContest))
		admin.POST("/contest/:id/problem/:pid/rejudge", response.Wrap(h.submissionController.RejudgeContestProblem))

		// 远程账号管理
		admin.GET("/remote-account", response.Wrap(h.remoteController.ListAccounts))
		admin.POST("/remote-account", response.Wrap(h.remoteController.CreateAccount))
		admin.PUT("/remote-account/:id", response.Wrap(h.remoteController.UpdateAccount))
		admin.DELETE("/remote-account/:id", response.Wrap(h.remoteController.DeleteAccount))
	}
}

// ==================== Remote(远程拉题) ====================

func (h *HttpServer) initRemoteRouter(r *gin.Engine) {
	admin := r.Group("/api/remote", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		admin.GET("/oj", response.Wrap(h.remoteController.ListOJ))
		admin.GET("/problem", response.Wrap(h.remoteController.GetRemoteProblem))
	}
}

// ==================== Agent(管理员 AI 工作台) ====================

// Agent is a standalone administrator workspace. It deliberately does not use
// /api/admin so its API can remain resource-oriented while sharing the same
// site-admin authorization boundary.
func (h *HttpServer) initAgentRouter(r *gin.Engine) {
	agent := r.Group("/api/agent", middleware.JWTAuth(h.auth), middleware.AdminRequired(),
		middleware.RateLimit(h.cache, "agent", 120, time.Hour))
	{
		agent.GET("/conversations", response.Wrap(h.agentController.ListConversations))
		agent.POST("/conversations", response.Wrap(h.agentController.CreateConversation))
		agent.GET("/conversations/:id", response.Wrap(h.agentController.ConversationDetail))
		agent.PATCH("/conversations/:id", response.Wrap(h.agentController.UpdateConversation))
		agent.POST("/conversations/:id/turns", h.agentController.Turn)
	}
}
