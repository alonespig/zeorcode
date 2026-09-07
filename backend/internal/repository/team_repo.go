package repository

import (
	"context"
	"time"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type TeamRepo struct {
	db *gorm.DB
}

func NewTeamRepo(db *gorm.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

// Transaction 由 service 定义事务边界，并为事务内的仓储重新绑定 tx。
func (r *TeamRepo) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// TeamQuery 团队列表查询条件。MemberOf 非 0 时只返回该用户加入的团队。
type TeamQuery struct {
	Page       int
	PageSize   int
	Keyword    string
	MemberOf   int64
	Visibility *int
}

// TeamCount 某团队的成员数/作业数聚合行
type TeamCount struct {
	TeamID int64 `gorm:"column:team_id"`
	Count  int   `gorm:"column:cnt"`
}

func (r *TeamRepo) List(ctx context.Context, q *TeamQuery) ([]model.Team, int64, error) {
	var teams []model.Team
	var total int64
	db := r.db.WithContext(ctx).Model(&model.Team{})

	if q.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+q.Keyword+"%")
	}
	if q.Visibility != nil {
		db = db.Where("visibility = ?", *q.Visibility)
	}
	if q.MemberOf != 0 {
		sub := r.db.WithContext(ctx).Table("team_members").
			Select("team_id").Where("user_id = ?", q.MemberOf)
		db = db.Where("id IN (?)", sub)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}
	return teams, total, nil
}

func (r *TeamRepo) GetByID(ctx context.Context, id int64) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&team).Error
	return &team, err
}

func (r *TeamRepo) PublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Team{}).
		Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *TeamRepo) ResolveIDByPublicID(ctx context.Context, publicID int64) (int64, error) {
	var team model.Team
	err := r.db.WithContext(ctx).Select("id").
		Where("public_id = ?", publicID).First(&team).Error
	return team.ID, err
}

// MemberCountsByTeamIDs 批量取成员数，列表页用，避免逐个团队查一次。
func (r *TeamRepo) MemberCountsByTeamIDs(ctx context.Context, teamIDs []int64) ([]TeamCount, error) {
	if len(teamIDs) == 0 {
		return nil, nil
	}
	var counts []TeamCount
	err := r.db.WithContext(ctx).Table("team_members").
		Select("team_id, COUNT(*) AS cnt").
		Where("team_id IN ?", teamIDs).
		Group("team_id").
		Find(&counts).Error
	return counts, err
}

// HomeworkCount 某团队的作业数
func (r *TeamRepo) HomeworkCount(ctx context.Context, teamID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Homework{}).
		Where("team_id = ?", teamID).Count(&n).Error
	return n, err
}

// MyRolesByTeamIDs 批量取当前用户在这些团队中的角色，列表页用。
func (r *TeamRepo) MyRolesByTeamIDs(ctx context.Context, userID int64, teamIDs []int64) (map[int64]int, error) {
	if userID == 0 || len(teamIDs) == 0 {
		return map[int64]int{}, nil
	}
	var rows []model.TeamMember
	err := r.db.WithContext(ctx).Model(&model.TeamMember{}).
		Select("team_id, role").
		Where("user_id = ? AND team_id IN ?", userID, teamIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	roles := make(map[int64]int, len(rows))
	for _, row := range rows {
		roles[row.TeamID] = row.Role
	}
	return roles, nil
}

// GetMember 取某人在某团队的成员记录；不是成员返回 gorm.ErrRecordNotFound。
func (r *TeamRepo) GetMember(ctx context.Context, teamID, userID int64) (*model.TeamMember, error) {
	var m model.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).First(&m).Error
	return &m, err
}

// ListMembers 团队成员列表，按角色降序（所有者→管理员→成员）再按加入时间。
func (r *TeamRepo) ListMembers(ctx context.Context, teamID int64) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("role DESC, joined_at ASC").
		Find(&members).Error
	return members, err
}

// CreateWithOwner 建团队并把创建者写成所有者，同一事务。
func (r *TeamRepo) CreateWithOwner(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		owner := model.TeamMember{
			TeamID:   team.ID,
			UserID:   team.OwnerID,
			Role:     model.TeamRoleOwner,
			JoinedAt: time.Now(),
		}
		return tx.Create(&owner).Error
	})
}

// Update 只更新可变字段。用 Select 显式列出：Updates 默认忽略零值，
// 而「改回公开(visibility=0)」「清空邀请码」都是要写零值的。
func (r *TeamRepo) Update(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Model(&model.Team{}).
		Where("id = ?", team.ID).
		Select("name", "cover_url", "description", "visibility", "invite_code").
		Updates(team).Error
}

// Delete 解散团队，连带清掉成员、作业、作业题目关联，避免留孤儿行。
// 提交记录保留：那是用户自己的评测历史，不该因为团队解散而消失。
func (r *TeamRepo) Delete(ctx context.Context, teamID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var hwIDs []int64
		if err := tx.Model(&model.Homework{}).Where("team_id = ?", teamID).
			Pluck("id", &hwIDs).Error; err != nil {
			return err
		}
		if len(hwIDs) > 0 {
			if err := tx.Where("homework_id IN ?", hwIDs).
				Delete(&model.HomeworkProblem{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("team_id = ?", teamID).Delete(&model.Homework{}).Error; err != nil {
			return err
		}
		if err := tx.Where("team_id = ?", teamID).Delete(&model.TeamMember{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", teamID).Delete(&model.Team{}).Error
	})
}

// AddMember 入队。并发重复加入会撞唯一键，这里当成已加入忽略。
func (r *TeamRepo) AddMember(ctx context.Context, teamID, userID int64) error {
	m := model.TeamMember{
		TeamID:   teamID,
		UserID:   userID,
		Role:     model.TeamRoleMember,
		JoinedAt: time.Now(),
	}
	return r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		FirstOrCreate(&m).Error
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&model.TeamMember{}).Error
}

func (r *TeamRepo) SetMemberRole(ctx context.Context, teamID, userID int64, role int) error {
	return r.db.WithContext(ctx).Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role).Error
}
