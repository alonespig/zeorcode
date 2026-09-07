package judgeworker

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"zoj/internal/infra/mq"
	"zoj/internal/model"
	"zoj/internal/repository"
)

type fakeSubmissionTaskStore struct {
	claimResult repository.DispatchClaimResult
	claimErr    error
	sub         *model.Submission
	loadErr     error
	complete    bool
	completeErr error
	loadCalls   int
	claims      []repository.SubmissionDispatch
	completed   []repository.SubmissionDispatch
}

func (f *fakeSubmissionTaskStore) ClaimDispatchForJudging(
	_ context.Context,
	dispatch repository.SubmissionDispatch,
	_ string,
) (repository.DispatchClaimResult, error) {
	f.claims = append(f.claims, dispatch)
	return f.claimResult, f.claimErr
}

func (f *fakeSubmissionTaskStore) GetSubmissionByID(context.Context, int64) (*model.Submission, error) {
	f.loadCalls++
	return f.sub, f.loadErr
}

func (f *fakeSubmissionTaskStore) CompleteDispatchForJudging(
	_ context.Context,
	dispatch repository.SubmissionDispatch,
	_ string,
) (bool, error) {
	f.completed = append(f.completed, dispatch)
	return f.complete, f.completeErr
}

type fakeSubmissionTaskAcker struct {
	err    error
	msgIDs []string
}

func (f *fakeSubmissionTaskAcker) AckSubmission(_ context.Context, msgID string) error {
	f.msgIDs = append(f.msgIDs, msgID)
	return f.err
}

func claimedTaskStore() *fakeSubmissionTaskStore {
	return &fakeSubmissionTaskStore{
		claimResult: repository.DispatchClaimed,
		sub:         &model.Submission{ID: 42, Version: 3},
		complete:    true,
	}
}

func versionedTask() mq.StreamTask {
	return mq.StreamTask{MsgID: "1-0", SubmissionID: 42, Version: 3, Versioned: true}
}

func TestHandleSubmissionTaskDoesNotAckOnClaimFailure(t *testing.T) {
	claimErr := errors.New("database unavailable")
	store := claimedTaskStore()
	store.claimErr = claimErr
	acker := &fakeSubmissionTaskAcker{}

	err := handleSubmissionTask(
		context.Background(), versionedTask(), store, acker,
		func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error) {
			t.Fatal("processor called after claim failure")
			return taskProcessCompleted, nil
		},
	)

	if !errors.Is(err, claimErr) {
		t.Fatalf("error = %v, want wrapped claim error", err)
	}
	if store.loadCalls != 0 || len(acker.msgIDs) != 0 {
		t.Fatalf("load calls/acked = %d/%v, want 0/none", store.loadCalls, acker.msgIDs)
	}
}

func TestHandleSubmissionTaskAcksTerminalClaimWithoutProcessing(t *testing.T) {
	for _, claim := range []repository.DispatchClaimResult{
		repository.DispatchDuplicate,
		repository.DispatchAlreadyCompleted,
		repository.DispatchStale,
		repository.DispatchSubmissionMissing,
	} {
		t.Run(fmt.Sprintf("claim_%d", claim), func(t *testing.T) {
			store := claimedTaskStore()
			store.claimResult = claim
			acker := &fakeSubmissionTaskAcker{}
			err := handleSubmissionTask(
				context.Background(), versionedTask(), store, acker,
				func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error) {
					t.Fatal("processor called for terminal claim")
					return taskProcessCompleted, nil
				},
			)
			if err != nil {
				t.Fatalf("handleSubmissionTask() error = %v", err)
			}
			if store.loadCalls != 0 || len(acker.msgIDs) != 1 {
				t.Fatalf("load calls/acked = %d/%v, want 0/[1-0]", store.loadCalls, acker.msgIDs)
			}
		})
	}
}

func TestHandleSubmissionTaskDoesNotAckOnLoadFailure(t *testing.T) {
	loadErr := errors.New("database unavailable")
	store := claimedTaskStore()
	store.loadErr = loadErr
	acker := &fakeSubmissionTaskAcker{}

	err := handleSubmissionTask(
		context.Background(), versionedTask(), store, acker,
		func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error) {
			t.Fatal("processor called after load failure")
			return taskProcessCompleted, nil
		},
	)

	if !errors.Is(err, loadErr) {
		t.Fatalf("error = %v, want wrapped load error", err)
	}
	if len(acker.msgIDs) != 0 {
		t.Fatalf("acked messages = %v, want none", acker.msgIDs)
	}
}

func TestHandleSubmissionTaskDoesNotAckOnProcessingFailure(t *testing.T) {
	processErr := errors.New("result was not persisted")
	store := claimedTaskStore()
	acker := &fakeSubmissionTaskAcker{}

	err := handleSubmissionTask(
		context.Background(), versionedTask(), store, acker,
		func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error) {
			return taskProcessDeferred, processErr
		},
	)

	if !errors.Is(err, processErr) {
		t.Fatalf("error = %v, want wrapped process error", err)
	}
	if len(store.completed) != 0 || len(acker.msgIDs) != 0 {
		t.Fatalf("completed/acked = %v/%v, want none", store.completed, acker.msgIDs)
	}
}

func TestHandleSubmissionTaskKeepsDeferredMessagePending(t *testing.T) {
	store := claimedTaskStore()
	acker := &fakeSubmissionTaskAcker{}

	err := handleSubmissionTask(
		context.Background(), versionedTask(), store, acker,
		func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error) {
			return taskProcessDeferred, nil
		},
	)
	if err != nil {
		t.Fatalf("handleSubmissionTask() error = %v", err)
	}
	if len(store.completed) != 0 || len(acker.msgIDs) != 0 {
		t.Fatalf("completed/acked = %v/%v, want none", store.completed, acker.msgIDs)
	}
}

func TestHandleSubmissionTaskCompletesBeforeAck(t *testing.T) {
	store := claimedTaskStore()
	acker := &fakeSubmissionTaskAcker{}
	processed := false

	err := handleSubmissionTask(
		context.Background(), versionedTask(), store, acker,
		func(_ context.Context, task mq.StreamTask, sub model.Submission) (taskProcessOutcome, error) {
			processed = task.Version == 3 && sub.ID == 42
			return taskProcessCompleted, nil
		},
	)

	if err != nil {
		t.Fatalf("handleSubmissionTask() error = %v", err)
	}
	if !processed {
		t.Fatal("submission was not processed")
	}
	want := repository.SubmissionDispatch{SubmissionID: 42, Version: 3}
	if len(store.completed) != 1 || store.completed[0] != want {
		t.Fatalf("completed dispatches = %v, want [%v]", store.completed, want)
	}
	if len(acker.msgIDs) != 1 || acker.msgIDs[0] != "1-0" {
		t.Fatalf("acked messages = %v, want [1-0]", acker.msgIDs)
	}
}

func TestHandleSubmissionTaskDoesNotAckWhenCompletionFails(t *testing.T) {
	completeErr := errors.New("completion update failed")
	store := claimedTaskStore()
	store.completeErr = completeErr
	acker := &fakeSubmissionTaskAcker{}

	err := handleSubmissionTask(
		context.Background(), versionedTask(), store, acker,
		func(context.Context, mq.StreamTask, model.Submission) (taskProcessOutcome, error) {
			return taskProcessCompleted, nil
		},
	)
	if !errors.Is(err, completeErr) {
		t.Fatalf("error = %v, want wrapped completion error", err)
	}
	if len(acker.msgIDs) != 0 {
		t.Fatalf("acked messages = %v, want none", acker.msgIDs)
	}
}
