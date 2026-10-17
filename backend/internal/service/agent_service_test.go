package service

import (
	"strings"
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

func TestIsProblemSetIntent(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{content: "帮我创建一个二分题单", want: true},
		{content: "创建一个提单", want: true},
		{content: "帮我建一个刷题单", want: true},
		{content: "来一份图论题目单", want: true},
		{content: "整理一份图论 problem set", want: true},
		{content: "这个题单里有哪些题", want: false},
		{content: "提单的状态是什么", want: false},
		{content: "帮我提交单元测试", want: false},
		{content: "创建团队作业", want: false},
	}
	for _, tt := range tests {
		if got := isProblemSetIntent(tt.content); got != tt.want {
			t.Errorf("isProblemSetIntent(%q) = %v, want %v", tt.content, got, tt.want)
		}
	}
}

func TestStartProblemSetWorkflowRequiresAdmin(t *testing.T) {
	service := &AgentService{}
	state := &agentConversationState{}
	reply, err := service.startProblemSetWorkflow(state, false)
	if err != nil {
		t.Fatalf("startProblemSetWorkflow() error = %v", err)
	}
	if state.ActiveWorkflow != "" || !strings.Contains(reply.Content, "管理员权限") {
		t.Fatalf("unexpected non-admin result: state=%+v reply=%+v", state, reply)
	}

	reply, err = service.startProblemSetWorkflow(state, true)
	if err != nil {
		t.Fatalf("startProblemSetWorkflow() error = %v", err)
	}
	if state.ActiveWorkflow != agentWorkflowProblemSet || state.ProblemSet == nil || state.Step != agentStepChooseTraining {
		t.Fatalf("unexpected admin state: %+v", state)
	}
	if len(reply.Blocks) != 1 || reply.Blocks[0].Type != "single_select" {
		t.Fatalf("unexpected prompt: %+v", reply)
	}
}

func TestApplyProblemSetSettings(t *testing.T) {
	draft := &problemSetDraft{}
	err := applyProblemSetSettings(draft, &problemSetSettingsInput{
		Title: "  二分专项  ", Description: "基础到提高", ProblemCount: 10,
		Published: model.ProblemSetDraft, Visibility: model.ProblemSetPublic,
	})
	if err != nil {
		t.Fatalf("applyProblemSetSettings() error = %v", err)
	}
	if draft.Title != "二分专项" || draft.ProblemCount != 10 || draft.Description != "基础到提高" {
		t.Fatalf("unexpected draft: %+v", draft)
	}

	err = applyProblemSetSettings(draft, &problemSetSettingsInput{
		Title: "私有题单", ProblemCount: 5, Published: model.ProblemSetPublished, Visibility: model.ProblemSetInviteOnly,
	})
	requireAppErrorCode(t, err, errcode.InvalidParams)
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
	if draft.Title != "指针练习" || draft.Sequence != 1 || draft.ProblemCount != 8 || draft.Difficulty != 2 {
		t.Fatalf("unexpected draft: %+v", draft)
	}
}

func TestSelectAgentTeamUsesNextHomeworkSequence(t *testing.T) {
	draft := &homeworkDraft{TrainingType: homeworkTrainingAlgorithm}
	selectAgentTeam(draft, repository.AgentTeam{
		ID: 7, PublicID: 12345678, Name: "算法班", HomeworkCount: 4,
	})

	if draft.Sequence != 5 {
		t.Fatalf("sequence = %d, want 5", draft.Sequence)
	}
	if got := homeworkDefaultTitle(draft); got != "第 5 次作业 · 算法竞赛" {
		t.Fatalf("homeworkDefaultTitle() = %q", got)
	}
}

