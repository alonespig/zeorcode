//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newDeadlineReminderTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	admin := openSubmissionOutboxTestDB(t)
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = adminSQL.Close() })
	name := fmt.Sprintf("zeorcode_reminder_test_%d", time.Now().UnixNano())
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
	if err := db.AutoMigrate(&model.Team{}, &model.TeamMember{}, &model.Homework{}, &model.HomeworkProblem{},
		&model.Problem{}, &model.Submission{}, &model.Notification{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestHomeworkDeadlineReminderDelivery(t *testing.T) {
	db := newDeadlineReminderTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	cutoff := now.Add(72 * time.Hour)
	start, end := now.Add(-24*time.Hour), now.Add(48*time.Hour)
	exec := func(query string, args ...any) {
		t.Helper()
		if err := db.Exec(query, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	// Simulate a local DB with the separately applied notification content extension.
	exec(`ALTER TABLE notifications MODIFY COLUMN content VARCHAR(2000) DEFAULT NULL`)
	longContent := strings.Repeat("a", 300)
	exec(`INSERT INTO notifications (recipient_id, type, content) VALUES (1, 'system', ?), (2, 'system', ?)`, longContent, longContent)
	migration, err := os.ReadFile("../../database/migrations/20261117_add_notification_dedupe_key.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(string(migration))
	exec(`INSERT INTO notifications (recipient_id, type, content) VALUES (1, 'system', ?)`, longContent)

	exec(`INSERT INTO teams (id, public_id, name) VALUES (10, 1000, '测试团队')`)
	exec(`INSERT INTO team_members (team_id, user_id, role) VALUES (10, 1, 0), (10, 2, 0), (10, 3, 0), (10, 4, 1), (10, 5, 2)`)
	exec(`INSERT INTO homeworks (id, public_id, team_id, title, start_time, end_time)
		VALUES (1, 2000, 10, '图论练习', ?, ?), (2, 2001, 10, '未开始', ?, ?),
		(3, 2002, 10, '恰好三天', ?, ?), (4, 2003, 10, '已截止', ?, ?), (5, 2004, 10, '超过三天', ?, ?)`,
		start, end, now.Add(time.Hour), end, start, cutoff, start, now, start, cutoff.Add(time.Minute))
	exec(`INSERT INTO problems (id, display_id) VALUES (1, 'P1'), (2, 'P2'), (3, 'P3')`)
	exec(`INSERT INTO homework_problems (homework_id, problem_id) VALUES (1, 1), (1, 2)`)
	exec(`INSERT INTO submissions (public_id, user_id, homework_id, problem_id, score, status, code, language, created_at)
		VALUES (100, 1, 1, 1, 30, 0, '', '', ?), (101, 1, 1, 1, 80, 0, '', '', ?),
		(102, 1, 1, 2, 20, 0, '', '', ?), (103, 1, 1, 3, 100, 0, '', '', ?),
		(104, 1, 2, 2, 100, 0, '', '', ?), (105, 1, 1, 2, 100, 0, '', '', ?),
		(106, 2, 1, 1, 100, 0, '', '', ?), (107, 2, 1, 2, 100, 0, '', '', ?)`,
		now, now, now, now, now, start.Add(-time.Second), now, now)

	repo := NewHomeworkDeadlineReminderRepo(db)
	checkPending := func(want []int64) {
		t.Helper()
		ids, err := repo.PendingHomeworkIDs(ctx, now, cutoff)
		if err != nil || !slices.Equal(ids, want) {
			t.Fatalf("pending homework IDs = %v, %v; want %v", ids, err, want)
		}
	}
	checkDelivery := func(id, want int64) {
		t.Helper()
		count, err := repo.Deliver(ctx, id, now, cutoff)
		if err != nil || count != want {
			t.Fatalf("delivered homework %d: count=%d, err=%v; want %d", id, count, err, want)
		}
	}
	checkPending([]int64{1})
	for _, id := range []int64{2, 3, 4, 5} {
		checkDelivery(id, 0)
	}
	checkDelivery(1, 3)

	var result []model.Notification
	if err := db.Where("source_type = ?", "homework_deadline").Order("recipient_id").Find(&result).Error; err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Fatalf("got %d reminders, want 3", len(result))
	}
	for i, want := range []string{"100 / 200", "200 / 200", "0 / 200"} {
		n := result[i]
		if n.RecipientID != int64(i+1) || n.Type != "system" || n.IsRead || n.SourceID != 1 ||
			n.Title != "作业截止提醒：《图论练习》" || n.Link != "/team/1000/homework/2000/problems" ||
			n.Content != "当前得分："+want+" · 截止时间：2026-11-19 10:00" {
			t.Errorf("reminder %d = %+v, want score %s", i, n, want)
		}
	}

	// A new process must not send the same deadline again.
	repo = NewHomeworkDeadlineReminderRepo(db)
	checkPending(nil)
	checkDelivery(1, 0)
	// Recheck current data just before writing, including a rescheduled deadline and a member leaving.
	exec(`UPDATE homeworks SET end_time = ? WHERE id = 1`, end.Add(time.Hour))
	checkPending([]int64{1})
	exec(`DELETE FROM team_members WHERE team_id = 10 AND user_id = 3`)
	checkDelivery(1, 2)
	checkPending(nil)
	// Returning to a deadline that already had a reminder must not duplicate it.
	exec(`UPDATE homeworks SET end_time = ? WHERE id = 1`, end)
	checkPending(nil)
	checkDelivery(1, 0)
}
