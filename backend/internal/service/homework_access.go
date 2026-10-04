package service

import (
	"context"
	"errors"

	"zoj/internal/model"
	"zoj/pkg/errcode"

	"gorm.io/gorm"
)

func (s *HomeworkService) teamAccess(ctx context.Context, teamID, userID int64, isSiteAdmin bool) (TeamAccess, error) {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TeamAccess{}, errcode.ErrTeamNotFound
		}
		return TeamAccess{}, errcode.ErrDatabase.Wrap(err)
	}
	return s.teamSrv.Access(ctx, teamID, userID, isSiteAdmin)
}

// loadWithAccess 取作业并算出调用方在其所属团队中的权限。
// 非团队成员一律返回「作业不存在」，不泄露作业的存在。
func (s *HomeworkService) loadWithAccess(ctx context.Context, id, userID int64, isSiteAdmin bool) (*model.Homework, TeamAccess, error) {
	hw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, TeamAccess{}, errcode.ErrHomeworkNotFound
		}
		return nil, TeamAccess{}, errcode.ErrDatabase.Wrap(err)
	}
	access, err := s.teamSrv.Access(ctx, hw.TeamID, userID, isSiteAdmin)
	if err != nil {
		return nil, TeamAccess{}, err
	}
	if !access.CanView() {
		return nil, TeamAccess{}, errcode.ErrHomeworkNotFound
	}
	return hw, access, nil
}

// resolveProblems 把对外题号列表转成关联行，数组下标即 Sort。
