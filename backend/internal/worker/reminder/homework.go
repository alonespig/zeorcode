package reminder

import (
	"context"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/internal/repository"
)

const (
	homeworkReminderLead     = 72 * time.Hour
	homeworkReminderInterval = time.Minute
	homeworkReminderTimeout  = 45 * time.Second
)

type HomeworkDeadlineWorker struct {
	repo *repository.HomeworkDeadlineReminderRepo
}

func NewHomeworkDeadlineWorker(repo *repository.HomeworkDeadlineReminderRepo) *HomeworkDeadlineWorker {
	return &HomeworkDeadlineWorker{repo: repo}
}

// Run scans immediately and then once a minute. One worker owns the scan, so
// ticks never start overlapping work.
func (w *HomeworkDeadlineWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(homeworkReminderInterval)
	defer ticker.Stop()
	for ctx.Err() == nil {
		scanCtx, cancel := context.WithTimeout(ctx, homeworkReminderTimeout)
		w.scan(scanCtx)
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *HomeworkDeadlineWorker) scan(ctx context.Context) {
	now := time.Now()
	ids, err := w.repo.PendingHomeworkIDs(ctx, now, now.Add(homeworkReminderLead))
	if err != nil {
		if ctx.Err() != context.Canceled {
			logger.Warnw("scan homework deadline reminders failed", "err", err)
		}
		return
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}
		now = time.Now()
		if _, err := w.repo.Deliver(ctx, id, now, now.Add(homeworkReminderLead)); err != nil && ctx.Err() != context.Canceled {
			logger.Warnw("send homework deadline reminders failed", "homework_id", id, "err", err)
		}
	}
}
