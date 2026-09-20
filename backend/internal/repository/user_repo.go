package repository

import (
	"context"
	"time"

	"zoj/internal/model"

	"gorm.io/gorm"
)

// RatingHistoryRow rating 历史一行（join 比赛名）
type RatingHistoryRow struct {
	ContestID   int64
	ContestName string
	Rank        int
	OldRating   int
	NewRating   int
	Delta       int
	CreatedAt   time.Time
}

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) GetUser(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetUserByLogin 登录用：按用户名或邮箱匹配（邮箱唯一，登录时二选一）
func (r *UserRepo) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ? OR email = ?", login, login).First(&user).Error
	return &user, err
}

// GetUserByUID 按对外用户号取用户。
func (r *UserRepo) GetUserByUID(ctx context.Context, uid int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
	return &user, err
}

// ResolveIDByUID 把对外用户号解析成内部主键；找不到返回 gorm.ErrRecordNotFound。
func (r *UserRepo) ResolveIDByUID(ctx context.Context, uid int64) (int64, error) {
	var id int64
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("uid = ?", uid).Limit(1).Pluck("id", &id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return id, nil
}

// ExistsByUID uid 查重（生成随机 uid 时防碰撞）。
func (r *UserRepo) ExistsByUID(ctx context.Context, uid int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("uid = ?", uid).Count(&count).Error
	return count > 0, err
}

// ExistsByUsername 注册查重：用户名是否已被占用
func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 注册查重：邮箱是否已被占用（仅对非空邮箱有意义）
func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByStudentNo 学号查重；空学号由业务层表示为 NULL，不调用此方法。
func (r *UserRepo) ExistsByStudentNo(ctx context.Context, studentNo string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("student_no = ?", studentNo).Count(&count).Error
	return count > 0, err
}

type UserPassCount struct {
	UserID    int64 `gorm:"column:user_id" json:"userID"`
	PassCount int64 `gorm:"column:count" json:"count"`
}

// GetRatingHistory 某用户的 rating 变化历史（按时间升序，含比赛名）。
func (r *UserRepo) GetRatingHistory(ctx context.Context, userID int64) ([]RatingHistoryRow, error) {
	var rows []RatingHistoryRow
	err := r.db.WithContext(ctx).
		Table("rating_changes rc").
		Select("c.public_id AS contest_id, c.name AS contest_name, rc.`rank` AS `rank`, rc.old_rating, rc.new_rating, rc.delta, rc.created_at").
		Joins("JOIN contests c ON c.id = rc.contest_id").
		Where("rc.user_id = ?", userID).
		Order("rc.created_at ASC").
		Scan(&rows).Error
	return rows, err
}

// ContestHistoryRow 用户参赛历史一行（join 比赛，LEFT JOIN 可空的 rating 变化）
type ContestHistoryRow struct {
	ContestID   int64
	ContestName string
	Type        int
	StartTime   time.Time
	EndTime     time.Time
	Rated       bool
	Settled     bool
	Rank        *int // 以下四项来自 rating_changes，未结算时为 NULL
	OldRating   *int
	NewRating   *int
	Delta       *int
	Total       int // 该场已结算人数（名次分母），未结算为 0
}

// GetContestHistory 用户报名过的全部比赛 + 若已结算则带上 rating 变化（个人页「比赛记录」用），按开始时间倒序。
func (r *UserRepo) GetContestHistory(ctx context.Context, userID int64) ([]ContestHistoryRow, error) {
	var rows []ContestHistoryRow
	err := r.db.WithContext(ctx).
		Table("contest_users cu").
		Select("c.public_id AS contest_id, c.name AS contest_name, c.type, c.start_time, c.end_time, c.rated, c.settled, "+
			"rc.`rank` AS `rank`, rc.old_rating, rc.new_rating, rc.delta, "+
			"(SELECT COUNT(*) FROM rating_changes rc2 WHERE rc2.contest_id = c.id) AS total").
		Joins("JOIN contests c ON c.id = cu.contest_id").
		Joins("LEFT JOIN rating_changes rc ON rc.contest_id = c.id AND rc.user_id = cu.user_id").
		Where("cu.user_id = ?", userID).
		Order("c.start_time DESC").
		Scan(&rows).Error
	return rows, err
}

// AllUserIDs 取全部用户 id（系统通知广播扇出用）。
func (r *UserRepo) AllUserIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Pluck("id", &ids).Error
	return ids, err
}

// RatingRankList 按 rating 降序分页取用户（rating 榜）。
func (r *UserRepo) RatingRankList(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("rating DESC, id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepo) UserRankList(ctx context.Context, page, pageSize int) ([]UserPassCount, int64, error) {
	var list []UserPassCount
	var total int64
	// total 统计有提交记录的用户数，与排名数据源一致
	err := r.db.WithContext(ctx).Model(&model.Submission{}).
		Where("contest_id = ?", 0).
		Distinct("user_id").
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.WithContext(ctx).Model(&model.Submission{}).
		Select("user_id, COUNT(DISTINCT problem_id) as count").
		Where("contest_id = ?", 0).
		Group("user_id").
		Order("count desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *UserRepo) GetUserNameByID(ctx context.Context, id int64) (string, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user.Username, err
}

func (r *UserRepo) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdatePassword 只更新密码字段，避免用 Save 覆盖并发修改的其他资料。
func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("password", passwordHash).Error
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *UserRepo) FindByIDs(ctx context.Context, ids []int64) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// FindByEmails 按邮箱批量查找用户。调用方应先把邮箱统一规范为小写。
func (r *UserRepo) FindByEmails(ctx context.Context, emails []string) ([]model.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}
	var users []model.User
	err := r.db.WithContext(ctx).Where("email IN ?", emails).Find(&users).Error
	return users, err
}

// FindByUsernames 按用户名批量查找用户。
func (r *UserRepo) FindByUsernames(ctx context.Context, usernames []string) ([]model.User, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var users []model.User
	err := r.db.WithContext(ctx).Where("username IN ?", usernames).Find(&users).Error
	return users, err
}

// ListUsers 分页列出所有用户（管理员后台用）
func (r *UserRepo) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.WithContext(ctx).Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// CreateUsersBatch 批量插入，包在事务里：任一失败则全部回滚
func (r *UserRepo) CreateUsersBatch(ctx context.Context, users []*model.User) error {
	if len(users) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(users).Error
}

func (r *UserRepo) SearchByUsername(ctx context.Context, name string) ([]model.User, error) {
	var userList []model.User

	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("username LIKE ?", "%"+name+"%").
		Find(&userList).Error; err != nil {
		return userList, err
	}

	return userList, nil
}

// CountByUsernames 用于批量创建前的预检：统计已存在的用户名
func (r *UserRepo) CountByUsernames(ctx context.Context, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var existing []string
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("username IN ?", names).
		Pluck("username", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

// CountByEmails 用于批量创建前一次性查出已占用邮箱。
func (r *UserRepo) CountByEmails(ctx context.Context, emails []string) ([]string, error) {
	if len(emails) == 0 {
		return nil, nil
	}
	var existing []string
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("email IN ?", emails).
		Pluck("email", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

// CountByStudentNos 用于批量创建前一次性查出已占用学号。
func (r *UserRepo) CountByStudentNos(ctx context.Context, studentNos []string) ([]string, error) {
	if len(studentNos) == 0 {
		return nil, nil
	}
	var existing []string
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("student_no IN ?", studentNos).
		Pluck("student_no", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}
