//go:build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"zoj/internal/model"
	"zoj/pkg/judge"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openSubmissionOutboxTestDB(t *testing.T) *gorm.DB {
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
	return db
}

func prepareSubmissionOutboxTestSchema(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Exec("DROP TEMPORARY TABLE IF EXISTS judge_results, submission_outboxes, submissions").Error
		_ = tx.Rollback().Error
	})

	statements := []string{
		`CREATE TEMPORARY TABLE submissions (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			public_id BIGINT NOT NULL, problem_id BIGINT, user_id BIGINT,
			contest_id BIGINT NOT NULL DEFAULT 0, homework_id BIGINT NOT NULL DEFAULT 0,
			code LONGTEXT NOT NULL, language LONGTEXT NOT NULL, status BIGINT NOT NULL,
			time_used BIGINT, memory_used BIGINT, compile_output MEDIUMTEXT,
			score BIGINT NOT NULL DEFAULT 0, version BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME(3), UNIQUE KEY idx_submissions_public_id (public_id)
		) ENGINE=InnoDB`,
		`CREATE TEMPORARY TABLE submission_outboxes (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			submission_id BIGINT NOT NULL, submission_version BIGINT NOT NULL,
			status TINYINT UNSIGNED NOT NULL DEFAULT 0, attempts BIGINT NOT NULL DEFAULT 0,
			available_at DATETIME(3) NOT NULL, locked_by VARCHAR(128) NOT NULL DEFAULT '',
			locked_until DATETIME(3), published_at DATETIME(3),
			last_error VARCHAR(1024) NOT NULL DEFAULT '', created_at DATETIME(3), updated_at DATETIME(3),
			UNIQUE KEY idx_submission_outbox_version (submission_id, submission_version)
		) ENGINE=InnoDB`,
		`CREATE TEMPORARY TABLE judge_results (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			submission_id BIGINT, case_id BIGINT, status BIGINT,
			time_used BIGINT, memory_used BIGINT, output_diff LONGTEXT
		) ENGINE=InnoDB`,
	}
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			t.Fatalf("prepare temporary schema: %v", err)
		}
	}
	return tx
}

func TestSubmissionRepoCreatePendingCommitsSubmissionAndOutbox(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)

	dispatch, err := repo.CreatePending(context.Background(), &model.Submission{
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
	})
	if err != nil {
		t.Fatalf("CreatePending() error = %v", err)
	}
	if dispatch.SubmissionID == 0 || dispatch.Version != 0 {
		t.Fatalf("dispatch = %+v", dispatch)
	}

	var submissionCount, outboxCount int64
	if err := db.Model(&model.Submission{}).Where("id = ?", dispatch.SubmissionID).Count(&submissionCount).Error; err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if err := db.Model(&model.SubmissionOutbox{}).
		Where("submission_id = ? AND submission_version = ?", dispatch.SubmissionID, dispatch.Version).
		Count(&outboxCount).Error; err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if submissionCount != 1 || outboxCount != 1 {
		t.Fatalf("submission/outbox counts = %d/%d, want 1/1", submissionCount, outboxCount)
	}
}

func TestSubmissionRepoCreatePendingRollsBackWhenOutboxInsertFails(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)

	if err := db.Create(&model.SubmissionOutbox{
		SubmissionID:      99,
		SubmissionVersion: 0,
		AvailableAt:       testNow(),
	}).Error; err != nil {
		t.Fatalf("seed conflicting outbox: %v", err)
	}

	_, err := repo.CreatePending(context.Background(), &model.Submission{
		ID:        99,
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
	})
	if err == nil {
		t.Fatal("CreatePending() error = nil, want duplicate outbox error")
	}

	var count int64
	if err := db.Model(&model.Submission{}).Where("id = ?", 99).Count(&count).Error; err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if count != 0 {
		t.Fatalf("submission count = %d, want rollback to 0", count)
	}
}

