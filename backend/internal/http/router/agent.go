package router

import (
	"time"

	"github.com/alonespig/zeorcode/backend/internal/http/middleware"
	"github.com/alonespig/zeorcode/backend/internal/http/response"

	"github.com/gin-gonic/gin"
)

// ==================== Agent(管理员 AI 工作台) ====================

// Agent is a standalone administrator workspace. It deliberately does not use
// /api/admin so its API can remain resource-oriented while sharing the same
// site-admin authorization boundary.
func (h *HttpServer) initAgentRouter(r *gin.Engine) {
	agent := r.Group("/api/agent", middleware.JWTAuth(h.auth), middleware.AdminRequired())
	{
		agent.GET("/conversations", response.Wrap(h.agentController.ListConversations))
		agent.POST("/conversations", response.Wrap(h.agentController.CreateConversation))
		agent.GET("/conversations/:id", response.Wrap(h.agentController.ConversationDetail))
		agent.PATCH("/conversations/:id", response.Wrap(h.agentController.UpdateConversation))
		agent.POST("/conversations/:id/turns",
			middleware.RateLimit(h.cache, "agent-turn", 120, time.Hour),
			h.agentController.Turn)
	}
}
