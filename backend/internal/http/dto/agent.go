package dto

import "encoding/json"

type AgentConversationListReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=50"`
	Keyword  string `form:"q"`
}

type AgentConversationItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

type AgentConversationListResp struct {
	Total int                     `json:"total"`
	List  []AgentConversationItem `json:"list"`
}

type CreateAgentConversationReq struct {
	Title string `json:"title" binding:"max=120"`
}

type UpdateAgentConversationReq struct {
	Title    *string `json:"title" binding:"omitempty,max=120"`
	Archived *bool   `json:"archived"`
}

type AgentOption struct {
	Label       string `json:"label"`
	Value       any    `json:"value"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
	Recommended bool   `json:"recommended,omitempty"`
}

type AgentTableColumn struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Width     int    `json:"width,omitempty"`
	Align     string `json:"align,omitempty"`
	Formatter string `json:"formatter,omitempty"`
}

type AgentBlock struct {
	Type        string             `json:"type"`
	RequestID   int64              `json:"requestId,omitempty"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Options     []AgentOption      `json:"options,omitempty"`
	Columns     []AgentTableColumn `json:"columns,omitempty"`
	Rows        []map[string]any   `json:"rows,omitempty"`
	Payload     any                `json:"payload,omitempty"`
}

type AgentMessageResp struct {
	ID        int64        `json:"id"`
	Role      string       `json:"role"`
	Kind      string       `json:"kind"`
	Content   string       `json:"content"`
	Blocks    []AgentBlock `json:"blocks"`
	CreatedAt string       `json:"createdAt"`
}

type AgentConversationDetailResp struct {
	ID               int64              `json:"id"`
	Title            string             `json:"title"`
	Status           int                `json:"status"`
	PendingRequestID int64              `json:"pendingRequestId,omitempty"`
	PendingActionID  int64              `json:"pendingActionId,omitempty"`
	Messages         []AgentMessageResp `json:"messages"`
}

type AgentTurnReq struct {
	Type         string          `json:"type" binding:"required,oneof=message interaction_response action_approval action_rejection"`
	Content      string          `json:"content" binding:"max=4000"`
	RequestID    int64           `json:"requestId"`
	Value        json.RawMessage `json:"value"`
	ActionID     int64           `json:"actionId"`
	DraftVersion int             `json:"draftVersion"`
}

type AgentEvent struct {
	Type  string `json:"type"`
	RunID int64  `json:"runId,omitempty"`
	Data  any    `json:"data,omitempty"`
}