func TestSubmissionRepoResetForRejudgeCreatesVersionedOutbox(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	submission := model.Submission{
		ID:         21,
		ProblemID:  7,
		UserID:     11,
		Language:   "c++",
		Code:       "int main() {}",
		Status:     judge.Accepted,
		TimeUsed:   12,
		MemoryUsed: 34,
		Score:      100,
		Version:    3,
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}
	if err := db.Create(&model.JudgeResult{SubmissionID: submission.ID, CaseID: 1, Status: judge.Accepted}).Error; err != nil {
		t.Fatalf("seed judge result: %v", err)
	}

	dispatches, err := repo.ResetForRejudge(context.Background(), []int64{submission.ID})
	if err != nil {
		t.Fatalf("ResetForRejudge() error = %v", err)
	}
	want := SubmissionDispatch{SubmissionID: submission.ID, Version: 4}
	if len(dispatches) != 1 || dispatches[0] != want {
		t.Fatalf("dispatches = %v, want [%v]", dispatches, want)
	}

	var got model.Submission
	if err := db.First(&got, submission.ID).Error; err != nil {
		t.Fatalf("load reset submission: %v", err)
	}
	if got.Version != 4 || got.Status != judge.Pending || got.TimeUsed != 0 || got.MemoryUsed != 0 || got.Score != 0 {
		t.Fatalf("reset submission = %+v", got)
	}

	var resultCount, outboxCount int64
	if err := db.Model(&model.JudgeResult{}).Where("submission_id = ?", submission.ID).Count(&resultCount).Error; err != nil {
		t.Fatalf("count judge results: %v", err)
	}
	if err := db.Model(&model.SubmissionOutbox{}).
		Where("submission_id = ? AND submission_version = ?", want.SubmissionID, want.Version).
		Count(&outboxCount).Error; err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if resultCount != 0 || outboxCount != 1 {
		t.Fatalf("judge result/outbox counts = %d/%d, want 0/1", resultCount, outboxCount)
	}
}

func TestSubmissionRepoDispatchClaimIsIdempotentByMessageID(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	dispatch, err := repo.CreatePending(context.Background(), &model.Submission{
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
	})
	if err != nil {
		t.Fatalf("CreatePending() error = %v", err)
	}
	if err := repo.MarkDispatchPublished(context.Background(), dispatch); err != nil {
		t.Fatalf("MarkDispatchPublished() error = %v", err)
	}

	claim, err := repo.ClaimDispatchForJudging(context.Background(), dispatch, "1-0")
	if err != nil || claim != DispatchClaimed {
		t.Fatalf("first claim = %v, %v; want claimed", claim, err)
	}
	claim, err = repo.ClaimDispatchForJudging(context.Background(), dispatch, "1-0")
	if err != nil || claim != DispatchClaimed {
		t.Fatalf("same-message claim = %v, %v; want claimed", claim, err)
	}
	claim, err = repo.ClaimDispatchForJudging(context.Background(), dispatch, "2-0")
	if err != nil || claim != DispatchDuplicate {
		t.Fatalf("duplicate claim = %v, %v; want duplicate", claim, err)
	}

	released, err := repo.ReleaseDispatchForRetry(context.Background(), dispatch, "1-0")
	if err != nil || !released {
		t.Fatalf("release = %v, %v; want true", released, err)
	}
	claim, err = repo.ClaimDispatchForJudging(context.Background(), dispatch, "2-0")
	if err != nil || claim != DispatchClaimed {
		t.Fatalf("claim after release = %v, %v; want claimed", claim, err)
	}

	completed, err := repo.CompleteDispatchForJudging(context.Background(), dispatch, "2-0")
	if err != nil || !completed {
		t.Fatalf("complete = %v, %v; want true", completed, err)
	}
	completed, err = repo.CompleteDispatchForJudging(context.Background(), dispatch, "2-0")
	if err != nil || !completed {
		t.Fatalf("idempotent complete = %v, %v; want true", completed, err)
	}
}