func TestHomeworkDefaultTitleUsesSelectedKnowledge(t *testing.T) {
	draft := &homeworkDraft{
		Sequence: 7, TrainingType: homeworkTrainingBeginner,
		TagNames: []string{"输入输出与顺序结构", "条件判断与循环", "一维与二维数组"},
	}
	if got := homeworkDefaultTitle(draft); got != "第 7 次作业 · 输入输出与顺序结构、条件判断与循环等" {
		t.Fatalf("homeworkDefaultTitle() = %q", got)
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
	if state.Step != agentStepChooseTraining || len(reply.Blocks) != 1 || len(reply.Blocks[0].Options) != 4 {
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

func TestHomeworkScopeOptionsSeparateDataStructuresAndCompetition(t *testing.T) {
	dataStructure := homeworkScopeOptions(homeworkTrainingDataStructure)
	if len(dataStructure) != 4 || dataStructure[0].Value != homeworkScopeLinearStructure {
		t.Fatalf("unexpected data-structure scopes: %+v", dataStructure)
	}
	competition := homeworkScopeOptions(homeworkTrainingAlgorithm)
	if len(competition) < 6 || competition[0].Value != homeworkScopeBasicAlgorithm {
		t.Fatalf("unexpected competition scopes: %+v", competition)
	}
	if !validHomeworkScope(homeworkTrainingAlgorithm, homeworkScopeGreedyConstruct) {
		t.Fatal("competition scopes should include greedy and construction")
	}
	if !tagMatchesHomeworkScope("最短路", homeworkScopeGraph) || tagMatchesHomeworkScope("背包DP", homeworkScopeGraph) {
		t.Fatal("graph scope should contain shortest path but not knapsack DP")
	}
	concepts := homeworkConceptOptions(homeworkScopeGreedyConstruct)
	if len(concepts) != 1 || concepts[0].Name != "构造思维" {
		t.Fatalf("unexpected construction concepts: %+v", concepts)
	}
}

func TestDecodeKnowledgeSelectionSupportsKeysAndLegacyTagIDs(t *testing.T) {
	selection, err := decodeKnowledgeSelection([]byte(`["concept:stack","custom:单调栈",14]`))
	if err != nil {
		t.Fatalf("decodeKnowledgeSelection() error = %v", err)
	}
	if len(selection) != 3 || selection[0] != "concept:stack" || selection[1] != "custom:单调栈" || selection[2] != "tag:14" {
		t.Fatalf("unexpected selection: %+v", selection)
	}
}

func TestDecodeCustomKnowledgeKey(t *testing.T) {
	if name, ok := decodeCustomKnowledgeKey("custom:  单调栈  "); !ok || name != "单调栈" {
		t.Fatalf("decodeCustomKnowledgeKey() = %q, %v", name, ok)
	}
	if _, ok := decodeCustomKnowledgeKey("custom:\n"); ok {
		t.Fatal("empty custom knowledge should be rejected")
	}
}

func TestAutomaticHomeworkDifficulty(t *testing.T) {
	if difficulty, automatic := automaticHomeworkDifficulty(&homeworkDraft{TrainingType: homeworkTrainingBeginner}); !automatic || difficulty != 1 {
		t.Fatalf("beginner difficulty = %d, automatic = %v", difficulty, automatic)
	}
	if _, automatic := automaticHomeworkDifficulty(&homeworkDraft{TrainingType: homeworkTrainingAlgorithm}); automatic {
		t.Fatal("algorithm training should still ask for difficulty")
	}
}

func TestFilterHomeworkCandidatesRejectsAdvancedProblemForBeginner(t *testing.T) {
	draft := &homeworkDraft{
		TrainingType:  homeworkTrainingBeginner,
		KnowledgeKeys: []string{"concept:input_output"},
		TagNames:      []string{"输入输出与顺序结构"},
	}
	candidates := []repository.AgentProblemCandidate{
		{DisplayID: "L1002", Name: "三维偏序", Difficulty: 1, Tags: []model.Tag{{Name: "差分约束"}, {Name: "网络流"}}},
		{DisplayID: "B1001", Name: "两数求和", Difficulty: 1},
	}

	filtered := filterHomeworkCandidates(candidates, draft)
	if len(filtered) != 1 || filtered[0].DisplayID != "B1001" {
		t.Fatalf("unexpected beginner candidates: %+v", filtered)
	}
}

func TestFilterHomeworkCandidatesMatchesCustomKnowledge(t *testing.T) {
	draft := &homeworkDraft{
		TrainingType:  homeworkTrainingDataStructure,
		KnowledgeKeys: []string{"custom:单调栈"},
		TagNames:      []string{"单调栈"},
	}
	candidates := []repository.AgentProblemCandidate{
		{DisplayID: "A", Name: "单调栈入门"},
		{DisplayID: "B", Name: "队列练习"},
	}

	filtered := filterHomeworkCandidates(candidates, draft)
	if len(filtered) != 1 || filtered[0].DisplayID != "A" {
		t.Fatalf("unexpected custom-knowledge candidates: %+v", filtered)
	}
}

func TestRequirementExcludesTag(t *testing.T) {
	if !requirementExcludesTag("希望是基础题，不要动态规划DP", "动态规划DP") {
		t.Fatal("negated tag should be excluded")
	}
	if requirementExcludesTag("希望练习动态规划DP", "动态规划DP") {
		t.Fatal("positive tag should not be excluded")
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

func TestApplyHomeworkSettingsRejectsInvalidSequence(t *testing.T) {
	err := applyHomeworkSettings(&homeworkDraft{}, &homeworkSettingsInput{
		Title: "练习", Sequence: 1000, ProblemCount: 5, RepeatPolicy: "exclude_all",
		StartTime: "2026-08-30 08:00:00", EndTime: "2026-09-06 08:00:00",
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

func TestApplyHomeworkCandidateAssessmentRanksByContentScore(t *testing.T) {
	candidates := []repository.AgentProblemCandidate{
		{DisplayID: "A", Name: "循环求和"},
		{DisplayID: "B", Name: "最短路"},
		{DisplayID: "C", Name: "数组统计"},
	}
	raw := "```json\n" + `{"items":[
		{"id":"C","score":82,"reason":"遍历数组并维护计数"},
		{"id":"A","score":91,"reason":"使用循环累加输入数据"},
		{"id":"B","score":45,"reason":"主要考查图论"},
		{"id":"INVENTED","score":100,"reason":"不存在"}
	]}` + "\n```"

	ranked, ok := applyHomeworkCandidateAssessment(candidates, raw, 3)
	if !ok {
		t.Fatal("assessment should be accepted")
	}
	if len(ranked) != 2 || ranked[0].DisplayID != "A" || ranked[1].DisplayID != "C" {
		t.Fatalf("unexpected ranked candidates: %+v", ranked)
	}
	if ranked[0].SelectionReason != "使用循环累加输入数据" {
		t.Fatalf("selection reason = %q", ranked[0].SelectionReason)
	}
}

func TestApplyHomeworkCandidateAssessmentRejectsIncompleteResult(t *testing.T) {
	candidates := []repository.AgentProblemCandidate{{DisplayID: "A"}, {DisplayID: "B"}}
	ranked, ok := applyHomeworkCandidateAssessment(candidates,
		`{"items":[{"id":"A","score":90,"reason":"匹配"}]}`, 2)
	if ok || ranked != nil {
		t.Fatalf("incomplete assessment should be rejected: %+v", ranked)
	}
}

func TestProblemCandidateSearchTextIncludesStatementDetails(t *testing.T) {
	candidate := repository.AgentProblemCandidate{
		Name: "基础练习", InputFormat: "输入一个字符串", SampleExplain: "逐个统计字符出现次数",
	}
	if !candidateMatchesHomeworkKnowledge(candidate, &homeworkDraft{
		KnowledgeKeys: []string{"concept:string"}, TagNames: []string{"字符串基础"},
	}) {
		t.Fatal("input format and sample explanation should participate in deterministic matching")
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
