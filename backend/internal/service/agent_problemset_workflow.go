package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/repository"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"gorm.io/gorm"
)

type problemSetDraft struct {
	Version          int                                `json:"version"`
	TrainingType     string                             `json:"trainingType,omitempty"`
	TopicScope       string                             `json:"topicScope,omitempty"`
	KnowledgeKeys    []string                           `json:"knowledgeKeys,omitempty"`
	TagIDs           []int64                            `json:"tagIds"`
	TagNames         []string                           `json:"tagNames"`
	Requirements     string                             `json:"requirements,omitempty"`
	ProblemCount     int                                `json:"problemCount"`
	Difficulty       int                                `json:"difficulty"`
	Title            string                             `json:"title"`
	Description      string                             `json:"description"`
	Published        int                                `json:"published"`
	Visibility       int                                `json:"visibility"`
	InviteCode       string                             `json:"inviteCode,omitempty"`
	SelectedProblems []repository.AgentProblemCandidate `json:"selectedProblems"`
	ActionPublicID   int64                              `json:"actionPublicId,omitempty"`
}

type problemSetSettingsInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ProblemCount int    `json:"problemCount"`
	Published    int    `json:"published"`
	Visibility   int    `json:"visibility"`
	InviteCode   string `json:"inviteCode"`
}

func (d *problemSetDraft) selectionDraft() *homeworkDraft {
	return &homeworkDraft{
		TrainingType: d.TrainingType, TopicScope: d.TopicScope, KnowledgeKeys: append([]string(nil), d.KnowledgeKeys...),
		TagIDs: append([]int64(nil), d.TagIDs...), TagNames: append([]string(nil), d.TagNames...),
		Requirements: d.Requirements, ProblemCount: d.ProblemCount, Difficulty: d.Difficulty,
		SelectedProblems: append([]repository.AgentProblemCandidate(nil), d.SelectedProblems...),
	}
}

func (d *problemSetDraft) syncSelection(selection *homeworkDraft) {
	d.TrainingType = selection.TrainingType
	d.TopicScope = selection.TopicScope
	d.KnowledgeKeys = append(d.KnowledgeKeys[:0], selection.KnowledgeKeys...)
	d.TagIDs = append(d.TagIDs[:0], selection.TagIDs...)
	d.TagNames = append(d.TagNames[:0], selection.TagNames...)
	d.Requirements = selection.Requirements
	d.ProblemCount = selection.ProblemCount
	d.Difficulty = selection.Difficulty
	d.SelectedProblems = append(d.SelectedProblems[:0], selection.SelectedProblems...)
}

func (s *AgentService) startProblemSetWorkflow(state *agentConversationState, isSiteAdmin bool) (agentReply, error) {
	if !isSiteAdmin {
		return agentReply{Content: "创建题单需要管理员权限。你仍然可以让我帮助规划题单内容。"}, nil
	}
	state.ActiveWorkflow = agentWorkflowProblemSet
	state.Homework = nil
	state.ProblemSet = &problemSetDraft{Version: 1, ProblemCount: 8, Published: model.ProblemSetDraft, Visibility: model.ProblemSetPublic}
	return s.askProblemSetTrainingType(state), nil
}