func TestSubmissionRepoMarkPublishedDoesNotRegressJudgingState(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	dispatch, err := repo.CreatePending(context.Background(), &model.Submission{
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
	})
	if err != nil {
		t.Fatalf("CreatePending() error = %v", err)
	}
	claim, err := repo.ClaimDispatchForJudging(context.Background(), dispatch, "1-0")
	if err != nil || claim != DispatchClaimed {
		t.Fatalf("claim pending dispatch = %v, %v; want claimed", claim, err)
	}
	if err := repo.MarkDispatchPublished(context.Background(), dispatch); err != nil {
		t.Fatalf("late MarkDispatchPublished() error = %v", err)
	}

	var outbox model.SubmissionOutbox
	if err := db.Where("submission_id = ? AND submission_version = ?", dispatch.SubmissionID, dispatch.Version).
		First(&outbox).Error; err != nil {
		t.Fatalf("load outbox: %v", err)
	}
	if outbox.Status != model.SubmissionOutboxJudging || outbox.LockedBy != "1-0" {
		t.Fatalf("outbox regressed after late publish = %+v", outbox)
	}
}

func TestSubmissionRepoMarkDispatchPublishedTransitionsPending(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	dispatch, err := repo.CreatePending(context.Background(), &model.Submission{
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
	})
	if err != nil {
		t.Fatalf("CreatePending() error = %v", err)
	}

	if err := repo.MarkDispatchPublished(context.Background(), dispatch); err != nil {
		t.Fatalf("MarkDispatchPublished() error = %v", err)
	}

	var outbox model.SubmissionOutbox
	if err := db.Where("submission_id = ? AND submission_version = ?", dispatch.SubmissionID, dispatch.Version).
		First(&outbox).Error; err != nil {
		t.Fatalf("load outbox: %v", err)
	}
	if outbox.Status != model.SubmissionOutboxPublished || outbox.PublishedAt == nil ||
		outbox.LockedBy != "" || outbox.LockedUntil != nil {
		t.Fatalf("published outbox = %+v", outbox)
	}
}

func TestSubmissionRepoClaimBackfillsLegacyOutbox(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	submission := model.Submission{
		ID:        31,
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}

	dispatch := SubmissionDispatch{SubmissionID: submission.ID, Version: 0}
	claim, err := repo.ClaimDispatchForJudging(context.Background(), dispatch, "legacy-1")
	if err != nil || claim != DispatchClaimed {
		t.Fatalf("legacy claim = %v, %v; want claimed", claim, err)
	}

	var outbox model.SubmissionOutbox
	if err := db.Where("submission_id = ? AND submission_version = ?", submission.ID, 0).First(&outbox).Error; err != nil {
		t.Fatalf("load backfilled outbox: %v", err)
	}
	if outbox.Status != model.SubmissionOutboxJudging || outbox.LockedBy != "legacy-1" {
		t.Fatalf("backfilled outbox = %+v", outbox)
	}
}

func TestSubmissionRepoClaimRejectsStaleVersion(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	submission := model.Submission{
		ID:        41,
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
		Version:   2,
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}
	if err := db.Create(&model.SubmissionOutbox{
		SubmissionID:      submission.ID,
		SubmissionVersion: 1,
		Status:            model.SubmissionOutboxPublished,
		AvailableAt:       testNow(),
	}).Error; err != nil {
		t.Fatalf("seed stale outbox: %v", err)
	}

	claim, err := repo.ClaimDispatchForJudging(
		context.Background(),
		SubmissionDispatch{SubmissionID: submission.ID, Version: 1},
		"1-0",
	)
	if err != nil || claim != DispatchStale {
		t.Fatalf("stale claim = %v, %v; want stale", claim, err)
	}

	var outbox model.SubmissionOutbox
	if err := db.Where("submission_id = ? AND submission_version = ?", submission.ID, 1).First(&outbox).Error; err != nil {
		t.Fatalf("load stale outbox: %v", err)
	}
	if outbox.Status != model.SubmissionOutboxCompleted {
		t.Fatalf("stale outbox status = %d, want completed", outbox.Status)
	}
}

