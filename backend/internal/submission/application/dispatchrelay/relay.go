// Package dispatchrelay 把数据库 Outbox 中的提交版本可靠投递到判题队列。
package dispatchrelay

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	maxOwnerLength = 128
	maxErrorLength = 1024
	maxBatchSize   = 1000
)

// Item 是 Relay 一次租约认领得到的最小投递信息。
type Item struct {
	OutboxID     int64
	SubmissionID int64
	Version      int
	Attempts     int
}

// Store 由 Relay 定义，只暴露 Outbox 投递所需的租约操作。
type Store interface {
	Claim(ctx context.Context, owner string, now, leaseUntil time.Time, limit int) ([]Item, error)
	MarkPublished(ctx context.Context, outboxID int64, owner string, publishedAt time.Time) (bool, error)
	MarkFailed(ctx context.Context, outboxID int64, owner string, availableAt time.Time, lastError string) (bool, error)
}

// Queue 是 Relay 面向判题队列的发送端口。
type Queue interface {
	EnqueueSubmission(ctx context.Context, submissionID int64, version int) error
}

type Config struct {
	BatchSize        int
	Concurrency      int
	PollInterval     time.Duration
	LeaseDuration    time.Duration
	OperationTimeout time.Duration
	RetryBase        time.Duration
	RetryMax         time.Duration
}

func DefaultConfig() Config {
	return Config{
		BatchSize:        100,
		Concurrency:      10,
		PollInterval:     time.Second,
		LeaseDuration:    2 * time.Minute,
		OperationTimeout: 5 * time.Second,
		RetryBase:        time.Second,
		RetryMax:         time.Minute,
	}
}

// Relay 不自行创建 goroutine；Run 的生命周期由 bootstrap 调用方持有。
type Relay struct {
	store Store
	queue Queue
	owner string
	cfg   Config
	now   func() time.Time
}

func New(store Store, queue Queue, owner string, cfg Config) (*Relay, error) {
	if store == nil {
		return nil, errors.New("dispatch relay store is nil")
	}
	if queue == nil {
		return nil, errors.New("dispatch relay queue is nil")
	}
	if owner == "" || len(owner) > maxOwnerLength {
		return nil, fmt.Errorf("dispatch relay owner length must be between 1 and %d", maxOwnerLength)
	}
	if cfg.BatchSize <= 0 || cfg.BatchSize > maxBatchSize {
		return nil, fmt.Errorf("dispatch relay batch size must be between 1 and %d", maxBatchSize)
	}
	if cfg.Concurrency <= 0 || cfg.Concurrency > cfg.BatchSize {
		return nil, errors.New("dispatch relay concurrency must be between 1 and batch size")
	}
	if cfg.PollInterval <= 0 || cfg.LeaseDuration <= 0 || cfg.OperationTimeout <= 0 {
		return nil, errors.New("dispatch relay intervals and timeouts must be positive")
	}
	if cfg.RetryBase <= 0 || cfg.RetryMax < cfg.RetryBase {
		return nil, errors.New("dispatch relay retry range is invalid")
	}
	waves := (cfg.BatchSize + cfg.Concurrency - 1) / cfg.Concurrency
	leaseFactor := time.Duration(waves * 2) // 每条最坏包含一次队列操作和一次状态写回
	if cfg.OperationTimeout > time.Duration(1<<63-1)/leaseFactor ||
		cfg.LeaseDuration < cfg.OperationTimeout*leaseFactor {
		return nil, errors.New("dispatch relay lease is shorter than the batch operation budget")
	}
	return &Relay{store: store, queue: queue, owner: owner, cfg: cfg, now: time.Now}, nil
}

// Run 立即执行一轮，之后在没有积压时按 PollInterval 轮询；ctx 取消后同步退出。
// report 只在循环边界接收内部错误，避免应用组件绑定具体日志实现。
func (r *Relay) Run(ctx context.Context, report func(error)) {
	for {
		claimed, err := r.DispatchOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil && report != nil {
			report(err)
		}
		if claimed == r.cfg.BatchSize {
			continue
		}

		timer := time.NewTimer(r.cfg.PollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}
	}
}

// DispatchOnce 租约认领一个有界批次并逐条投递。
// Redis 成功而数据库完成标记失败时，记录保持 Processing，租约到期后会安全重试；
// 重复队列消息由 worker 的 Submission 版本幂等状态机消化。
func (r *Relay) DispatchOnce(ctx context.Context) (int, error) {
	now := r.now()
	claimCtx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	items, err := r.store.Claim(claimCtx, r.owner, now, now.Add(r.cfg.LeaseDuration), r.cfg.BatchSize)
	cancel()
	if err != nil {
		return 0, fmt.Errorf("claim submission outbox: %w", err)
	}

	sem := make(chan struct{}, r.cfg.Concurrency)
	errCh := make(chan error, len(items))
	var wg sync.WaitGroup

dispatchLoop:
	for _, item := range items {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break dispatchLoop
		}
		wg.Add(1)
		go func(item Item) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := r.dispatch(ctx, item); err != nil {
				errCh <- err
			}
		}(item)
	}
	wg.Wait()
	close(errCh)
	dispatchErrors := make([]error, 0, len(errCh))
	for err := range errCh {
		dispatchErrors = append(dispatchErrors, err)
	}
	return len(items), errors.Join(dispatchErrors...)
}

func (r *Relay) dispatch(ctx context.Context, item Item) error {
	queueCtx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	err := r.queue.EnqueueSubmission(queueCtx, item.SubmissionID, item.Version)
	cancel()
	if err == nil {
		markCtx, markCancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
		_, markErr := r.store.MarkPublished(markCtx, item.OutboxID, r.owner, r.now())
		markCancel()
		if markErr != nil {
			return fmt.Errorf("mark outbox %d published: %w", item.OutboxID, markErr)
		}
		return nil
	}

	availableAt := r.now().Add(r.retryDelay(item.Attempts))
	markCtx, markCancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	_, markErr := r.store.MarkFailed(markCtx, item.OutboxID, r.owner, availableAt, errorSummary(err))
	markCancel()
	dispatchErr := fmt.Errorf("enqueue submission %d version %d: %w", item.SubmissionID, item.Version, err)
	if markErr != nil {
		return errors.Join(dispatchErr, fmt.Errorf("mark outbox %d failed: %w", item.OutboxID, markErr))
	}
	return dispatchErr
}

func (r *Relay) retryDelay(attempts int) time.Duration {
	delay := r.cfg.RetryBase
	for attempt := 1; attempt < attempts && delay < r.cfg.RetryMax; attempt++ {
		if delay > r.cfg.RetryMax/2 {
			return r.cfg.RetryMax
		}
		delay *= 2
	}
	if delay > r.cfg.RetryMax {
		return r.cfg.RetryMax
	}
	return delay
}

func errorSummary(err error) string {
	if err == nil {
		return ""
	}
	clean := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, err.Error())
	runes := []rune(clean)
	if len(runes) > maxErrorLength {
		runes = runes[:maxErrorLength]
	}
	return string(runes)
}
