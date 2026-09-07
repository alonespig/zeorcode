package repository

import (
	"context"
	"time"

	"zoj/internal/model"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

type AgentRepo struct {
	db *gorm.DB
}

func NewAgentRepo(db *gorm.DB) *AgentRepo { return &AgentRepo{db: db} }

func (r *AgentRepo) ConversationPublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AgentConversation{}).Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *AgentRepo) RunPublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AgentRun{}).Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *AgentRepo) MessagePublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AgentMessage{}).Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *AgentRepo) ActionPublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AgentAction{}).Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *AgentRepo) CreateConversation(ctx context.Context, conversation *model.AgentConversation) error {
	return r.db.WithContext(ctx).Create(conversation).Error
}

func (r *AgentRepo) ListConversations(ctx context.Context, userID int64, keyword string, page, pageSize int) ([]model.AgentConversation, int64, error) {
	var list []model.AgentConversation
	var total int64
	db := r.db.WithContext(ctx).Model(&model.AgentConversation{}).
		Where("user_id = ? AND status = ?", userID, model.AgentConversationActive)
	if keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("updated_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *AgentRepo) GetConversation(ctx context.Context, publicID, userID int64) (*model.AgentConversation, error) {
	var conversation model.AgentConversation
	err := r.db.WithContext(ctx).Where("public_id = ? AND user_id = ?", publicID, userID).First(&conversation).Error
	return &conversation, err
}

func (r *AgentRepo) UpdateConversation(ctx context.Context, conversation *model.AgentConversation) error {
	return r.db.WithContext(ctx).Model(&model.AgentConversation{}).Where("id = ? AND user_id = ?", conversation.ID, conversation.UserID).
		Select("title", "state_json", "summary", "status", "updated_at").Updates(conversation).Error
}

func (r *AgentRepo) TouchConversation(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.AgentConversation{}).Where("id = ?", id).Update("updated_at", time.Now()).Error
}

func (r *AgentRepo) CreateMessage(ctx context.Context, message *model.AgentMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *AgentRepo) Messages(ctx context.Context, conversationID int64, limit int) ([]model.AgentMessage, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	var reversed []model.AgentMessage
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).
		Order("id DESC").Limit(limit).Find(&reversed).Error; err != nil {
		return nil, err
	}
	list := make([]model.AgentMessage, len(reversed))
	for i := range reversed {
		list[len(reversed)-1-i] = reversed[i]
	}
	return list, nil
}

func (r *AgentRepo) CreateRun(ctx context.Context, run *model.AgentRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *AgentRepo) FinishRun(ctx context.Context, id int64, status, errorCode string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AgentRun{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "error_code": errorCode, "finished_at": &now,
	}).Error
}

func (r *AgentRepo) CreateAction(ctx context.Context, action *model.AgentAction) error {
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *AgentRepo) GetAction(ctx context.Context, publicID, conversationID int64) (*model.AgentAction, error) {
	var action model.AgentAction
	err := r.db.WithContext(ctx).Where("public_id = ? AND conversation_id = ?", publicID, conversationID).First(&action).Error
	return &action, err
}

func (r *AgentRepo) BeginAction(ctx context.Context, actionID, approvedBy int64) (bool, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.AgentAction{}).
		Where("id = ? AND status = ?", actionID, model.AgentActionPending).
		Updates(map[string]any{"status": model.AgentActionExecuting, "approved_by": approvedBy, "approved_at": &now})
	return result.RowsAffected == 1, result.Error
}

func (r *AgentRepo) FinishAction(ctx context.Context, actionID int64, status string, resultPublicID int64, errorCode string) error {
	return r.db.WithContext(ctx).Model(&model.AgentAction{}).Where("id = ?", actionID).Updates(map[string]any{
		"status": status, "result_public_id": resultPublicID, "error_code": errorCode,
	}).Error
}

