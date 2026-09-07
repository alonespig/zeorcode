package outboxstore

import (
	"strings"
	"testing"
	"time"

	"zoj/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newOutboxDryRunDB(t *testing.T) *gorm.DB {
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

func TestClaimRowsQueryExpandsStatusValues(t *testing.T) {
	db := newOutboxDryRunDB(t)
	var rows []model.SubmissionOutbox
	stmt := claimRowsQuery(db, time.Unix(0, 0), 100, &rows).Statement

	if sql := stmt.SQL.String(); !strings.Contains(sql, "status IN (?,?)") {
		t.Fatalf("claim SQL = %q, want two expanded status placeholders", sql)
	}
	if len(stmt.Vars) < 2 || stmt.Vars[0] != int(model.SubmissionOutboxPending) ||
		stmt.Vars[1] != int(model.SubmissionOutboxProcessing) {
		t.Fatalf("claim vars = %#v, want separate pending and processing values", stmt.Vars)
	}
}