func (s *AgentService) resumeProblemSetInteraction(ctx context.Context, conversation *model.AgentConversation,
	run *model.AgentRun, state *agentConversationState, isSiteAdmin bool, req *AgentTurnParams) (agentReply, error) {
	if !isSiteAdmin || state.ProblemSet == nil {
		return agentReply{}, errcode.ErrPermissionDenied.WithMsg("创建题单需要管理员权限")
	}
	draft := state.ProblemSet
	switch state.Step {
	case agentStepChooseTraining:
		var trainingType string
		if err := json.Unmarshal(req.Value, &trainingType); err != nil || !validHomeworkTrainingType(trainingType) {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("请选择有效的训练类型")
		}
		draft.TrainingType = trainingType
		draft.TopicScope, draft.Requirements = "", ""
		draft.KnowledgeKeys, draft.TagIDs, draft.TagNames = nil, nil, nil
		if trainingType == homeworkTrainingBeginner {
			return s.askProblemSetTags(ctx, state)
		}
		return s.askProblemSetScope(state), nil
	case agentStepChooseScope:
		var scope string
		if err := json.Unmarshal(req.Value, &scope); err != nil || !validHomeworkScope(draft.TrainingType, scope) {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("请选择有效的知识领域")
		}
		draft.TopicScope = scope
		draft.KnowledgeKeys, draft.TagIDs, draft.TagNames = nil, nil, nil
		return s.askProblemSetTags(ctx, state)
	case agentStepChooseTags:
		selection, err := decodeKnowledgeSelection(req.Value)
		if err != nil || len(selection) == 0 {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("请至少选择一个知识点")
		}
		if len(selection) > 12 {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("一次最多选择 12 个知识点")
		}
		selectionDraft := draft.selectionDraft()
		options, err := s.homeworkKnowledgeOptions(ctx, selectionDraft)
		if err != nil {
			return agentReply{}, errcode.ErrDatabase.Wrap(err)
		}
		allowed := make(map[string]homeworkKnowledgeOption, len(options))
		for _, option := range options {
			allowed[option.Key] = option
		}
		selectionDraft.TagIDs = selectionDraft.TagIDs[:0]
		selectionDraft.TagNames = selectionDraft.TagNames[:0]
		selectionDraft.KnowledgeKeys = selectionDraft.KnowledgeKeys[:0]
		seen := make(map[string]struct{}, len(selection))
		for _, key := range selection {
			option, ok := allowed[key]
			if !ok {
				customName, custom := decodeCustomKnowledgeKey(key)
				if !custom {
					return agentReply{}, errcode.ErrInvalidParams.WithMsg("包含不存在的知识点")
				}
				option = homeworkKnowledgeOption{Key: key, Name: customName}
			}
			normalizedName := strings.ToLower(strings.TrimSpace(option.Name))
			if _, exists := seen[normalizedName]; exists {
				continue
			}
			seen[normalizedName] = struct{}{}
			if option.TagID > 0 {
				selectionDraft.TagIDs = append(selectionDraft.TagIDs, option.TagID)
			}
			selectionDraft.KnowledgeKeys = append(selectionDraft.KnowledgeKeys, option.Key)
			selectionDraft.TagNames = append(selectionDraft.TagNames, option.Name)
		}
		if err := s.attachMatchingKnowledgeTags(ctx, selectionDraft); err != nil {
			return agentReply{}, errcode.ErrDatabase.Wrap(err)
		}
		if difficulty, automatic := automaticHomeworkDifficulty(selectionDraft); automatic {
			selectionDraft.Difficulty = difficulty
			draft.syncSelection(selectionDraft)
			return s.askProblemSetRequirements(state), nil
		}
		draft.syncSelection(selectionDraft)
		return s.askProblemSetDifficulty(state), nil
	case agentStepChooseDifficulty:
		difficulty, err := decodeInt(req.Value)
		if err != nil || difficulty < 0 || difficulty > 3 {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("请选择有效的题目难度")
		}
		draft.Difficulty = difficulty
		return s.askProblemSetRequirements(state), nil
	case agentStepRequirements:
		var requirements string
		if err := json.Unmarshal(req.Value, &requirements); err != nil {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("补充要求格式错误")
		}
		requirements = strings.TrimSpace(requirements)
		if utf8.RuneCountInString(requirements) > 500 {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("补充要求不能超过 500 个字符")
		}
		selectionDraft := draft.selectionDraft()
		selectionDraft.Requirements = requirements
		if err := s.applyRequirementTagHints(ctx, selectionDraft); err != nil {
			return agentReply{}, errcode.ErrDatabase.Wrap(err)
		}
		draft.syncSelection(selectionDraft)
		return s.askProblemSetSettings(state), nil
	case agentStepProblemSetSettings:
		var input problemSetSettingsInput
		if err := json.Unmarshal(req.Value, &input); err != nil {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("题单设置格式错误").Wrap(err)
		}
		if err := applyProblemSetSettings(draft, &input); err != nil {
			return agentReply{}, err
		}
		return s.prepareProblemSetCandidates(ctx, conversation, run, state)
	case agentStepShortage:
		var choice string
		if err := json.Unmarshal(req.Value, &choice); err != nil {
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("请选择处理方式")
		}
		switch choice {
		case "reduce_count":
			if len(draft.SelectedProblems) == 0 {
				return agentReply{}, errcode.ErrInvalidParams.WithMsg("当前没有可用题目，请调整训练类型、难度或知识点")
			}
			draft.ProblemCount = len(draft.SelectedProblems)
			return s.buildProblemSetPreview(ctx, conversation, run, state)
		case "change_tags":
			return s.askProblemSetTags(ctx, state)
		case "change_preferences":
			return s.askProblemSetTrainingType(state), nil
		default:
			return agentReply{}, errcode.ErrInvalidParams.WithMsg("不支持的处理方式")
		}
	default:
		return agentReply{}, errcode.ErrAgentInteractionExpired
	}
}

