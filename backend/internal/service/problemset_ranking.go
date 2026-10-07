package service

import (
	"context"
	"errors"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"

	"gorm.io/gorm"
)

// Rank 返回题单排行榜。通过数复用公共题库的 user_problems 状态，
// 因此用户在题单创建前通过的题目也会立即计入。
func (s *ProblemSetService) Rank(
	ctx context.Context,
	id int64,
	params ProblemSetRankParams,
	userID *int64,
	isAdmin bool,
) (*ProblemSetRank, error) {
	set, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrProblemSetNotFound
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if set.Published != model.ProblemSetPublished && !isAdmin {
		return nil, errcode.ErrProblemSetNotFound
	}

	unlocked, err := s.hasAccess(ctx, set, userID, isAdmin)
	if err != nil {
		return nil, err
	}
	if !unlocked {
		return nil, errcode.ErrProblemSetLocked
	}

	refs, err := s.repo.ProblemRefsBySetIDs(ctx, []int64{id})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	result := &ProblemSetRank{List: make([]ProblemSetRankItem, 0), ProblemIDs: []string{}}
	if len(refs) == 0 {
		if userID != nil {
			result.Mine, err = s.emptyProblemSetRankItem(ctx, *userID)
		}
		return result, err
	}

	problemInternalIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		problemInternalIDs = append(problemInternalIDs, ref.ProblemID)
	}
	problems, err := s.problemRepo.FindByIDs(ctx, problemInternalIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	displayByID := make(map[int64]string, len(problems))
	for _, problem := range problems {
		displayByID[problem.ID] = problem.DisplayID
	}
	orderedProblemIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		displayID, ok := displayByID[ref.ProblemID]
		if !ok {
			continue
		}
		orderedProblemIDs = append(orderedProblemIDs, ref.ProblemID)
		result.ProblemIDs = append(result.ProblemIDs, displayID)
	}
	result.ProblemCount = len(result.ProblemIDs)

	rows, total, err := s.repo.ListRank(ctx, id, params.Page, params.PageSize)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	result.Total = total
	result.List = make([]ProblemSetRankItem, 0, len(rows))
	pageUserIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		pageUserIDs = append(pageUserIDs, row.UserID)
	}
	statusRows, err := s.repo.RankStatuses(ctx, id, pageUserIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	statusByUserProblem := make(map[problemSetRankStatusKey]int, len(statusRows))
	for _, statusRow := range statusRows {
		statusByUserProblem[problemSetRankStatusKey{UserID: statusRow.UserID, ProblemID: statusRow.ProblemID}] = statusRow.Status
	}
	for _, row := range rows {
		item := problemSetRankItem(row, userID)
		item.Cells = buildProblemSetRankCells(row.UserID, orderedProblemIDs, displayByID, statusByUserProblem)
		result.List = append(result.List, item)
		if item.IsSelf {
			mine := item
			result.Mine = &mine
		}
	}

	if userID != nil && result.Mine == nil {
		result.Mine, err = s.problemSetMine(ctx, id, *userID)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

type problemSetRankStatusKey struct {
	UserID    int64
	ProblemID int64
}

func buildProblemSetRankCells(
	userID int64,
	problemIDs []int64,
	displayByID map[int64]string,
	statuses map[problemSetRankStatusKey]int,
) []ProblemSetRankCell {
	cells := make([]ProblemSetRankCell, 0, len(problemIDs))
	for _, problemID := range problemIDs {
		cell := ProblemSetRankCell{ProblemID: displayByID[problemID]}
		if status, ok := statuses[problemSetRankStatusKey{UserID: userID, ProblemID: problemID}]; ok && status != model.UserProblemUntried {
			statusCopy := status
			cell.Status = &statusCopy
		}
		cells = append(cells, cell)
	}
	return cells
}

func (s *ProblemSetService) problemSetMine(ctx context.Context, setID, userID int64) (*ProblemSetRankItem, error) {
	row, err := s.repo.GetUserRank(ctx, setID, userID)
	if err == nil {
		returnItem := problemSetRankItem(*row, &userID)
		return &returnItem, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return s.emptyProblemSetRankItem(ctx, userID)
}

func (s *ProblemSetService) emptyProblemSetRankItem(ctx context.Context, userID int64) (*ProblemSetRankItem, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return &ProblemSetRankItem{
		ID:       user.UID,
		Username: user.Username,
		Avatar:   avatarOr(user.Avatar),
		Gender:   user.Gender,
		IsSelf:   true,
	}, nil
}

func problemSetRankItem(row repository.ProblemSetRankRecord, userID *int64) ProblemSetRankItem {
	return ProblemSetRankItem{
		Rank:           row.RankIndex,
		ID:             row.UID,
		Username:       row.Username,
		Avatar:         avatarOr(row.Avatar),
		Gender:         row.Gender,
		SolvedCount:    row.SolvedCount,
		AttemptedCount: row.AttemptedCount,
		IsSelf:         userID != nil && row.UserID == *userID,
	}
}
