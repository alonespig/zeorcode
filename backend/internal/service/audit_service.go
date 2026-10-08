package service

import (
	"context"
	"time"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
)

type AuditService struct {
	repo *repository.AuditRepo
}

func NewAuditService(repo *repository.AuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

type AuditRecordParams struct {
	ActorID  int64
	Method   string
	Path     string
	Target   string
	ClientIP string
	Success  bool
	Code     int
}

type AuditQueryParams struct {
	Page     int
	PageSize int
	Keyword  string
	Method   string
	Success  *bool
}

type AuditLogItem struct {
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

type AuditLogList struct {
	Total int64
	List  []AuditLogItem
}

func (s *AuditService) Record(ctx context.Context, params AuditRecordParams) {
	if params.ActorID <= 0 || params.Path == "" {
		return
	}
	_ = s.repo.Create(ctx, &model.AdminAuditLog{
		ActorID: params.ActorID, Method: params.Method, Path: params.Path,
		Target: params.Target, ClientIP: params.ClientIP,
		Success: params.Success, Code: params.Code,
	})
}

func (s *AuditService) List(ctx context.Context, params AuditQueryParams) (*AuditLogList, error) {
	rows, total, err := s.repo.List(ctx, repository.AuditListParams{
		Page: params.Page, PageSize: params.PageSize, Keyword: params.Keyword,
		Method: params.Method, Success: params.Success,
	})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	result := &AuditLogList{Total: total, List: make([]AuditLogItem, 0, len(rows))}
	for _, row := range rows {
		result.List = append(result.List, AuditLogItem{
			ID: row.ID, ActorID: row.ActorID, ActorName: row.ActorName,
			Method: row.Method, Path: row.Path, Target: row.Target,
			ClientIP: row.ClientIP, Success: row.Success, Code: row.Code, CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}
