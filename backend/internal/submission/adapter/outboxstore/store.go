// Package outboxstore 提供 Submission Outbox 的 GORM 租约适配器。
package outboxstore

import (
	"context"
	"fmt"
	"time"

	"zoj/internal/model"
	"zoj/internal/submission/application/dispatchrelay"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

var _ dispatchrelay.Store = (*Store)(nil)

// Claim 在短事务中锁定一个有界批次。Processing 记录把 available_at 推进到租约结束时间，
// 因此同一个 (status, available_at) 索引也能用于回收过期租约。
func (s *Store) Claim(
	ctx context.Context,
	owner string,
	now, leaseUntil time.Time,
	limit int,
) ([]dispatchrelay.Item, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("claim submission outbox: store is not initialized")
	}
	if owner == "" || limit <= 0 || !leaseUntil.After(now) {
		return nil, fmt.Errorf("claim submission outbox: invalid lease arguments")
	}

	var items []dispatchrelay.Item
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []model.SubmissionOutbox
		if err := claimRowsQuery(tx, now, limit, &rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}

		ids := make([]int64, len(rows))
		for i := range rows {
			ids[i] = rows[i].ID
		}
		updated := tx.Model(&model.SubmissionOutbox{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"status":       model.SubmissionOutboxProcessing,
				"attempts":     gorm.Expr("attempts + 1"),
				"available_at": leaseUntil,
				"locked_by":    owner,
				"locked_until": leaseUntil,
			})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != int64(len(rows)) {
			return fmt.Errorf("claimed %d submission outboxes but updated %d", len(rows), updated.RowsAffected)
		}

		items = make([]dispatchrelay.Item, 0, len(rows))
		for _, row := range rows {
			items = append(items, dispatchrelay.Item{
				OutboxID:     row.ID,
				SubmissionID: row.SubmissionID,
				Version:      row.SubmissionVersion,
				Attempts:     row.Attempts + 1,
			})
		}
		return nil
	})
	return items, err
}

func claimRowsQuery(tx *gorm.DB, now time.Time, limit int, rows *[]model.SubmissionOutbox) *gorm.DB {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status IN ? AND available_at <= ?",
			[]int{int(model.SubmissionOutboxPending), int(model.SubmissionOutboxProcessing)}, now).
		Order("available_at ASC").
		Order("id ASC").
		Limit(limit).
		Find(rows)
}

func (s *Store) MarkPublished(
	ctx context.Context,
	outboxID int64,
	owner string,
	publishedAt time.Time,
) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("mark submission outbox published: store is not initialized")
	}
	result := s.db.WithContext(ctx).Model(&model.SubmissionOutbox{}).
		Where("id = ? AND status = ? AND locked_by = ?", outboxID, model.SubmissionOutboxProcessing, owner).
		Updates(map[string]any{
			"status":       model.SubmissionOutboxPublished,
			"published_at": publishedAt,
			"locked_by":    "",
			"locked_until": nil,
			"last_error":   "",
		})
	return result.RowsAffected > 0, result.Error
}

func (s *Store) MarkFailed(
	ctx context.Context,
	outboxID int64,
	owner string,
	availableAt time.Time,
	lastError string,
) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("mark submission outbox failed: store is not initialized")
	}
	result := s.db.WithContext(ctx).Model(&model.SubmissionOutbox{}).
		Where("id = ? AND status = ? AND locked_by = ?", outboxID, model.SubmissionOutboxProcessing, owner).
		Updates(map[string]any{
			"status":       model.SubmissionOutboxPending,
			"available_at": availableAt,
			"locked_by":    "",
			"locked_until": nil,
			"last_error":   lastError,
		})
	return result.RowsAffected > 0, result.Error
}
