package repository

import (
	"context"
	"errors"

	"zoj/internal/model"
	"zoj/pkg/judge"

	"gorm.io/gorm"
)

type ProblemRepo struct {
	db *gorm.DB
}

func NewProblemRepo(db *gorm.DB) *ProblemRepo {
	return &ProblemRepo{db: db}
}

// UpsertUserProblem 重算用：按 (user,problem) 找到就更新，找不到就插入。
// 注意 user_problem 上 (user_id,problem_id) 只是普通索引（非唯一），故不用 OnConflict，走 find-then-save。
func (r *ProblemRepo) UpsertUserProblem(ctx context.Context, up *model.UserProblem) error {
	var existing model.UserProblem
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND problem_id = ?", up.UserID, up.ProblemID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(up).Error
	}
	if err != nil {
		return err
	}
	existing.Status = up.Status
	existing.AcCount = up.AcCount
	existing.SubmitCount = up.SubmitCount
	return r.db.WithContext(ctx).Save(&existing).Error
}

// DeleteUserProblem 重算用：某用户某题已无任何有效提交时，删掉残留聚合行。
func (r *ProblemRepo) DeleteUserProblem(ctx context.Context, userID, problemID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND problem_id = ?", userID, problemID).
		Delete(&model.UserProblem{}).Error
}

func (r *ProblemRepo) Create(ctx context.Context, po model.Problem) (int64, error) {
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return 0, err
	}
	return po.ID, nil
}

type ProblemQuery struct {
	Keyword       string
	Difficulty    *int    // 按难度筛选（1 简单 / 2 中等 / 3 困难）
	TagIDs        []int64 // 按标签筛选（AND：题目须同时含所有选中标签）
	Page          int
	PageSize      int
	Order         string
	IncludeHidden bool // true=含隐藏题（管理员）；false=只返回公开题
}

func (r *ProblemRepo) List(ctx context.Context, q *ProblemQuery) ([]model.Problem, int64, error) {
	var problems []model.Problem
	var total int64
	db := r.db.WithContext(ctx).Model(&model.Problem{})

	if !q.IncludeHidden {
		db = db.Where("hidden = ?", false)
	}
	if q.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+q.Keyword+"%")
	}
	if q.Difficulty != nil {
		db = db.Where("difficulty = ?", *q.Difficulty)
	}
	if len(q.TagIDs) > 0 {
		// AND 交集：题目须同时含全部选中标签 —— 命中标签数 == 选中标签数
		sub := r.db.WithContext(ctx).Table("problem_tags").
			Select("problem_id").
			Where("tag_id IN ?", q.TagIDs).
			Group("problem_id").
			Having("COUNT(DISTINCT tag_id) = ?", len(q.TagIDs))
		db = db.Where("id IN (?)", sub)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize)
	if q.Order == "latest" {
		tx = tx.Order("id DESC")
	}
	if err := tx.Find(&problems).Error; err != nil {
		return nil, 0, err
	}
	return problems, total, nil
}

// GetByDisplayID 按对外题号取题目。
func (r *ProblemRepo) GetByDisplayID(ctx context.Context, displayID string) (*model.Problem, error) {
	var p model.Problem
	err := r.db.WithContext(ctx).Where("display_id = ?", displayID).First(&p).Error
	return &p, err
}