func (s *AgentService) askProblemSetTrainingType(state *agentConversationState) agentReply {
	requestID := newAgentRequestID()
	state.Step, state.PendingRequestID = agentStepChooseTraining, requestID
	options := []AgentOption{
		{Label: "新手入门", Value: homeworkTrainingBeginner, Description: "基础语法、简单模拟和编程习惯"},
		{Label: "数据结构", Value: homeworkTrainingDataStructure, Description: "栈、队列、树、图以及常用高级数据结构"},
		{Label: "算法竞赛", Value: homeworkTrainingAlgorithm, Description: "搜索、图论、动态规划、字符串和数学专题", Recommended: true},
		{Label: "综合练习", Value: homeworkTrainingComprehensive, Description: "跨多个领域组织阶段性练习"},
	}
	block := AgentBlock{Type: "single_select", RequestID: requestID, Title: "选择题单类型",
		Description: "先确定题单面向的学习阶段，后面再逐步缩小知识范围。", Required: true, Options: options}
	return agentReply{Content: "可以。我先了解这个题单用于哪类训练。", Blocks: []AgentBlock{block}, Waiting: true}
}

func (s *AgentService) askProblemSetScope(state *agentConversationState) agentReply {
	requestID := newAgentRequestID()
	state.Step, state.PendingRequestID = agentStepChooseScope, requestID
	options := homeworkScopeOptions(state.ProblemSet.TrainingType)
	block := AgentBlock{Type: "single_select", RequestID: requestID, Title: "选择知识领域",
		Description: "先确定大方向，下一步再选择具体知识点。", Required: true, Options: options}
	return agentReply{Content: fmt.Sprintf("已确定为%s题单。接下来选择知识领域。", homeworkTrainingTypeLabel(state.ProblemSet.TrainingType)),
		Blocks: []AgentBlock{block}, Waiting: true}
}

func (s *AgentService) askProblemSetTags(ctx context.Context, state *agentConversationState) (agentReply, error) {
	draft := state.ProblemSet.selectionDraft()
	knowledge, err := s.homeworkKnowledgeOptions(ctx, draft)
	if err != nil {
		return agentReply{}, errcode.ErrDatabase.Wrap(err)
	}
	if len(knowledge) == 0 {
		return agentReply{}, errcode.ErrOperationDenied.WithMsg("当前知识领域还没有可用知识点，请先补充题目标签")
	}
	requestID := newAgentRequestID()
	state.Step, state.PendingRequestID = agentStepChooseTags, requestID
	options := make([]AgentOption, 0, len(knowledge))
	for _, item := range knowledge {
		options = append(options, AgentOption{Label: item.Name, Value: item.Key, Description: item.Description})
	}
	payload := map[string]any{"allowCustomInput": true, "customInputPlaceholder": "输入其他知识点，例如单调栈、状态压缩",
		"customInputMaxLength": 30, "customInputLimit": 8}
	block := AgentBlock{Type: "multi_select", RequestID: requestID, Title: "选择具体知识点",
		Description: "可以选择系统知识点，也可以自行补充。后续会结合题面检查实际解法。", Required: true, Options: options, Payload: payload}
	return agentReply{Content: "现在选择题单需要覆盖的具体知识点。", Blocks: []AgentBlock{block}, Waiting: true}, nil
}

func (s *AgentService) askProblemSetDifficulty(state *agentConversationState) agentReply {
	requestID := newAgentRequestID()
	state.Step, state.PendingRequestID = agentStepChooseDifficulty, requestID
	recommended := recommendedHomeworkDifficulty(state.ProblemSet.TrainingType)
	options := []AgentOption{
		{Label: "简单", Value: 1, Description: "适合入门和基础巩固", Recommended: recommended == 1},
		{Label: "中等", Value: 2, Description: "需要综合运用知识点", Recommended: recommended == 2},
		{Label: "困难", Value: 3, Description: "适合提高和挑战训练", Recommended: recommended == 3},
		{Label: "不限难度", Value: 0, Description: "组合不同难度形成练习梯度", Recommended: recommended == 0},
	}
	block := AgentBlock{Type: "single_select", RequestID: requestID, Title: "选择题目难度",
		Description: "内容匹配后，再按该难度组织候选题。", Required: true, Options: options}
	return agentReply{Content: "知识点已经确定，接下来选择题单整体难度。", Blocks: []AgentBlock{block}, Waiting: true}
}

