//go:build integration

package outboxstore

import (
	"context"
	"os"
	"testing"
	"time"

	"zoj/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("ZOJ_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = os.Getenv("FASTOJ_TEST_MYSQL_DSN")
	}
	if dsn == "" {
		t.Skip("ZOJ_TEST_MYSQL_DSN and legacy FASTOJ_TEST_MYSQL_DSN are not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open MySQL: %v", err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Exec("DROP TEMPORARY TABLE IF EXISTS submission_outboxes").Error
		_ = tx.Rollback().Error
	})
	if err := tx.Exec(`CREATE TEMPORARY TABLE submission_outboxes (
		id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		submission_id BIGINT NOT NULL, submission_version BIGINT NOT NULL,
		status TINYINT UNSIGNED NOT NULL DEFAULT 0, attempts BIGINT NOT NULL DEFAULT 0,
		available_at DATETIME(3) NOT NULL, locked_by VARCHAR(128) NOT NULL DEFAULT '',
		locked_until DATETIME(3), published_at DATETIME(3),
		last_error VARCHAR(1024) NOT NULL DEFAULT '', created_at DATETIME(3), updated_at DATETIME(3),
		UNIQUE KEY idx_submission_outbox_version (submission_id, submission_version),
		KEY idx_submission_outbox_dispatch (status, available_at)
	) ENGINE=InnoDB`).Error; err != nil {
		t.Fatalf("prepare temporary schema: %v", err)
	}
	return tx
}

func TestStoreClaimSelectsPendingAndExpiredProcessing(t *testing.T) {
	db := openTestDB(t)
	store := New(db)
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	leaseUntil := now.Add(30 * time.Second)
	rows := []model.SubmissionOutbox{
		{SubmissionID: 1, SubmissionVersion: 0, Status: model.SubmissionOutboxPending, AvailableAt: now.Add(-time.Second)},
		{SubmissionID: 2, SubmissionVersion: 3, Status: model.SubmissionOutboxProcessing, Attempts: 2, AvailableAt: now.Add(-time.Second)},
		{SubmissionID: 3, SubmissionVersion: 0, Status: model.SubmissionOutboxPending, AvailableAt: now.Add(time.Second)},
		{SubmissionID: 4, SubmissionVersion: 0, Status: model.SubmissionOutboxProcessing, AvailableAt: now.Add(time.Second)},
		{SubmissionID: 5, SubmissionVersion: 0, Status: model.SubmissionOutboxPublished, AvailableAt: now.Add(-time.Second)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed outboxes: %v", err)
	}

	items, err := store.Claim(context.Background(), "relay-a", now, leaseUntil, 10)
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if len(items) != 2 || items[0].SubmissionID != 1 || items[0].Attempts != 1 ||
		items[1].SubmissionID != 2 || items[1].Version != 3 || items[1].Attempts != 3 {
		t.Fatalf("claimed items = %+v", items)
	}

	for _, item := range items {
		var row model.SubmissionOutbox
		if err := db.First(&row, item.OutboxID).Error; err != nil {
			t.Fatalf("load claimed outbox %d: %v", item.OutboxID, err)
		}
		if row.Status != model.SubmissionOutboxProcessing || row.LockedBy != "relay-a" ||
			row.LockedUntil == nil || !row.LockedUntil.Equal(leaseUntil) || !row.AvailableAt.Equal(leaseUntil) {
			t.Fatalf("claimed outbox = %+v", row)
		}
	}
}

func TestStoreMarkPublishedRequiresLeaseOwner(t *testing.T) {
	db := openTestDB(t)
	store := New(db)
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	row := model.SubmissionOutbox{
		SubmissionID: 1, SubmissionVersion: 0,
		Status: model.SubmissionOutboxProcessing, AvailableAt: now, LockedBy: "relay-a", LockedUntil: &now,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	updated, err := store.MarkPublished(context.Background(), row.ID, "relay-b", now)
	if err != nil || updated {
		t.Fatalf("wrong-owner MarkPublished() = %v, %v", updated, err)
	}
	updated, err = store.MarkPublished(context.Background(), row.ID, "relay-a", now)
	if err != nil || !updated {
		t.Fatalf("owner MarkPublished() = %v, %v", updated, err)
	}

	var got model.SubmissionOutbox
	if err := db.First(&got, row.ID).Error; err != nil {
		t.Fatalf("load outbox: %v", err)
	}
	if got.Status != model.SubmissionOutboxPublished || got.LockedBy != "" || got.LockedUntil != nil || got.PublishedAt == nil {
		t.Fatalf("published outbox = %+v", got)
	}
}

func TestStoreMarkFailedSchedulesRetry(t *testing.T) {
	db := openTestDB(t)
	store := New(db)
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	row := model.SubmissionOutbox{
		SubmissionID: 1, SubmissionVersion: 0,
		Status: model.SubmissionOutboxProcessing, AvailableAt: now, LockedBy: "relay-a", LockedUntil: &now,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	next := now.Add(8 * time.Second)
	updated, err := store.MarkFailed(context.Background(), row.ID, "relay-a", next, "redis unavailable")
	if err != nil || !updated {
		t.Fatalf("MarkFailed() = %v, %v", updated, err)
	}

	var got model.SubmissionOutbox
	if err := db.First(&got, row.ID).Error; err != nil {
		t.Fatalf("load outbox: %v", err)
	}
	if got.Status != model.SubmissionOutboxPending || !got.AvailableAt.Equal(next) ||
		got.LockedBy != "" || got.LockedUntil != nil || got.LastError != "redis unavailable" {
		t.Fatalf("failed outbox = %+v", got)
	}
}
