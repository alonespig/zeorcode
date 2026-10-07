package repository

import (
	"context"
	"strings"
	"time"

	"zoj/internal/model"

	"gorm.io/gorm"
)

type AuditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

type AuditListParams struct {
	Page     int
	PageSize int
	Keyword  string
	Method   string
	Success  *bool
}

type AuditLogRow struct {
	ID        int64
	ActorID   int64
	ActorName string
	Method    string
	Path      string
	Target    string
	ClientIP  string
	Success   bool
	Code      int
	CreatedAt time.Time
}

func (r *AuditRepo) Create(ctx context.Context, log *model.AdminAuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditRepo) List(ctx context.Context, params AuditListParams) ([]AuditLogRow, int64, error) {
	base := r.db.WithContext(ctx).Table("admin_audit_logs AS a").
		Joins("LEFT JOIN users AS u ON u.id = a.actor_id")
	keyword := strings.TrimSpace(params.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("a.path LIKE ? OR a.target LIKE ? OR u.username LIKE ?", like, like, like)
	}
	if params.Method != "" {
		base = base.Where("a.method = ?", strings.ToUpper(params.Method))
	}
	if params.Success != nil {
		base = base.Where("a.success = ?", *params.Success)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []AuditLogRow
	err := base.Select(`a.id, u.uid AS actor_id, u.username AS actor_name, a.method,
			a.path, a.target, a.client_ip, a.success, a.code, a.created_at`).
		Order("a.id DESC").
		Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		Scan(&list).Error
	return list, total, err
}