func (s *AgentService) askProblemSetRequirements(state *agentConversationState) agentReply {
	requestID := newAgentRequestID()
	state.Step, state.PendingRequestID = agentStepRequirements, requestID
	payload := map[string]any{"placeholder": "例如：从模板题逐步过渡到综合题；不要网络流；偏重思维训练。", "maxLength": 500}
	block := AgentBlock{Type: "text_input", RequestID: requestID, Title: "补充选题要求（可选）",
		Description: "可以说明题目风格、希望避免的知识点或难度梯度。", Payload: payload}
	content := "如果有更具体的选题要求，可以继续补充；没有可直接跳过。"
	if state.ProblemSet.TrainingType == homeworkTrainingBeginner {
		content = "已为新手入门题单采用简单难度。如果还有选题要求，可以继续补充。"
	}
	return agentReply{Content: content, Blocks: []AgentBlock{block}, Waiting: true}
}

func (s *AgentService) askProblemSetSettings(state *agentConversationState) agentReply {
	requestID := newAgentRequestID()
	state.Step, state.PendingRequestID = agentStepProblemSetSettings, requestID
	draft := state.ProblemSet
	selection := draft.selectionDraft()
	defaultTitle := homeworkTitleTopic(selection) + "专项题单"
	payload := map[string]any{
		"defaultTitle": defaultTitle, "defaultProblemCount": 8, "defaultPublished": model.ProblemSetDraft,
		"defaultVisibility": model.ProblemSetPublic, "trainingType": homeworkTrainingTypeLabel(draft.TrainingType),
		"difficulty": homeworkDifficultyLabel(draft.Difficulty), "knowledgePoints": draft.TagNames, "requirements": draft.Requirements,
	}
	block := AgentBlock{Type: "problemset_settings", RequestID: requestID, Title: "完善题单设置",
		Description: "设置题单名称、题量和发布方式。默认先保存为草稿。", Required: true, Payload: payload}
	return agentReply{Content: "选题范围已经确定，最后完善题单信息。", Blocks: []AgentBlock{block}, Waiting: true}
}

func applyProblemSetSettings(draft *problemSetDraft, input *problemSetSettingsInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.InviteCode = strings.TrimSpace(input.InviteCode)
	if input.Title == "" || utf8.RuneCountInString(input.Title) > 100 {
		return errcode.ErrInvalidParams.WithMsg("题单名称不能为空且不能超过 100 个字符")
	}
	if utf8.RuneCountInString(input.Description) > 5000 {
		return errcode.ErrInvalidParams.WithMsg("题单说明不能超过 5000 个字符")
	}
	if input.ProblemCount < 1 || input.ProblemCount > 30 {
		return errcode.ErrInvalidParams.WithMsg("题目数量必须在 1 到 30 之间")
	}
	if input.Published != model.ProblemSetDraft && input.Published != model.ProblemSetPublished {
		return errcode.ErrInvalidParams.WithMsg("发布状态无效")
	}
	if input.Visibility != model.ProblemSetPublic && input.Visibility != model.ProblemSetInviteOnly {
		return errcode.ErrInvalidParams.WithMsg("可见性设置无效")
	}
	if utf8.RuneCountInString(input.InviteCode) > 32 {
		return errcode.ErrInvalidParams.WithMsg("邀请码不能超过 32 个字符")
	}
	if input.Visibility == model.ProblemSetInviteOnly && input.InviteCode == "" {
		return errcode.ErrInvalidParams.WithMsg("邀请码可见的题单必须填写邀请码")
	}
	draft.Title, draft.Description, draft.ProblemCount = input.Title, input.Description, input.ProblemCount
	draft.Published, draft.Visibility, draft.InviteCode = input.Published, input.Visibility, input.InviteCode
	return nil
}