func TestSubmissionRepoSaveJudgingResultUpdatesSubmissionAndCasesAtomically(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	submission := model.Submission{
		ID:        51,
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
		Version:   2,
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}
	if err := db.Create(&model.JudgeResult{SubmissionID: submission.ID, CaseID: 99, Status: judge.WrongAnswer}).Error; err != nil {
		t.Fatalf("seed old judge result: %v", err)
	}

	submission.Status = judge.Accepted
	submission.TimeUsed = 12
	submission.MemoryUsed = 34
	submission.Score = 100
	results := []*model.JudgeResult{
		{SubmissionID: submission.ID, CaseID: 1, Status: judge.Accepted},
		{SubmissionID: submission.ID, CaseID: 2, Status: judge.Accepted},
	}
	saved, err := repo.SaveJudgingResult(context.Background(), &submission, results)
	if err != nil || !saved {
		t.Fatalf("SaveJudgingResult() = %v, %v; want saved", saved, err)
	}

	var got model.Submission
	if err := db.First(&got, submission.ID).Error; err != nil {
		t.Fatalf("load submission: %v", err)
	}
	if got.Status != judge.Accepted || got.TimeUsed != 12 || got.MemoryUsed != 34 || got.Score != 100 || got.Version != 2 {
		t.Fatalf("saved submission = %+v", got)
	}
	var cases []model.JudgeResult
	if err := db.Where("submission_id = ?", submission.ID).Order("case_id ASC").Find(&cases).Error; err != nil {
		t.Fatalf("load judge results: %v", err)
	}
	if len(cases) != 2 || cases[0].CaseID != 1 || cases[1].CaseID != 2 {
		t.Fatalf("saved judge results = %+v", cases)
	}
}

func TestSubmissionRepoSaveJudgingResultDoesNotWriteStaleCases(t *testing.T) {
	db := prepareSubmissionOutboxTestSchema(t, openSubmissionOutboxTestDB(t))
	repo := NewSubmissionRepo(db)
	current := model.Submission{
		ID:        61,
		ProblemID: 7,
		UserID:    11,
		Language:  "c++",
		Code:      "int main() {}",
		Status:    judge.Pending,
		Version:   2,
	}
	if err := db.Create(&current).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}
	if err := db.Create(&model.JudgeResult{SubmissionID: current.ID, CaseID: 99, Status: judge.WrongAnswer}).Error; err != nil {
		t.Fatalf("seed current judge result: %v", err)
	}

	stale := current
	stale.Version = 1
	stale.Status = judge.Accepted
	saved, err := repo.SaveJudgingResult(context.Background(), &stale, []*model.JudgeResult{
		{SubmissionID: current.ID, CaseID: 1, Status: judge.Accepted},
	})
	if err != nil || saved {
		t.Fatalf("stale SaveJudgingResult() = %v, %v; want not saved", saved, err)
	}

	var got model.Submission
	if err := db.First(&got, current.ID).Error; err != nil {
		t.Fatalf("load submission: %v", err)
	}
	if got.Version != 2 || got.Status != judge.Pending {
		t.Fatalf("stale write changed submission = %+v", got)
	}
	var cases []model.JudgeResult
	if err := db.Where("submission_id = ?", current.ID).Find(&cases).Error; err != nil {
		t.Fatalf("load judge results: %v", err)
	}
	if len(cases) != 1 || cases[0].CaseID != 99 || cases[0].Status != judge.WrongAnswer {
		t.Fatalf("stale write changed judge results = %+v", cases)
	}
}

func testNow() time.Time { return time.Now() }
