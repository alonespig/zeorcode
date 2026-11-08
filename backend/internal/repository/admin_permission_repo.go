package repository

import (
	"context"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/model"

	"gorm.io/gorm"
)

// AdminPermissionRepo 管理员权限授权表。
type AdminPermissionRepo struct {
	db *gorm.DB
}

func NewAdminPermissionRepo(db *gorm.DB) *AdminPermissionRepo {
	return &AdminPermissionRepo{db: db}
}

// List 返回某个管理员被授予的全部权限点。
func (r *AdminPermissionRepo) List(ctx context.Context, userID int64) ([]string, error) {
	permissions := make([]string, 0)
	err := r.db.WithContext(ctx).
		Model(&model.AdminPermission{}).
		Where("user_id = ?", userID).
		Order("permission").
		Pluck("permission", &permissions).Error
	return permissions, err
}

// Replace 用给定集合整组替换某个管理员的权限，空集合表示收回全部权限。
func (r *AdminPermissionRepo) Replace(ctx context.Context, userID int64, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.AdminPermission{}).Error; err != nil {
			return err
		}
		if len(permissions) == 0 {
			return nil
		}

		seen := make(map[string]struct{}, len(permissions))
		rows := make([]model.AdminPermission, 0, len(permissions))
		now := time.Now()
		for _, permission := range permissions {
			if _, ok := seen[permission]; ok {
				continue
			}
			seen[permission] = struct{}{}
			rows = append(rows, model.AdminPermission{
				UserID:     userID,
				Permission: permission,
				CreatedAt:  now,
			})
		}
		return tx.Create(&rows).Error
	})
}

// DeleteByUser 收回某个用户的全部权限（角色变成非管理员时调用）。
func (r *AdminPermissionRepo) DeleteByUser(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.AdminPermission{}).Error
}
