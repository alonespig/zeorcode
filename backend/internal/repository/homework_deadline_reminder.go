package repository

import (
	"context"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/model"

	"gorm.io/gorm"
)

type HomeworkDeadlineReminderRepo struct {
	db *gorm.DB
}

func NewHomeworkDeadlineReminderRepo(db *gorm.DB) *HomeworkDeadlineReminderRepo {
	return &HomeworkDeadlineReminderRepo{db: db}
}

// Use the stored deadline (including its precision), so rescheduling allows a new reminder.
const deadlineReminderKey = `CONCAT('homework-deadline:', h.id, ':', DATE_FORMAT(h.end_time, '%Y%m%d%H%i%s%f'), ':', tm.user_id)`

const deadlineReminderWindow = `h.start_time <= ? AND h.end_time > ? AND h.end_time < ?`

func (r *HomeworkDeadlineReminderRepo) PendingHomeworkIDs(ctx context.Context, now, before time.Time) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).Table("homeworks AS h").
		Where(deadlineReminderWindow, now, now, before).
		Where(`EXISTS (
			SELECT 1 FROM team_members AS tm
			LEFT JOIN notifications AS sent ON sent.dedupe_key = `+deadlineReminderKey+`
			WHERE tm.team_id = h.team_id AND tm.role = ? AND tm.user_id > 0 AND sent.id IS NULL
		)`, model.TeamRoleMember).
		Order("h.end_time, h.id").Pluck("h.id", &ids).Error
	return ids, err
}

// Deliver inserts one assignment's reminders atomically. Membership, deadline and
// scores are read in the insert itself in case they changed since the scan.
func (r *HomeworkDeadlineReminderRepo) Deliver(ctx context.Context, homeworkID int64, now, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Exec(`
		INSERT INTO notifications
			(recipient_id, type, actor_id, title, content, link, source_type, source_id, created_at, dedupe_key)
		SELECT tm.user_id, 'system', 0, CONCAT('作业截止提醒：《', h.title, '》'),
			CONCAT('当前得分：', COALESCE(scores.current_score, 0), ' / ', totals.total_score,
				' · 截止时间：', DATE_FORMAT(h.end_time, '%Y-%m-%d %H:%i')),
			CONCAT('/team/', t.public_id, '/homework/', h.public_id, '/problems'),
			'homework_deadline', h.id, ?, `+deadlineReminderKey+`
		FROM homeworks AS h
		JOIN teams AS t ON t.id = h.team_id
		JOIN team_members AS tm ON tm.team_id = h.team_id AND tm.role = ? AND tm.user_id > 0
		LEFT JOIN notifications AS sent ON sent.dedupe_key = `+deadlineReminderKey+`
		LEFT JOIN (
			SELECT best.user_id, SUM(best.score) AS current_score
			FROM (
				SELECT s.user_id, s.problem_id, MAX(s.score) AS score
				FROM submissions AS s
				JOIN homeworks AS hw ON hw.id = s.homework_id
				JOIN homework_problems AS hp ON hp.homework_id = s.homework_id AND hp.problem_id = s.problem_id
				JOIN problems AS p ON p.id = hp.problem_id
				WHERE s.homework_id = ? AND s.created_at BETWEEN hw.start_time AND hw.end_time
				GROUP BY s.user_id, s.problem_id
			) AS best GROUP BY best.user_id
		) AS scores ON scores.user_id = tm.user_id
		CROSS JOIN (
			SELECT COUNT(*) * ? AS total_score FROM homework_problems WHERE homework_id = ?
		) AS totals
		WHERE h.id = ? AND `+deadlineReminderWindow+` AND sent.id IS NULL
		ON DUPLICATE KEY UPDATE dedupe_key = notifications.dedupe_key`,
		now, model.TeamRoleMember, homeworkID, model.HomeworkProblemFullScore, homeworkID, homeworkID, now, now, before)
	return result.RowsAffected, result.Error
}
