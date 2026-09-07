package service

import (
	"testing"

	"zoj/internal/common/errcode"
	"zoj/internal/model"
	"zoj/internal/repository"
)

func TestIsHomeworkIntent(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{content: "给 C 语言班创建一个作业", want: true},
		{content: "帮我安排一组练习", want: true},
		{content: "作业的排行榜在哪里", want: false},
		{content: "介绍一下题库", want: false},
	}
	for _, tt := range tests {
		if got := isHomeworkIntent(tt.content); got != tt.want {
			t.Errorf("isHomeworkIntent(%q) = %v, want %v", tt.content, got, tt.want)
		}
	}
}

func TestApplyHomeworkSettings(t *testing.T) {
	draft := &homeworkDraft{}
	err := applyHomeworkSettings(draft, &homeworkSettingsInput{
		Title: "  指针练习  ", ProblemCount: 8, Difficulty: 2, RepeatPolicy: "exclude_all",
		StartTime: "2026-08-30 08:00:00", EndTime: "2026-09-06 08:00:00",
	})
	if err != nil {
		t.Fatalf("applyHomeworkSettings() error = %v", err)
	}
	if draft.Title != "指针练习" || draft.ProblemCount != 8 || draft.Difficulty != 2 {
		t.Fatalf("unexpected draft: %+v", draft)
	}
}

func TestApplyHomeworkSettingsRejectsInvalidWindow(t *testing.T) {
	err := applyHomeworkSettings(&homeworkDraft{}, &homeworkSettingsInput{
		Title: "练习", ProblemCount: 5, RepeatPolicy: "exclude_all",
		StartTime: "2026-09-06 08:00:00", EndTime: "2026-08-30 08:00:00",
	})
	requireAppErrorCode(t, err, errcode.InvalidParams)
}

func TestSortCandidatesPrefersTagCoverage(t *testing.T) {
	candidates := []repository.AgentProblemCandidate{
		{DisplayID: "1002", Difficulty: 1, Tags: []model.Tag{{ID: 1}}},
		{DisplayID: "1003", Difficulty: 3, Tags: []model.Tag{{ID: 1}, {ID: 2}}},
		{DisplayID: "1001", Difficulty: 2, Tags: []model.Tag{{ID: 2}}},
	}
	sortCandidates(candidates, []int64{1, 2})
	if candidates[0].DisplayID != "1003" {
		t.Fatalf("first candidate = %s, want 1003", candidates[0].DisplayID)
	}
	if candidates[1].DisplayID != "1002" || candidates[2].DisplayID != "1001" {
		t.Fatalf("single-tag candidates should then sort by difficulty: %+v", candidates)
	}
}

func TestDecodeAgentState(t *testing.T) {
	state, err := decodeAgentState(`{"activeWorkflow":"create_homework","pendingRequestId":12345678}`)
	if err != nil {
		t.Fatalf("decodeAgentState() error = %v", err)
	}
	if state.ActiveWorkflow != agentWorkflowHomework || state.PendingRequestID != 12345678 {
		t.Fatalf("unexpected state: %+v", state)
	}
	if _, err := decodeAgentState("{"); err == nil {
		t.Fatal("decodeAgentState() should reject malformed JSON")
	}
}
