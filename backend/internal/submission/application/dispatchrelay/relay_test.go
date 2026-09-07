package dispatchrelay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	claimFn         func(context.Context, string, time.Time, time.Time, int) ([]Item, error)
	markPublishedFn func(context.Context, int64, string, time.Time) (bool, error)
	markFailedFn    func(context.Context, int64, string, time.Time, string) (bool, error)
}

func (f *fakeStore) Claim(ctx context.Context, owner string, now, leaseUntil time.Time, limit int) ([]Item, error) {
	return f.claimFn(ctx, owner, now, leaseUntil, limit)
}

func (f *fakeStore) MarkPublished(ctx context.Context, id int64, owner string, at time.Time) (bool, error) {
	return f.markPublishedFn(ctx, id, owner, at)
}

func (f *fakeStore) MarkFailed(ctx context.Context, id int64, owner string, at time.Time, lastError string) (bool, error) {
	return f.markFailedFn(ctx, id, owner, at, lastError)
}

type fakeQueue struct {
	enqueueFn func(context.Context, int64, int) error
}

func (f *fakeQueue) EnqueueSubmission(ctx context.Context, submissionID int64, version int) error {
	return f.enqueueFn(ctx, submissionID, version)
}

func newTestRelay(t *testing.T, store Store, queue Queue) *Relay {
	t.Helper()
	relay, err := New(store, queue, "relay-test", DefaultConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return relay
}

func TestDispatchOncePublishesClaimedItem(t *testing.T) {
	fixed := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{}
	queue := &fakeQueue{}
	store.claimFn = func(_ context.Context, owner string, now, leaseUntil time.Time, limit int) ([]Item, error) {
		if owner != "relay-test" || !now.Equal(fixed) || !leaseUntil.Equal(fixed.Add(2*time.Minute)) || limit != 100 {
			t.Fatalf("claim args = %q, %v, %v, %d", owner, now, leaseUntil, limit)
		}
		return []Item{{OutboxID: 7, SubmissionID: 42, Version: 3, Attempts: 1}}, nil
	}
	queue.enqueueFn = func(_ context.Context, submissionID int64, version int) error {
		if submissionID != 42 || version != 3 {
			t.Fatalf("enqueue args = %d/%d", submissionID, version)
		}
		return nil
	}
	store.markPublishedFn = func(_ context.Context, id int64, owner string, at time.Time) (bool, error) {
		if id != 7 || owner != "relay-test" || !at.Equal(fixed) {
			t.Fatalf("mark published args = %d/%q/%v", id, owner, at)
		}
		return true, nil
	}
	store.markFailedFn = func(context.Context, int64, string, time.Time, string) (bool, error) {
		t.Fatal("MarkFailed called after successful enqueue")
		return false, nil
	}

	relay := newTestRelay(t, store, queue)
	relay.now = func() time.Time { return fixed }
	claimed, err := relay.DispatchOnce(context.Background())
	if err != nil || claimed != 1 {
		t.Fatalf("DispatchOnce() = %d, %v; want 1, nil", claimed, err)
	}
}

func TestDispatchOnceReleasesFailedItemWithBackoff(t *testing.T) {
	fixed := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	queueErr := errors.New("redis\nunavailable")
	store := &fakeStore{}
	store.claimFn = func(context.Context, string, time.Time, time.Time, int) ([]Item, error) {
		return []Item{{OutboxID: 7, SubmissionID: 42, Version: 3, Attempts: 4}}, nil
	}
	store.markPublishedFn = func(context.Context, int64, string, time.Time) (bool, error) {
		t.Fatal("MarkPublished called after failed enqueue")
		return false, nil
	}
	store.markFailedFn = func(_ context.Context, id int64, owner string, at time.Time, lastError string) (bool, error) {
		if id != 7 || owner != "relay-test" {
			t.Fatalf("mark failed args = %d/%q", id, owner)
		}
		if want := fixed.Add(8 * time.Second); !at.Equal(want) {
			t.Fatalf("availableAt = %v, want %v", at, want)
		}
		if lastError != "redis unavailable" {
			t.Fatalf("lastError = %q", lastError)
		}
		return true, nil
	}
	queue := &fakeQueue{enqueueFn: func(context.Context, int64, int) error { return queueErr }}

	relay := newTestRelay(t, store, queue)
	relay.now = func() time.Time { return fixed }
	claimed, err := relay.DispatchOnce(context.Background())
	if claimed != 1 || !errors.Is(err, queueErr) {
		t.Fatalf("DispatchOnce() = %d, %v; want claimed item and queue error", claimed, err)
	}
}

func TestDispatchOnceKeepsLeaseWhenPublishedMarkFails(t *testing.T) {
	markErr := errors.New("database unavailable")
	store := &fakeStore{
		claimFn: func(context.Context, string, time.Time, time.Time, int) ([]Item, error) {
			return []Item{{OutboxID: 7, SubmissionID: 42}}, nil
		},
		markPublishedFn: func(context.Context, int64, string, time.Time) (bool, error) {
			return false, markErr
		},
		markFailedFn: func(context.Context, int64, string, time.Time, string) (bool, error) {
			t.Fatal("MarkFailed must not release a message already written to Redis")
			return false, nil
		},
	}
	queue := &fakeQueue{enqueueFn: func(context.Context, int64, int) error { return nil }}
	relay := newTestRelay(t, store, queue)

	_, err := relay.DispatchOnce(context.Background())
	if !errors.Is(err, markErr) {
		t.Fatalf("error = %v, want mark error", err)
	}
}

func TestDispatchOnceBoundsConcurrentQueueCalls(t *testing.T) {
	items := make([]Item, 6)
	for i := range items {
		items[i] = Item{OutboxID: int64(i + 1), SubmissionID: int64(i + 10)}
	}
	store := &fakeStore{
		claimFn: func(context.Context, string, time.Time, time.Time, int) ([]Item, error) {
			return items, nil
		},
		markPublishedFn: func(context.Context, int64, string, time.Time) (bool, error) { return true, nil },
		markFailedFn:    func(context.Context, int64, string, time.Time, string) (bool, error) { return false, nil },
	}
	started := make(chan struct{}, len(items))
	release := make(chan struct{})
	queue := &fakeQueue{enqueueFn: func(context.Context, int64, int) error {
		started <- struct{}{}
		<-release
		return nil
	}}
	cfg := DefaultConfig()
	cfg.BatchSize = len(items)
	cfg.Concurrency = 2
	cfg.OperationTimeout = time.Second
	cfg.LeaseDuration = 10 * time.Second
	relay, err := New(store, queue, "relay-test", cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := relay.DispatchOnce(context.Background())
		done <- err
	}()
	<-started
	<-started
	select {
	case <-started:
		t.Fatal("more queue calls started than configured concurrency")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("DispatchOnce() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("DispatchOnce did not finish")
	}
}

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	store := &fakeStore{
		claimFn: func(ctx context.Context, _ string, _, _ time.Time, _ int) ([]Item, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
		markPublishedFn: func(context.Context, int64, string, time.Time) (bool, error) { return false, nil },
		markFailedFn:    func(context.Context, int64, string, time.Time, string) (bool, error) { return false, nil },
	}
	queue := &fakeQueue{enqueueFn: func(context.Context, int64, int) error { return nil }}
	relay := newTestRelay(t, store, queue)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		relay.Run(ctx, nil)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after cancellation")
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	store := &fakeStore{}
	queue := &fakeQueue{}
	shortLease := DefaultConfig()
	shortLease.LeaseDuration = time.Second
	tests := []struct {
		name  string
		owner string
		cfg   Config
	}{
		{name: "empty owner", cfg: DefaultConfig()},
		{name: "owner too long", owner: strings.Repeat("x", maxOwnerLength+1), cfg: DefaultConfig()},
		{name: "empty config", owner: "relay", cfg: Config{}},
		{name: "short lease", owner: "relay", cfg: shortLease},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(store, queue, tt.owner, tt.cfg); err == nil {
				t.Fatal("New() error = nil")
			}
		})
	}
}
