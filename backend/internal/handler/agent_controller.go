package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type AgentController struct {
	service *service.AgentService
}

func NewAgentController(agentService *service.AgentService) *AgentController {
	return &AgentController{service: agentService}
}

func (h *AgentController) ListConversations(c *gin.Context) (any, error) {
	var req dto.AgentConversationListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return h.service.ListConversations(c.Request.Context(), userID, &req)
}

func (h *AgentController) CreateConversation(c *gin.Context) (any, error) {
	var req dto.CreateAgentConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return h.service.CreateConversation(c.Request.Context(), userID, &req)
}

func (h *AgentController) ConversationDetail(c *gin.Context) (any, error) {
	conversationID, err := parsePublicIDParam(c, "id", "对话")
	if err != nil {
		return nil, err
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return h.service.ConversationDetail(c.Request.Context(), conversationID, userID)
}

func (h *AgentController) UpdateConversation(c *gin.Context) (any, error) {
	conversationID, err := parsePublicIDParam(c, "id", "对话")
	if err != nil {
		return nil, err
	}
	var req dto.UpdateAgentConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	return nil, h.service.UpdateConversation(c.Request.Context(), conversationID, userID, &req)
}

// Turn streams a single Agent turn. The service emits transport-neutral events;
// this controller is the only place that knows about SSE framing.
func (h *AgentController) Turn(c *gin.Context) {
	conversationID, err := parsePublicIDParam(c, "id", "对话")
	if err != nil {
		writeAgentJSONError(c, err)
		return
	}
	var req dto.AgentTurnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeAgentJSONError(c, errcode.ErrInvalidParams.Wrap(err))
		return
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		writeAgentJSONError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	emit := func(event dto.AgentEvent) error {
		payload, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			return marshalErr
		}
		if _, writeErr := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, payload); writeErr != nil {
			return writeErr
		}
		c.Writer.Flush()
		return nil
	}

	if err := h.service.StreamTurn(c.Request.Context(), conversationID, userID, role == 1, &req, emit); err != nil {
		code, message := agentPublicError(err)
		_ = emit(dto.AgentEvent{Type: "run.failed", Data: map[string]any{"code": code, "message": message}})
	}
}

func writeAgentJSONError(c *gin.Context, err error) {
	code, message := agentPublicError(err)
	c.JSON(http.StatusOK, gin.H{"code": code, "msg": message})
}

func agentPublicError(err error) (int, string) {
	var appErr *errcode.AppErr
	if errors.As(err, &appErr) {
		return int(appErr.Code), appErr.Msg
	}
	return int(errcode.UnknownError), "AI 助手暂时不可用"
}
