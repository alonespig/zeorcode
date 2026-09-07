package repository

import (
	"context"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type ProblemSetRepo struct {
	db *gorm.DB
}

func NewProblemSetRepo(db *gorm.DB) *ProblemSetRepo {
	return &ProblemSetRepo{db: db}
}

// ProblemSetQuery 题单列表查询条件。IncludeDraft 只在后台列表为 true。
type ProblemSetQuery struct {
	Page         int
	PageSize     int
	Keyword      string
	TagIDs       []int64
	Visibility   *int
	IncludeDraft bool
}

// ProblemSetTagItem 批量取题单标签的行，避免逐个题单查一次。
type ProblemSetTagItem struct {
	ProblemSetID int64  `gorm:"column:problem_set_id"`
	ID           int64  `gorm:"column:id"`
	Name         string `gorm:"column:name"`
}

// ProblemSetProblemRef 题单-题目引用（按 sort 升序）。
type ProblemSetProblemRef struct {
	ProblemSetID int64 `gorm:"column:problem_set_id"`
	ProblemID    int64 `gorm:"column:problem_id"`
	Sort         int   `gorm:"column:sort"`
}

func (r *ProblemSetRepo) List(ctx context.Context, q *ProblemSetQuery) ([]model.ProblemSet, int64, error) {
	var sets []model.ProblemSet
	var total int64
	db := r.db.WithContext(ctx).Model(&model.ProblemSet{})

	if !q.IncludeDraft {
		db = db.Where("published = ?", model.ProblemSetPublished)
	}
	if q.Keyword != "" {
		db = db.Where("title LIKE ?", "%"+q.Keyword+"%")
	}
	if q.Visibility != nil {
		db = db.Where("visibility = ?", *q.Visibility)
	}
	if len(q.TagIDs) > 0 {
		// AND 交集：题单须同时含全部选中标签 —— 命中标签数 == 选中标签数
		sub := r.db.WithContext(ctx).Table("problem_set_tags").
			Select("problem_set_id").
			Where("tag_id IN ?", q.TagIDs).
			Group("problem_set_id").
			Having("COUNT(DISTINCT tag_id) = ?", len(q.TagIDs))
		db = db.Where("id IN (?)", sub)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("updated_at DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&sets).Error
	if err != nil {
		return nil, 0, err
	}
	return sets, total, nil
}

func (r *ProblemSetRepo) GetByID(ctx context.Context, id int64) (*model.ProblemSet, error) {
	var set model.ProblemSet
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&set).Error
	return &set, err
}

func (r *ProblemSetRepo) PublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProblemSet{}).
		Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *ProblemSetRepo) ResolveIDByPublicID(ctx context.Context, publicID int64) (int64, error) {
	var set model.ProblemSet
	err := r.db.WithContext(ctx).Select("id").
		Where("public_id = ?", publicID).First(&set).Error
	return set.ID, err
}

// ProblemRefsBySetIDs 批量取多个题单的题目引用，按 (题单, sort) 升序。
// 列表页用它一次算出所有题单的题数和进度，避免 N+1。
func (r *ProblemSetRepo) ProblemRefsBySetIDs(ctx context.Context, setIDs []int64) ([]ProblemSetProblemRef, error) {
	if len(setIDs) == 0 {
		return nil, nil
	}
	var refs []ProblemSetProblemRef
	err := r.db.WithContext(ctx).Table("problem_set_problems").
		Select("problem_set_id, problem_id, sort").
		Where("problem_set_id IN ?", setIDs).
		Order("problem_set_id ASC, sort ASC").
		Find(&refs).Error
	return refs, err
}

// TagsBySetIDs 批量取多个题单的标签（复用题目那套 tags 表）。
func (r *ProblemSetRepo) TagsBySetIDs(ctx context.Context, setIDs []int64) ([]ProblemSetTagItem, error) {
	if len(setIDs) == 0 {
		return nil, nil
	}
	var tags []ProblemSetTagItem
	err := r.db.WithContext(ctx).Table("problem_set_tags").
		Select("problem_set_tags.problem_set_id, tags.id, tags.name").
		Joins("JOIN tags ON tags.id = problem_set_tags.tag_id").
		Where("problem_set_tags.problem_set_id IN ?", setIDs).
		Find(&tags).Error
	return tags, err
}

// CreateWithRelations 原子创建题单及其题目关联和标签关联。
func (r *ProblemSetRepo) CreateWithRelations(ctx context.Context, set *model.ProblemSet, problems []model.ProblemSetProblem, tags []model.ProblemSetTag) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(set).Error; err != nil {
			return err
		}
		return writeProblemSetRelations(tx, set.ID, problems, tags)
	})
}

// UpdateWithRelations 原子更新题单：主记录 Updates + 关联「先删后插」全量替换。
// 与 ContestRepo.UpdateContestWithProblems 同一套做法。
func (r *ProblemSetRepo) UpdateWithRelations(ctx context.Context, set *model.ProblemSet, problems []model.ProblemSetProblem, tags []model.ProblemSetTag) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 用 Select 显式列出可变字段：Updates 默认忽略零值，
		// 而「取消发布」「改回公开」都是要写 0 的。
		err := tx.Model(&model.ProblemSet{}).Where("id = ?", set.ID).
			Select("title", "description", "published", "visibility", "invite_code").
			Updates(set).Error
		if err != nil {
			return err
		}
		if err := tx.Where("problem_set_id = ?", set.ID).Delete(&model.ProblemSetProblem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("problem_set_id = ?", set.ID).Delete(&model.ProblemSetTag{}).Error; err != nil {
			return err
		}
		return writeProblemSetRelations(tx, set.ID, problems, tags)
	})
}

func writeProblemSetRelations(tx *gorm.DB, setID int64, problems []model.ProblemSetProblem, tags []model.ProblemSetTag) error {
	if len(problems) > 0 {
		for i := range problems {
			problems[i].ProblemSetID = setID
		}
		if err := tx.Create(&problems).Error; err != nil {
			return err
		}
	}
	if len(tags) > 0 {
		for i := range tags {
			tags[i].ProblemSetID = setID
		}
		if err := tx.Create(&tags).Error; err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除题单，连带清掉题目、标签和解锁记录，避免留孤儿行。
func (r *ProblemSetRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_set_id = ?", id).Delete(&model.ProblemSetProblem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("problem_set_id = ?", id).Delete(&model.ProblemSetTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("problem_set_id = ?", id).Delete(&model.ProblemSetUnlock{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.ProblemSet{}).Error
	})
}

func (r *ProblemSetRepo) IsUnlocked(ctx context.Context, userID, setID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProblemSetUnlock{}).
		Where("user_id = ? AND problem_set_id = ?", userID, setID).
		Count(&count).Error
	return count > 0, err
}

// CreateUnlock 记录解锁。并发重复解锁会撞唯一键，这里当成已解锁忽略。
func (r *ProblemSetRepo) CreateUnlock(ctx context.Context, userID, setID int64) error {
	unlock := model.ProblemSetUnlock{UserID: userID, ProblemSetID: setID}
	return r.db.WithContext(ctx).
		Where("user_id = ? AND problem_set_id = ?", userID, setID).
		FirstOrCreate(&unlock).Error
}