func (s *AgentService) prepareProblemSetCandidates(ctx context.Context, conversation *model.AgentConversation,
	run *model.AgentRun, state *agentConversationState) (agentReply, error) {
	draft := state.ProblemSet
	selection := draft.selectionDraft()
	candidates, err := s.repo.ListProblemCandidates(ctx, draft.TagIDs, draft.Difficulty, 0, 100)
	if err != nil {
		return agentReply{}, errcode.ErrDatabase.Wrap(err)
	}
	candidates = filterHomeworkCandidates(candidates, selection)
	sortCandidates(candidates, draft.TagIDs)
	candidates = homeworkCandidateShortlist(candidates, draft.ProblemCount)
	if ranked, applied, rankErr := s.rankHomeworkCandidatesByContent(ctx, candidates, selection); rankErr != nil {
		logger.Warnw("agent problemset content assessment failed", "conversationID", conversation.ID, "err", rankErr)
	} else if applied {
		candidates = ranked
	}
	for index := range candidates {
		if candidates[index].SelectionReason == "" {
			candidates[index].SelectionReason = fallbackHomeworkSelectionReason(candidates[index], selection)
		}
	}
	selectedCount := draft.ProblemCount
	if selectedCount > len(candidates) {
		selectedCount = len(candidates)
	}
	draft.SelectedProblems = append(draft.SelectedProblems[:0], candidates[:selectedCount]...)
	if len(candidates) < draft.ProblemCount {
		requestID := newAgentRequestID()
		state.Step, state.PendingRequestID = agentStepShortage, requestID
		options := []AgentOption{{Label: fmt.Sprintf("改为 %d 题", len(candidates)), Value: "reduce_count", Disabled: len(candidates) == 0}}
		if len(draft.TagNames) > 0 {
			options = append(options, AgentOption{Label: "重新选择知识点", Value: "change_tags"})
		}
		options = append(options, AgentOption{Label: "重新选择训练方向", Value: "change_preferences"})
		block := AgentBlock{Type: "notice", RequestID: requestID, Title: "符合条件的题目不足",
			Description: fmt.Sprintf("需要 %d 题，阅读题面后找到 %d 道匹配题。", draft.ProblemCount, len(candidates)), Required: true,
			Options: options, Payload: map[string]any{"requested": draft.ProblemCount, "available": len(candidates), "tagNames": draft.TagNames}}
		return agentReply{Content: "按题面内容复核后，匹配题目不足。你可以缩减题量或调整范围。",
			Blocks: []AgentBlock{block}, Waiting: true}, nil
	}
	return s.buildProblemSetPreview(ctx, conversation, run, state)
}

func (s *AgentService) buildProblemSetPreview(ctx context.Context, conversation *model.AgentConversation,
	run *model.AgentRun, state *agentConversationState) (agentReply, error) {
	draft := state.ProblemSet
	draft.Version++
	draft.ActionPublicID = 0
	artifactJSON, err := json.Marshal(draft)
	if err != nil {
		return agentReply{}, errcode.ErrInternal.Wrap(err)
	}
	hashBytes := sha256.Sum256(artifactJSON)
	actionPublicID, err := newUniquePublicID(ctx, s.repo.ActionPublicIDExists)
	if err != nil {
		return agentReply{}, err
	}
	action := &model.AgentAction{PublicID: actionPublicID, ConversationID: conversation.ID, RunID: run.ID,
		ActionType: agentWorkflowProblemSet, ArtifactJSON: string(artifactJSON), ArtifactVersion: draft.Version,
		PayloadHash: hex.EncodeToString(hashBytes[:]), Status: model.AgentActionPending}
	if err := s.repo.CreateAction(ctx, action); err != nil {
		return agentReply{}, errcode.ErrDatabase.Wrap(err)
	}
	draft.ActionPublicID = action.PublicID
	state.Step, state.PendingRequestID = agentStepConfirm, 0
	table := agentProblemSelectionTable(draft.SelectedProblems)
	previewPayload := map[string]any{
		"action": agentWorkflowProblemSet, "actionId": action.PublicID, "draftVersion": draft.Version,
		"title": draft.Title, "description": draft.Description, "problemCount": len(draft.SelectedProblems),
		"trainingType": homeworkTrainingTypeLabel(draft.TrainingType), "difficulty": homeworkDifficultyLabel(draft.Difficulty),
		"knowledgeTags": draft.TagNames, "requirements": draft.Requirements, "published": draft.Published,
		"visibility": draft.Visibility, "hasInviteCode": draft.InviteCode != "",
	}
	preview := AgentBlock{Type: "action_preview", Title: "创建题单", Description: "请确认下面的信息。确认后才会真正创建题单。", Payload: previewPayload}
	return agentReply{Content: "题单草稿已经准备好，请检查题目和发布方式后确认创建。",
		Blocks: []AgentBlock{table, preview}, Waiting: true}, nil
}

