package repository

import (
	"context"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type LanguageRepo struct {
	db *gorm.DB
}

func NewLanguageRepo(db *gorm.DB) *LanguageRepo {
	return &LanguageRepo{db: db}
}

func (r *LanguageRepo) List(ctx context.Context, onlyEnabled bool) ([]model.Language, error) {
	var list []model.Language
	db := r.db.WithContext(ctx)
	if onlyEnabled {
		db = db.Where("status = ?", model.LanguageEnabled)
	}
	if err := db.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *LanguageRepo) GetEnabledByID(ctx context.Context, id int) (*model.Language, error) {
	var language model.Language
	if err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, model.LanguageEnabled).
		First(&language).Error; err != nil {
		return nil, err
	}
	return &language, nil
}

func (r *LanguageRepo) GetByID(ctx context.Context, id int) (*model.Language, error) {
	var language model.Language
	if err := r.db.WithContext(ctx).First(&language, id).Error; err != nil {
		return nil, err
	}
	return &language, nil
}

func (r *LanguageRepo) ExistsByName(ctx context.Context, name string, excludeID int) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&model.Language{}).Where("name = ?", name)
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *LanguageRepo) Create(ctx context.Context, language *model.Language) error {
	return r.db.WithContext(ctx).Create(language).Error
}

func (r *LanguageRepo) Update(ctx context.Context, id int, fields map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.Language{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Language{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (r *LanguageRepo) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Language{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
