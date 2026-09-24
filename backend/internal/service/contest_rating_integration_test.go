//go:build integration

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"zoj/internal/common/consts"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/judge"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 使用独立数据库，以便多连接测试真实行锁；不修改 DSN 指定库中的表。
func newRatingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("ZOJ_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("ZOJ_TEST_MYSQL_DSN is not set (CREATE DATABASE privilege required)")
	}
	admin, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = adminSQL.Close() })
	name := fmt.Sprintf("zoj_rating_test_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE DATABASE `" + name + "`").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec("DROP DATABASE `" + name + "`").Error; err != nil {
			t.Error(err)
		}
	})
	config := admin.Dialector.(*mysql.Dialector).DSNConfig.Clone()
	config.DBName = name
	db, err := gorm.Open(mysql.Open(config.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&model.Contest{}, &model.ContestUser{}, &model.ContestProblem{},
		&model.Submission{}, &model.SubmissionOutbox{}, &model.JudgeResult{}, &model.User{}, &model.RatingChange{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedRatingContest(t *testing.T, db *gorm.DB, status int) (model.Contest, []model.Submission) {
	t.Helper()
	end := time.Now().Add(-2 * time.Minute)
	contest := model.Contest{PublicID: 100, Type: consts.ContestACM, Rated: true,
		StartTime: end.Add(-time.Hour), EndTime: end}
	if err := db.Create(&contest).Error; err != nil {
		t.Fatal(err)
	}
	users := []model.User{{ID: 1, UID: 1, Username: "first"}, {ID: 2, UID: 2, Username: "second"}}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	registrations := []model.ContestUser{{ContestID: contest.ID, UserID: 1}, {ContestID: contest.ID, UserID: 2}}
	if err := db.Create(&registrations).Error; err != nil {
		t.Fatal(err)
	}
	subs := []model.Submission{
		{PublicID: 1, ContestID: contest.ID, UserID: 1, ProblemID: 1, Status: judge.Accepted, CreatedAt: end.Add(-10 * time.Minute)},
		{PublicID: 2, ContestID: contest.ID, UserID: 2, ProblemID: 1, Status: status, CreatedAt: end.Add(-20 * time.Minute)},
	}
	if err := db.Create(&subs).Error; err != nil {
		t.Fatal(err)
	}
	return contest, subs
}

func runRatingSettlement(ctx context.Context, db *gorm.DB, contestID int64) ([]ratingSettlement, error) {
	var notes []ratingSettlement
	err := repository.NewContestRepo(db).SettleTx(ctx, contestID, func(tx *gorm.DB, c *model.Contest) error {
		var err error
		notes, err = settleContestRating(ctx, tx, c)
		return err
	})
	return notes, err
}

func TestRatingWaitsForPendingAndSettlesOnce(t *testing.T) {
	db := newRatingTestDB(t)
	contest, subs := seedRatingContest(t, db, judge.Pending)
	ctx := t.Context()
	if notes, err := runRatingSettlement(ctx, db, contest.ID); err != nil || len(notes) != 0 {
		t.Fatalf("pending settlement = %v, %v", notes, err)
	}
	var changes int64
	if err := db.Model(&model.RatingChange{}).Count(&changes).Error; err != nil || changes != 0 {
		t.Fatalf("pending rating changes = %d, %v", changes, err)
	}
	if err := db.Model(&model.Submission{}).Where("id = ?", subs[1].ID).Update("status", judge.Accepted).Error; err != nil {
		t.Fatal(err)
	}
	if notes, err := runRatingSettlement(ctx, db, contest.ID); err != nil || len(notes) != 2 {
		t.Fatalf("finished settlement = %v, %v", notes, err)
	}
	var history []model.RatingChange
	if err := db.Order("user_id").Find(&history).Error; err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Rank != 2 || history[1].Rank != 1 {
		t.Fatalf("must include final result of previously pending submission: %+v", history)
	}
	if notes, err := runRatingSettlement(ctx, db, contest.ID); err != nil || len(notes) != 0 {
		t.Fatalf("duplicate settlement = %v, %v", notes, err)
	}
	if err := db.Model(&model.RatingChange{}).Count(&changes).Error; err != nil || changes != 2 {
		t.Fatalf("rating changes = %d, %v", changes, err)
	}
	var users []model.User
	if err := db.Order("id").Find(&users).Error; err != nil {
		t.Fatal(err)
	}
	for i, user := range users {
		if user.Rating != history[i].NewRating {
			t.Fatalf("user %d rating changed twice: %d", user.ID, user.Rating)
		}
	}
	if _, err := repository.NewSubmissionRepo(db).ResetForRejudge(ctx, []int64{subs[0].ID, subs[1].ID}); !errors.Is(err, repository.ErrContestRatingSettled) {
		t.Fatalf("settled rejudge error = %v", err)
	}
	var unchanged model.Submission
	if err := db.First(&unchanged, subs[0].ID).Error; err != nil || unchanged.Version != 0 || unchanged.Status != judge.Accepted {
		t.Fatalf("rejected rejudge changed submission: %+v, %v", unchanged, err)
	}
}

func TestRatingSeesRejudgeCommittedWhileWaitingForContestLock(t *testing.T) {
	db := newRatingTestDB(t)
	contest, subs := seedRatingContest(t, db, judge.Accepted)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	if _, err := repository.NewSubmissionRepo(tx).ResetForRejudge(ctx, []int64{subs[1].ID}); err != nil {
		t.Fatal(err)
	}
	type result struct {
		notes []ratingSettlement
		err   error
	}
	started, done := make(chan struct{}), make(chan result, 1)
	go func() {
		close(started)
		notes, err := runRatingSettlement(ctx, db, contest.ID)
		done <- result{notes, err}
	}()
	<-started
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	got := <-done
	if got.err != nil || len(got.notes) != 0 {
		t.Fatalf("settlement after rejudge = %v, %v", got.notes, got.err)
	}
	var c model.Contest
	if err := db.First(&c, contest.ID).Error; err != nil || c.Settled {
		t.Fatalf("pending contest settled: %+v, %v", c, err)
	}
}

func TestContestSubmissionRechecksDeadlineAtPersistence(t *testing.T) {
	db := newRatingTestDB(t)
	contest, _ := seedRatingContest(t, db, judge.Accepted)
	repo := repository.NewSubmissionRepo(db)
	// 模拟截止前通过校验，但延迟到结束后才尝试落库的请求。
	sub := model.Submission{PublicID: 3, ContestID: contest.ID, UserID: 1, ProblemID: 1,
		Status: judge.Pending, CreatedAt: contest.EndTime.Add(-time.Second)}
	if _, err := repo.CreatePending(t.Context(), &sub); !errors.Is(err, repository.ErrContestSubmissionClosed) {
		t.Fatalf("late submission error = %v", err)
	}
	var count int64
	if err := db.Model(&model.Submission{}).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("late submission persisted: %d, %v", count, err)
	}
	if err := db.Model(&model.SubmissionOutbox{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("late outbox persisted: %d, %v", count, err)
	}
	if err := db.Model(&model.Contest{}).Where("id = ?", contest.ID).Update("end_time", time.Now().Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreatePending(t.Context(), &sub); err != nil {
		t.Fatalf("open contest submission: %v", err)
	}
	if err := db.Model(&model.SubmissionOutbox{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("open contest outbox = %d, %v", count, err)
	}
}

func TestRejudgeRejectsSettlementCommittedWhileWaitingForContestLock(t *testing.T) {
	db := newRatingTestDB(t)
	contest, subs := seedRatingContest(t, db, judge.Accepted)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	if notes, err := runRatingSettlement(ctx, tx, contest.ID); err != nil || len(notes) != 2 {
		t.Fatalf("settlement = %v, %v", notes, err)
	}
	started, done := make(chan struct{}), make(chan error, 1)
	go func() {
		close(started)
		_, err := repository.NewSubmissionRepo(db).ResetForRejudge(ctx, []int64{subs[1].ID})
		done <- err
	}()
	<-started
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, repository.ErrContestRatingSettled) {
		t.Fatalf("rejudge after concurrent settlement = %v", err)
	}
	var sub model.Submission
	if err := db.First(&sub, subs[1].ID).Error; err != nil || sub.Status != judge.Accepted || sub.Version != 0 {
		t.Fatalf("settled result changed: %+v, %v", sub, err)
	}
}
