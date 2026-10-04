package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

const agentConversationTimeLayout = "2006-01-02 15:04:05"

func toAgentConversationListParams(req dto.AgentConversationListReq) service.AgentConversationListParams {
	return service.AgentConversationListParams{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword}
}

func toCreateAgentConversationParams(req dto.CreateAgentConversationReq) service.CreateAgentConversationParams {
	return service.CreateAgentConversationParams{Title: req.Title}
}

func toUpdateAgentConversationParams(req dto.UpdateAgentConversationReq) service.UpdateAgentConversationParams {
	return service.UpdateAgentConversationParams{Title: req.Title, Archived: req.Archived}
}

func toAgentOptions(options []service.AgentOption) []dto.AgentOption {
	items := make([]dto.AgentOption, 0, len(options))
	for _, o := range options {
		items = append(items, dto.AgentOption{Label: o.Label, Value: o.Value, Description: o.Description, Disabled: o.Disabled, Recommended: o.Recommended})
	}
	return items
}

func toAgentColumns(columns []service.AgentTableColumn) []dto.AgentTableColumn {
	items := make([]dto.AgentTableColumn, 0, len(columns))
	for _, c := range columns {
		items = append(items, dto.AgentTableColumn{Key: c.Key, Label: c.Label, Width: c.Width, Align: c.Align, Formatter: c.Formatter})
	}
	return items
}

func toAgentBlocks(blocks []service.AgentBlock) []dto.AgentBlock {
	items := make([]dto.AgentBlock, 0, len(blocks))
	for _, b := range blocks {
		items = append(items, dto.AgentBlock{
			Type: b.Type, RequestID: b.RequestID, Title: b.Title, Description: b.Description,
			Required: b.Required, Options: toAgentOptions(b.Options), Columns: toAgentColumns(b.Columns),
			Rows: b.Rows, Payload: b.Payload,
		})
	}
	return items
}

func toAgentMessageResp(m service.AgentMessage) dto.AgentMessageResp {
	return dto.AgentMessageResp{
		ID:        m.ID,
		Role:      m.Role,
		Kind:      m.Kind,
		Content:   m.Content,
		Blocks:    toAgentBlocks(m.Blocks),
		CreatedAt: m.CreatedAt.Format(agentConversationTimeLayout),
	}
}

func toAgentTurnParams(req dto.AgentTurnReq) service.AgentTurnParams {
	return service.AgentTurnParams{
		Type:         req.Type,
		Content:      req.Content,
		RequestID:    req.RequestID,
		Value:        req.Value,
		ActionID:     req.ActionID,
		DraftVersion: req.DraftVersion,
	}
}

func toAgentEventData(data any) any {
	switch v := data.(type) {
	case service.AgentMessage:
		return toAgentMessageResp(v)
	case []service.AgentBlock:
		return toAgentBlocks(v)
	default:
		return data
	}
}

func toAgentConversationResp(c service.AgentConversation) dto.AgentConversationItem {
	return dto.AgentConversationItem{
		ID:        c.ID,
		Title:     c.Title,
		Status:    c.Status,
		UpdatedAt: c.UpdatedAt.Format(agentConversationTimeLayout),
	}
}

func toAgentConversationListResp(r *service.AgentConversationList) *dto.AgentConversationListResp {
	items := make([]dto.AgentConversationItem, 0, len(r.List))
	for _, c := range r.List {
		items = append(items, toAgentConversationResp(c))
	}
	return &dto.AgentConversationListResp{Total: r.Total, List: items}
}

func toAgentConversationDetailResp(r *service.AgentConversationDetail) *dto.AgentConversationDetailResp {
	messages := make([]dto.AgentMessageResp, 0, len(r.Messages))
	for _, m := range r.Messages {
		messages = append(messages, toAgentMessageResp(m))
	}
	return &dto.AgentConversationDetailResp{
		ID:               r.ID,
		Title:            r.Title,
		Status:           r.Status,
		PendingRequestID: r.PendingRequestID,
		PendingActionID:  r.PendingActionID,
		Messages:         messages,
	}
}