// ResolveID 把对外题号解析成内部主键；找不到返回 gorm.ErrRecordNotFound。
func (r *ProblemRepo) ResolveID(ctx context.Context, displayID string) (int64, error) {
	var id int64
	err := r.db.WithContext(ctx).Model(&model.Problem{}).
		Where("display_id = ?", displayID).Limit(1).Pluck("id", &id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return id, nil
}

// ExistsByDisplayID 题号查重（excludeID 排除自身，用于编辑）。
func (r *ProblemRepo) ExistsByDisplayID(ctx context.Context, displayID string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Problem{}).
		Where("display_id = ? AND id <> ?", displayID, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *ProblemRepo) GetByID(ctx context.Context, id int64) (*model.Problem, error) {
	var problem model.Problem
	err := r.db.WithContext(ctx).First(&problem, id).Error
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *ProblemRepo) FindByIDs(ctx context.Context, ids []int64) ([]model.Problem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var problem []model.Problem
	err := r.db.WithContext(ctx).Where("id IN ?", ids).
		Find(&problem).Error
	if err != nil {
		return nil, err
	}
	return problem, nil
}

func (r *ProblemRepo) GetProblemSamplesByID(ctx context.Context, problemID int64) ([]model.ProblemSamples, error) {
	var samples []model.ProblemSamples
	if err := r.db.WithContext(ctx).Where("problem_id = ?", problemID).Find(&samples).Error; err != nil {
		return nil, err
	}
	return samples, nil
}

func (r *ProblemRepo) GetProblemTagByID(ctx context.Context, problemID int64) ([]model.Tag, error) {
	var tags []model.Tag
	if err := r.db.WithContext(ctx).Table("tags").
		Joins("JOIN problem_tags ON problem_tags.tag_id = tags.id").
		Where("problem_tags.problem_id = ?", problemID).
		Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

type ProblemTagItem struct {
	ProblemID int64  `gorm:"column:problem_id"`
	ID        int64  `gorm:"column:id"`
	Name      string `gorm:"column:name"`
}

// GetProblemTagsByIDs loads tags for one page of problems in a single query,
// avoiding one join query per problem in the problem list API.
func (r *ProblemRepo) GetProblemTagsByIDs(ctx context.Context, problemIDs []int64) ([]ProblemTagItem, error) {
	if len(problemIDs) == 0 {
		return nil, nil
	}
	var tags []ProblemTagItem
	err := r.db.WithContext(ctx).Table("problem_tags").
		Select("problem_tags.problem_id, tags.id, tags.name").
		Joins("JOIN tags ON tags.id = problem_tags.tag_id").
		Where("problem_tags.problem_id IN ?", problemIDs).
		Find(&tags).Error
	return tags, err
}

func (r *ProblemRepo) CreateProblemTag(ctx context.Context, tags []model.ProblemTag) error {
	return r.db.WithContext(ctx).Create(tags).Error
}

func (r *ProblemRepo) GetTagsList(ctx context.Context) ([]model.Tag, error) {
	var tags []model.Tag
	if err := r.db.WithContext(ctx).Model(&model.Tag{}).Order("name ASC, id ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

type AdminTagItem struct {
	ID              int64  `gorm:"column:id"`
	Name            string `gorm:"column:name"`
	ProblemCount    int64  `gorm:"column:problem_count"`
	ProblemSetCount int64  `gorm:"column:problem_set_count"`
}

func (r *ProblemRepo) GetAdminTagsList(ctx context.Context) ([]AdminTagItem, error) {
	var tags []AdminTagItem
	err := r.db.WithContext(ctx).Table("tags").
		Select(`tags.id, tags.name,
			COUNT(DISTINCT problem_tags.problem_id) AS problem_count,
			COUNT(DISTINCT problem_set_tags.problem_set_id) AS problem_set_count`).
		Joins("LEFT JOIN problem_tags ON problem_tags.tag_id = tags.id").
		Joins("LEFT JOIN problem_set_tags ON problem_set_tags.tag_id = tags.id").
		Group("tags.id, tags.name").
		Order("tags.name ASC, tags.id ASC").
		Scan(&tags).Error
	return tags, err
}

func (r *ProblemRepo) GetTagByID(ctx context.Context, id int64) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).First(&tag, id).Error
	return &tag, err
}

func (r *ProblemRepo) ExistsTagName(ctx context.Context, name string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Tag{}).
		Where("name = ? AND id <> ?", name, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *ProblemRepo) CreateTag(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *ProblemRepo) UpdateTagName(ctx context.Context, id int64, name string) error {
	return r.db.WithContext(ctx).Model(&model.Tag{}).
		Where("id = ?", id).
		Update("name", name).Error
}

// DeleteTagWithRelations 删除标签以及题目、题单中的引用，保证不会留下无效关联。
func (r *ProblemRepo) DeleteTagWithRelations(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", id).Delete(&model.ProblemTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("tag_id = ?", id).Delete(&model.ProblemSetTag{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.Tag{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *ProblemRepo) GetSubmissionCount(ctx context.Context, userID int64) (int64, int64, error) {
	var submissionCount int64
	var acCount int64
	if err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
		Select("COALESCE(SUM(submit_count), 0)").
		Where("user_id = ?", userID).
		Pluck("COALESCE(SUM(submit_count), 0)", &submissionCount).Error; err != nil {
		return 0, 0, err
	}
	if err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
		Select("COALESCE(SUM(ac_count), 0)").
		Where("user_id = ?", userID).
		Pluck("COALESCE(SUM(ac_count), 0)", &acCount).Error; err != nil {
		return 0, 0, err
	}
	return submissionCount, acCount, nil
}

type UserSubmitCount struct {
	UserID      int64 `gorm:"column:user_id" json:"userID"`
	ACCount     int64 `gorm:"column:ac_count" json:"acCount"`
	SubmitCount int64 `gorm:"column:submit_count" json:"submitCount"`
}

type UserSubmitCountFilter struct {
	QueryName bool
	UserIDs   []int
	Page      int
	PageSize  int
}

func (r *ProblemRepo) GetUserSubmitCount(ctx context.Context, filter *UserSubmitCountFilter) ([]UserSubmitCount, int64, error) {
	var result []UserSubmitCount
	var count int64

	query := r.db.WithContext(ctx).Model(&model.UserProblem{}).
		Select("user_id, SUM(ac_count) as ac_count, SUM(submit_count) as submit_count")

	if filter.QueryName {
		query = query.Where("user_id IN ?", filter.UserIDs)
		if err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
			Select("COUNT(DISTINCT user_id) as count").
			Where("user_id IN ?", filter.UserIDs).
			Pluck("count", &count).Error; err != nil {
			return result, 0, err
		}
	} else {
		if err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
			Select("COUNT(DISTINCT user_id) as count").
			Pluck("count", &count).Error; err != nil {
			return result, 0, err
		}
	}

	query = query.
		Group("user_id").
		Order("ac_count DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize)

	err := query.Find(&result).Error
	if err != nil {
		return result, 0, nil
	}
	return result, count, nil
}

func (r *ProblemRepo) GetProblemIDByUserID(ctx context.Context, userID int64) ([]int64, []int64, error) {
	var acIds []int64
	var unacIds []int64

	if err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
		Where("user_id = ? AND status = ?", userID, judge.Accepted).
		Pluck("problem_id", &acIds).Error; err != nil {
		return nil, nil, err
	}
	if err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
		Where("user_id = ? AND status != ?", userID, judge.Accepted).
		Pluck("problem_id", &unacIds).Error; err != nil {
		return nil, nil, err
	}

	return acIds, unacIds, nil
}

// GetUserProblemStatusByIDs only loads the current page's user/problem statuses.
// Loading every problem the user has touched becomes expensive for active users.
func (r *ProblemRepo) GetUserProblemStatusByIDs(ctx context.Context, userID int64, problemIDs []int64) (map[int64]int, error) {
	if len(problemIDs) == 0 {
		return map[int64]int{}, nil
	}
	var records []model.UserProblem
	err := r.db.WithContext(ctx).Model(&model.UserProblem{}).
		Select("problem_id, status").
		Where("user_id = ? AND problem_id IN ?", userID, problemIDs).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	statusMap := make(map[int64]int, len(records))
	for _, record := range records {
		statusMap[record.ProblemID] = record.Status
	}
	return statusMap, nil
}
