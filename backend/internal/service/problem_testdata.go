package service

import (
	"context"
	"fmt"

	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"
)

// ProblemTestDataReader is defined by the application services that need to
// expose or enforce local test-data readiness.
type ProblemTestDataReader interface {
	HasTestData(ctx context.Context, problemID int64) (bool, error)
}

// problemTestDataReady treats remote problems as ready because their test data
// is owned by the remote OJ rather than the local filesystem.
func problemTestDataReady(ctx context.Context, reader ProblemTestDataReader, problem *model.Problem) (bool, error) {
	if problem.OJ != "" {
		return true, nil
	}
	ready, err := reader.HasTestData(ctx, problem.ID)
	if err != nil {
		return false, errcode.ErrInternal.Wrap(fmt.Errorf("check test data for problem %d: %w", problem.ID, err))
	}
	return ready, nil
}

func requireProblemTestData(ctx context.Context, reader ProblemTestDataReader, problem *model.Problem) error {
	ready, err := problemTestDataReady(ctx, reader, problem)
	if err != nil {
		return err
	}
	if !ready {
		return errcode.ErrProblemTestDataMissing
	}
	return nil
}
