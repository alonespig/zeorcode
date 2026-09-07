package repository

import (
	"context"
	"strings"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// CreateInBatches 广播扇出用，一次批量插入。
func (r *NotificationRepo) CreateInBatches(ctx context.Context, ns []*model.Notification) error {
	if len(ns) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(ns, 200).Error
}

// List 某用户的通知；typ 为空=全部；按 id 倒序分页。
func (r *NotificationRepo) List(ctx context.Context, recipientID int64, typ string, page, pageSize int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Notification{}).Where("recipient_id = ?", recipientID)
	if typ != "" {
		q = q.Where("type IN ?", strings.Split(typ, ",")) // 支持逗号多类型，如 comment,reply
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// UnreadCountByType 未读按类型分组计数。
func (r *NotificationRepo) UnreadCountByType(ctx context.Context, recipientID int64) (map[string]int64, error) {
	type row struct {
		Type string
		Cnt  int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Select("type, COUNT(*) AS cnt").
		Where("recipient_id = ? AND is_read = ?", recipientID, false).
		Group("type").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, rw := range rows {
		m[rw.Type] = rw.Cnt
	}
	return m, nil
}

// MarkRead：id>0 标单条；否则 typ 非空标该类型；都空标全部。
func (r *NotificationRepo) MarkRead(ctx context.Context, recipientID, id int64, typ string) error {
	q := r.db.WithContext(ctx).Model(&model.Notification{}).Where("recipient_id = ?", recipientID)
	if id > 0 {
		q = q.Where("id = ?", id)
	} else if typ != "" {
		q = q.Where("type = ?", typ)
	}
	return q.Update("is_read", true).Error
}
