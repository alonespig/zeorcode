package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newSubmissionRepoDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	dialector := mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(127.0.0.1:3306)/gorm?parseTime=true",
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}
	return db
}

func TestMarkDispatchPublishedQueryExpandsStatusValues(t *testing.T) {
	db := newSubmissionRepoDryRunDB(t)
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return NewSubmissionRepo(tx).markDispatchPublishedQuery(
			context.Background(),
			SubmissionDispatch{SubmissionID: 7, Version: 2},
			time.Unix(0, 0),
		)
	})

	if !strings.Contains(sql, "status IN (0,1)") {
		t.Fatalf("mark-published SQL = %q, want two expanded status values", sql)
	}
}
