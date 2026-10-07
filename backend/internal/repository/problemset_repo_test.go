package repository

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestProblemSetRankQueryUsesGlobalProblemStatus(t *testing.T) {
	db := newSubmissionRepoDryRunDB(t)
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return NewProblemSetRepo(tx).problemSetRankQuery(context.Background(), 7).
			Order("stats.solved_count DESC, users.id ASC").
			Find(&[]ProblemSetRankRecord{})
	})

	for _, want := range []string{
		"user_problems",
		"problem_set_problems",
		"up.status = 1",
		"up.status <> 0",
		"psp.problem_set_id = 7",
		"DENSE_RANK() OVER",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("rank SQL = %q, want fragment %q", sql, want)
		}
	}
}
