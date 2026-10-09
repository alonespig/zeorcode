package service

import (
	"testing"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
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
	difficulty := 2
	draft := &homeworkDraft{}
	err := applyHomeworkSettings(draft, &homeworkSettingsInput{
		Title: "  指针练习  ", ProblemCount: 8, Difficulty: &difficulty, RepeatPolicy: "exclude_all",
		StartTime: "2026-08-30 08:00:00", EndTime: "2026-09-06 08:00:00",
	})
	if err != nil {
		t.Fatalf("applyHomeworkSettings() error = %v", err)
	}
	if draft.Title != "指针练习" || draft.ProblemCount != 8 || draft.Difficulty != 2 {
		t.Fatalf("unexpected draft: %+v", draft)
	}
}

func TestApplyHomeworkSettingsKeepsSelectedDifficulty(t *testing.T) {
	draft := &homeworkDraft{Difficulty: 3}
	err := applyHomeworkSettings(draft, &homeworkSettingsInput{
		Title: "提高练习", ProblemCount: 6, RepeatPolicy: "exclude_all",
		StartTime: "2026-08-30 08:00:00", EndTime: "2026-09-06 08:00:00",
	})
	if err != nil {
		t.Fatalf("applyHomeworkSettings() error = %v", err)
	}
	if draft.Difficulty != 3 {
		t.Fatalf("difficulty = %d, want 3", draft.Difficulty)
	}
}

func TestHomeworkPromptGraphBranchesByTrainingType(t *testing.T) {
	service := &AgentService{}
	state := &agentConversationState{Homework: &homeworkDraft{TeamName: "C 语言班", TeamDescription: "零基础入门"}}

	reply := service.askHomeworkTrainingType(state)
	if state.Step != agentStepChooseTraining || len(reply.Blocks) != 1 || len(reply.Blocks[0].Options) != 3 {
		t.Fatalf("unexpected training prompt: state=%+v reply=%+v", state, reply)
	}
	if !reply.Blocks[0].Options[0].Recommended {
		t.Fatal("beginner option should be recommended for a beginner team")
	}

	state.Homework.TrainingType = homeworkTrainingAlgorithm
	reply = service.askHomeworkDifficulty(state)
	if state.Step != agentStepChooseDifficulty || len(reply.Blocks[0].Options) != 4 {
		t.Fatalf("unexpected difficulty prompt: state=%+v reply=%+v", state, reply)
	}
	if !reply.Blocks[0].Options[1].Recommended {
		t.Fatal("medium difficulty should be recommended for algorithm training")
	}
}

func TestSelectHomeworkPromptTagsKeepsTeamRecommendation(t *testing.T) {
	tags := []repository.AgentProblemTagStat{
		{ID: 1, Name: "二分", ProblemCount: 20},
		{ID: 2, Name: "动态规划", ProblemCount: 5},
		{ID: 3, Name: "搜索", ProblemCount: 10},
	}
	selected := selectHomeworkPromptTags(tags, "本周学习动态规划", 2)
	if len(selected) != 2 || selected[0].ID != 2 {
		t.Fatalf("selected tags = %+v, want recommended tag first", selected)
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
