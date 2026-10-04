package service

import "time"

// ===== 查询 / 命令输入 =====

type AgentConversationListParams struct {
	Page     int
	PageSize int
	Keyword  string
}

type CreateAgentConversationParams struct {
	Title string
}

type UpdateAgentConversationParams struct {
	Title    *string
	Archived *bool
}

// ===== Block 值对象（JSON tag 与旧 DTO 完全一致，保证 DB 持久化与 SSE 序列化无损）=====

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

// ===== 结果 =====

type AgentConversation struct {
	ID        int64
	Title     string
	Status    int
	UpdatedAt time.Time
}

type AgentConversationList struct {
	Total int
	List  []AgentConversation
}

// AgentMessage 一条消息。CreatedAt 保留字符串：会直接放进 AgentEvent.Data 序列化，
// 必须与旧 DTO 的格式化字符串一致。
type AgentMessage struct {
	ID        int64        `json:"id"`
	Role      string       `json:"role"`
	Kind      string       `json:"kind"`
	Content   string       `json:"content"`
	Blocks    []AgentBlock `json:"blocks"`
	CreatedAt string       `json:"createdAt"`
}

type AgentConversationDetail struct {
	ID               int64
	Title            string
	Status           int
	PendingRequestID int64
	PendingActionID  int64
	Messages         []AgentMessage
}
