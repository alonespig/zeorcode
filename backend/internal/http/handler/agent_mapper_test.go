package handler

import (
	"encoding/json"
	"testing"
	"time"

	"zoj/internal/service"
)

// 验证 message.completed 的 Data（service.AgentMessage）经 mapper 后 JSON 与旧协议一致。
func TestToAgentMessageRespSSEProtocol(t *testing.T) {
	createdAt := time.Date(2026, 10, 5, 12, 34, 56, 0, time.Local)
	msg := service.AgentMessage{
		ID:        123,
		Role:      "assistant",
		Kind:      "blocks",
		Content:   "hello",
		Blocks:    []service.AgentBlock{{Type: "notice", Title: "标题"}},
		CreatedAt: createdAt,
	}
	got, err := json.Marshal(toAgentMessageResp(msg))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":123,"role":"assistant","kind":"blocks","content":"hello","blocks":[{"type":"notice","title":"标题"}],"createdAt":"2026-10-05 12:34:56"}`
	if string(got) != want {
		t.Fatalf("message JSON = %s\nwant %s", got, want)
	}
}

// 验证 interaction.required 的 Data（[]service.AgentBlock）经 mapper 后 JSON 与旧协议一致。
func TestToAgentBlocksSSEProtocol(t *testing.T) {
	blocks := []service.AgentBlock{{
		Type:      "single_select",
		RequestID: 9,
		Options:   []service.AgentOption{{Label: "A", Value: float64(1)}},
	}}
	got, err := json.Marshal(toAgentBlocks(blocks))
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"type":"single_select","requestId":9,"options":[{"label":"A","value":1}]}]`
	if string(got) != want {
		t.Fatalf("blocks JSON = %s\nwant %s", got, want)
	}
}