func (s *AgentService) approveProblemSetAction(ctx context.Context, conversation *model.AgentConversation,
	state *agentConversationState, userID int64, isSiteAdmin bool, req *AgentTurnParams) (agentReply, error) {
	if !isSiteAdmin || state.Step != agentStepConfirm || state.ProblemSet == nil || req.ActionID == 0 ||
		req.ActionID != state.ProblemSet.ActionPublicID || req.DraftVersion != state.ProblemSet.Version {
		return agentReply{}, errcode.ErrAgentActionConflict
	}
	action, err := s.repo.GetAction(ctx, req.ActionID, conversation.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return agentReply{}, errcode.ErrAgentActionConflict
		}
		return agentReply{}, errcode.ErrDatabase.Wrap(err)
	}
	if action.ActionType != agentWorkflowProblemSet {
		return agentReply{}, errcode.ErrAgentActionConflict
	}
	if action.Status == model.AgentActionSucceeded {
		return problemSetSuccessReply(state.ProblemSet, action.ResultPublicID), nil
	}
	if action.ArtifactVersion != req.DraftVersion {
		return agentReply{}, errcode.ErrAgentActionConflict
	}
	artifactHash := sha256.Sum256([]byte(action.ArtifactJSON))
	if !strings.EqualFold(action.PayloadHash, hex.EncodeToString(artifactHash[:])) {
		return agentReply{}, errcode.ErrAgentActionConflict.WithMsg("操作草稿校验失败，请重新生成")
	}
	started, err := s.repo.BeginAction(ctx, action.ID, userID)
	if err != nil {
		return agentReply{}, errcode.ErrDatabase.Wrap(err)
	}
	if !started {
		return agentReply{}, errcode.ErrAgentActionConflict
	}
	var draft problemSetDraft
	if err := json.Unmarshal([]byte(action.ArtifactJSON), &draft); err != nil {
		_ = s.repo.FinishAction(ctx, action.ID, model.AgentActionFailed, 0, "ARTIFACT_INVALID")
		return agentReply{}, errcode.ErrInternal.Wrap(err)
	}
	problems := make([]string, 0, len(draft.SelectedProblems))
	for _, problem := range draft.SelectedProblems {
		problems = append(problems, problem.DisplayID)
	}
	params := SaveProblemSetParams{Title: draft.Title, Description: draft.Description, Published: draft.Published,
		Visibility: draft.Visibility, InviteCode: draft.InviteCode, TagIDs: draft.TagIDs, Problems: problems}
	problemSetPublicID, err := s.problemSetSrv.Save(ctx, 0, params, userID)
	if err != nil {
		_ = s.repo.FinishAction(ctx, action.ID, model.AgentActionFailed, 0, appErrorCode(err))
		return agentReply{}, err
	}
	if err := s.repo.FinishAction(ctx, action.ID, model.AgentActionSucceeded, problemSetPublicID, ""); err != nil {
		return agentReply{}, errcode.ErrDatabase.Wrap(err)
	}
	reply := problemSetSuccessReply(&draft, problemSetPublicID)
	*state = agentConversationState{}
	return reply, nil
}

func problemSetSuccessReply(draft *problemSetDraft, problemSetPublicID int64) agentReply {
	link := fmt.Sprintf("/problemset/%d", problemSetPublicID)
	block := AgentBlock{Type: "action_result", Title: "题单创建成功", Payload: map[string]any{
		"action": agentWorkflowProblemSet, "problemSetId": problemSetPublicID, "title": draft.Title, "link": link,
	}}
	return agentReply{Content: fmt.Sprintf("题单“%s”已经创建完成。", draft.Title), Blocks: []AgentBlock{block}}
}

func isProblemSetIntent(content string) bool {
	content = strings.ToLower(strings.Join(strings.Fields(content), ""))
	hasObject := containsAgentIntentPhrase(content, "题单", "提单", "题目单", "刷题单", "problemset")
	hasAction := containsAgentIntentPhrase(content, "创建", "生成", "新建", "整理", "制作", "出一", "建一", "做一", "来一", "弄一")
	return hasObject && hasAction
}

func containsAgentIntentPhrase(content string, phrases ...string) bool {
	for _, phrase := range phrases {
		if strings.Contains(content, phrase) {
			return true
		}
	}
	return false
}
