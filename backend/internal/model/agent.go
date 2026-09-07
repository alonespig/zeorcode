package model

import "time"

const (
	AgentConversationActive   = 0
	AgentConversationArchived = 1
)

const (
	AgentRunRunning   = "running"
	AgentRunWaiting   = "waiting"
	AgentRunCompleted = "completed"
	AgentRunFailed    = "failed"
	AgentRunCancelled = "cancelled"
)

const (
	AgentActionPending   = "pending"
	AgentActionExecuting = "executing"
	AgentActionSucceeded = "succeeded"
	AgentActionRejected  = "rejected"
	AgentActionFailed    = "failed"
)

// AgentConversation stores durable conversation state. Redis is only used for
// short-lived run locks; losing Redis data must not lose the conversation.
type AgentConversation struct {
	ID        int64  `gorm:"primaryKey"`
	PublicID  int64  `gorm:"column:public_id;uniqueIndex;not null"`
	UserID    int64  `gorm:"index:idx_ai_conversation_user_updated,priority:1;not null"`
	Title     string `gorm:"size:120;not null;default:'新对话'"`
	Summary   string `gorm:"type:text"`
	StateJSON string `gorm:"column:state_json;type:longtext"`
	Status    int    `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"index:idx_ai_conversation_user_updated,priority:2"`
}

func (AgentConversation) TableName() string { return "ai_conversations" }

type AgentMessage struct {
	ID             int64     `gorm:"primaryKey"`
	PublicID       int64     `gorm:"column:public_id;uniqueIndex;not null"`
	ConversationID int64     `gorm:"index:idx_ai_message_conversation_created,priority:1;not null"`
	RunID          int64     `gorm:"index"`
	Role           string    `gorm:"size:16;not null"`
	Kind           string    `gorm:"size:24;not null;default:'text'"`
	Content        string    `gorm:"type:longtext"`
	BlocksJSON     string    `gorm:"column:blocks_json;type:longtext"`
	CreatedAt      time.Time `gorm:"index:idx_ai_message_conversation_created,priority:2"`
}

func (AgentMessage) TableName() string { return "ai_messages" }

type AgentRun struct {
	ID             int64  `gorm:"primaryKey"`
	PublicID       int64  `gorm:"column:public_id;uniqueIndex;not null"`
	ConversationID int64  `gorm:"index;not null"`
	Status         string `gorm:"size:16;not null"`
	Model          string `gorm:"size:100;not null;default:''"`
	ErrorCode      string `gorm:"size:64;not null;default:''"`
	StartedAt      time.Time
	FinishedAt     *time.Time
}

func (AgentRun) TableName() string { return "ai_runs" }

type AgentAction struct {
	ID              int64  `gorm:"primaryKey"`
	PublicID        int64  `gorm:"column:public_id;uniqueIndex;not null"`
	ConversationID  int64  `gorm:"index;not null"`
	RunID           int64  `gorm:"index"`
	ActionType      string `gorm:"size:40;not null"`
	ArtifactJSON    string `gorm:"column:artifact_json;type:longtext;not null"`
	ArtifactVersion int    `gorm:"column:artifact_version;not null"`
	PayloadHash     string `gorm:"column:payload_hash;size:64;not null"`
	Status          string `gorm:"size:16;index;not null"`
	ResultPublicID  int64  `gorm:"column:result_public_id;not null;default:0"`
	ApprovedBy      int64  `gorm:"column:approved_by;not null;default:0"`
	ApprovedAt      *time.Time
	ErrorCode       string `gorm:"size:64;not null;default:''"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (AgentAction) TableName() string { return "ai_actions" }