type AgentTeam struct {
	ID          int64  `gorm:"column:id"`
	PublicID    int64  `gorm:"column:public_id"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	MemberCount int    `gorm:"column:member_count"`
}

func (r *AgentRepo) ListManageableTeams(ctx context.Context, userID int64, isSiteAdmin bool) ([]AgentTeam, error) {
	var teams []AgentTeam
	db := r.db.WithContext(ctx).Table("teams").
		Select("teams.id, teams.public_id, teams.name, teams.description, COUNT(team_members.id) AS member_count").
		Joins("LEFT JOIN team_members ON team_members.team_id = teams.id")
	if !isSiteAdmin {
		db = db.Where("teams.id IN (?)", r.db.WithContext(ctx).Table("team_members").
			Select("team_id").Where("user_id = ? AND role >= ?", userID, model.TeamRoleAdmin))
	}
	err := db.Group("teams.id, teams.public_id, teams.name, teams.description").Order("teams.created_at DESC").Limit(100).Scan(&teams).Error
	return teams, err
}

type AgentProblemCandidate struct {
	ID            int64       `gorm:"column:id" json:"-"`
	DisplayID     string      `gorm:"column:display_id" json:"id"`
	Name          string      `gorm:"column:name" json:"name"`
	Difficulty    int         `gorm:"column:difficulty" json:"difficulty"`
	SubmitCount   int64       `gorm:"column:submit_count" json:"submitCount"`
	AcceptedCount int64       `gorm:"column:accepted_count" json:"acceptedCount"`
	Tags          []model.Tag `gorm:"-" json:"tags"`
}

func (r *AgentRepo) ListProblemCandidates(ctx context.Context, tagIDs []int64, difficulty int, excludedTeamID int64, limit int) ([]AgentProblemCandidate, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	var candidates []AgentProblemCandidate
	db := r.db.WithContext(ctx).Table("problems AS p").
		Select(`p.id, p.display_id, p.name, p.difficulty,
			COUNT(s.id) AS submit_count,
			COALESCE(SUM(CASE WHEN s.status = ? THEN 1 ELSE 0 END), 0) AS accepted_count`, judge.Accepted).
		Joins("LEFT JOIN submissions AS s ON s.problem_id = p.id").
		Where("p.hidden = ?", false)
	if difficulty > 0 {
		db = db.Where("p.difficulty = ?", difficulty)
	}
	if len(tagIDs) > 0 {
		tagged := r.db.WithContext(ctx).Table("problem_tags").Select("problem_id").Where("tag_id IN ?", tagIDs)
		db = db.Where("p.id IN (?)", tagged)
	}
	if excludedTeamID > 0 {
		used := r.db.WithContext(ctx).Table("homework_problems AS hp").Select("hp.problem_id").
			Joins("JOIN homeworks AS h ON h.id = hp.homework_id").Where("h.team_id = ?", excludedTeamID)
		db = db.Where("p.id NOT IN (?)", used)
	}
	if err := db.Group("p.id, p.display_id, p.name, p.difficulty").
		Order("p.difficulty ASC, p.id ASC").Limit(limit).Scan(&candidates).Error; err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return candidates, nil
	}
	ids := make([]int64, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.ID)
	}
	type tagRow struct {
		ProblemID int64  `gorm:"column:problem_id"`
		ID        int64  `gorm:"column:id"`
		Name      string `gorm:"column:name"`
	}
	var rows []tagRow
	if err := r.db.WithContext(ctx).Table("problem_tags AS pt").
		Select("pt.problem_id, tags.id, tags.name").Joins("JOIN tags ON tags.id = pt.tag_id").
		Where("pt.problem_id IN ?", ids).Order("tags.name ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	tagsByProblem := make(map[int64][]model.Tag)
	for _, row := range rows {
		tagsByProblem[row.ProblemID] = append(tagsByProblem[row.ProblemID], model.Tag{ID: row.ID, Name: row.Name})
	}
	for i := range candidates {
		candidates[i].Tags = tagsByProblem[candidates[i].ID]
	}
	return candidates, nil
}
