package service

import (
	"context"
	"time"

	"zoj/internal/model"
	"zoj/pkg/errcode"
)

func contestProblemAccessPolicy(contest *model.Contest, registered, isAdmin bool, now time.Time) error {
	if isAdmin {
		return nil
	}
	if now.Before(contest.StartTime) {
		return errcode.ErrContestNotStart
	}
	if !registered {
		return errcode.ErrContestNotRegistered
	}
	return nil
}

func contestSubmissionPolicy(contest *model.Contest, registered bool, now time.Time) error {
	if now.Before(contest.StartTime) {
		return errcode.ErrContestNotStart
	}
	if !now.Before(contestEndTime(*contest)) {
		return errcode.ErrContestFinished
	}
	if !registered {
		return errcode.ErrContestNotRegistered
	}
	return nil
}

func (s *ContestService) authorizeContestProblemAccess(
	ctx context.Context,
	contestID, userID int64,
	isAdmin bool,
) (*model.Contest, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		return contest, nil
	}
	registered, err := s.repo.ExistByID(ctx, contestID, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if err := contestProblemAccessPolicy(contest, registered, false, time.Now()); err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *ContestService) authorizeContestSubmission(
	ctx context.Context,
	contestID, userID int64,
) (*model.Contest, error) {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	registered, err := s.repo.ExistByID(ctx, contestID, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if err := contestSubmissionPolicy(contest, registered, time.Now()); err != nil {
		return nil, err
	}
	return contest, nil
}
