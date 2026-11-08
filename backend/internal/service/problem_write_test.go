package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func TestProblemWriteError(t *testing.T) {
	tests := []struct {
		name     string
		source   error
		writeRow bool
		wantCode errcode.Code
	}{
		{
			name:     "concurrent display ID conflict",
			source:   fmt.Errorf("insert problem: %w", &mysql.MySQLError{Number: 1062}),
			writeRow: true,
			wantCode: errcode.ProblemDisplayIDExists,
		},
		{
			name:     "duplicate translated by GORM",
			source:   gorm.ErrDuplicatedKey,
			writeRow: true,
			wantCode: errcode.ProblemDisplayIDExists,
		},
		{
			name:     "duplicate in a related table",
			source:   &mysql.MySQLError{Number: 1062},
			wantCode: errcode.DatabaseError,
		},
		{
			name:     "other problem write failure",
			source:   &mysql.MySQLError{Number: 1213},
			writeRow: true,
			wantCode: errcode.DatabaseError,
		},
		{
			name:     "existing business error",
			source:   errcode.ErrInvalidParams,
			wantCode: errcode.InvalidParams,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source
			if tt.writeRow {
				err = problemRowWriteError(err)
			}
			err = problemTransactionError(err)

			var appErr *errcode.AppErr
			if !errors.As(err, &appErr) || appErr.Code != tt.wantCode {
				t.Fatalf("got %v, want code %d", err, tt.wantCode)
			}
			if !errors.Is(err, tt.source) {
				t.Fatalf("lost original cause: %v", err)
			}
		})
	}
}
