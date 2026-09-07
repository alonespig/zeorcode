package repository

import (
	"context"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type RemoteAccountRepo struct {
	db *gorm.DB
}

func NewRemoteAccountRepo(db *gorm.DB) *RemoteAccountRepo {
	return &RemoteAccountRepo{db: db}
}

func (r *RemoteAccountRepo) List(ctx context.Context) ([]model.RemoteAccount, error) {
	var list []model.RemoteAccount
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *RemoteAccountRepo) GetByID(ctx context.Context, id int64) (*model.RemoteAccount, error) {
	var a model.RemoteAccount
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// FirstEnabledByOJ 取某 OJ 一个启用的账号(给爬虫/远程判题复用)
func (r *RemoteAccountRepo) FirstEnabledByOJ(ctx context.Context, oj string) (*model.RemoteAccount, error) {
	var a model.RemoteAccount
	if err := r.db.WithContext(ctx).Where("oj = ? AND enabled = ?", oj, true).
		Order("id ASC").First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// AcquireByOJ 从账号池抢占一个该 OJ 的空闲启用账号(原子置 busy)，无可用则返回 (nil,nil)。
// 用 “UPDATE ... WHERE id=? AND busy=false” 的条件更新做 CAS：谁先把 false 改成 true 谁拿到。
func (r *RemoteAccountRepo) AcquireByOJ(ctx context.Context, oj string) (*model.RemoteAccount, error) {
	var list []model.RemoteAccount
	if err := r.db.WithContext(ctx).Where("oj = ? AND enabled = ? AND busy = ?", oj, true, false).
		Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		res := r.db.WithContext(ctx).Model(&model.RemoteAccount{}).
			Where("id = ? AND busy = ?", list[i].ID, false).
			Update("busy", true)
		if res.Error == nil && res.RowsAffected > 0 {
			list[i].Busy = true
			return &list[i], nil
		}
	}
	return nil, nil
}

// CountEnabledByOJ 该 OJ 启用账号总数(用来区分"没配账号"与"账号都忙")
func (r *RemoteAccountRepo) CountEnabledByOJ(ctx context.Context, oj string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.RemoteAccount{}).
		Where("oj = ? AND enabled = ?", oj, true).Count(&n).Error
	return n, err
}

// Release 归还账号(置回空闲)
func (r *RemoteAccountRepo) Release(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.RemoteAccount{}).Where("id = ?", id).Update("busy", false).Error
}

// ResetBusy 单实例启动时重置所有 busy（回收上次崩溃残留的占用）。
// 多实例部署必须禁用，避免清除其他实例正在使用的账号。
func (r *RemoteAccountRepo) ResetBusy(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&model.RemoteAccount{}).Where("busy = ?", true).Update("busy", false).Error
}

func (r *RemoteAccountRepo) Create(ctx context.Context, a *model.RemoteAccount) (int64, error) {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return 0, err
	}
	return a.ID, nil
}

// Update 用 map 更新，避免 GORM 结构体更新忽略 false/"" 零值。
func (r *RemoteAccountRepo) Update(ctx context.Context, id int64, fields map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.RemoteAccount{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// MySQL 更新为相同值时 RowsAffected 也可能为 0，需再确认记录是否存在。
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.RemoteAccount{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func (r *RemoteAccountRepo) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.RemoteAccount{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
