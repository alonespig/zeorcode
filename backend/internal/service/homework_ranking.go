package service

import (
	"context"
	"sort"
	"time"

	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
)

func (s *HomeworkService) Rank(ctx context.Context, id, userID int64, isSiteAdmin bool) (*HomeworkRank, error) {
	hw, access, err := s.loadWithAccess(ctx, id, userID, isSiteAdmin)
	if err != nil {
		return nil, err
	}
	if !access.CanView() {
		return nil, errcode.ErrPermissionDenied.WithMsg("仅团队成员可查看")
	}

	refs, err := s.repo.ProblemRefsByHomeworkIDs(ctx, []int64{id})
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp := &HomeworkRank{
		StartTime:  hw.StartTime,
		EndTime:    hw.EndTime,
		ProblemIDs: make([]string, 0, len(refs)),
		TotalScore: len(refs) * model.HomeworkProblemFullScore,
		List:       []HomeworkRankRow{},
	}
	if len(refs) == 0 {
		return resp, nil
	}

	problemIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		problemIDs = append(problemIDs, ref.ProblemID)
	}
	problems, err := s.problemRepo.FindByIDs(ctx, problemIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	displayByID := make(map[int64]string, len(problems))
	for _, p := range problems {
		displayByID[p.ID] = p.DisplayID
	}
	// 列头顺序与作业题目顺序一致
	for _, ref := range refs {
		if d, ok := displayByID[ref.ProblemID]; ok {
			resp.ProblemIDs = append(resp.ProblemIDs, d)
		}
	}

	best, err := s.repo.BestScores(ctx, id, hw.StartTime, hw.EndTime)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}

	// 排行榜以当前团队成员为基准：没有提交的人也显示 0 分；
	// 退队的人即使有历史成绩，也不再占榜位。
	members, err := s.teamRepo.ListMembers(ctx, hw.TeamID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}
	if len(userIDs) == 0 {
		return resp, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	resp.List = buildHomeworkRankRows(refs, displayByID, members, users, best)
	return resp, nil
}

type homeworkRankAggregate struct {
	total         int
	solved        int
	scores        map[int64]int
	achievedAt    time.Time // 各题达成最高分的最晚时刻，用于同分排序
	hasSubmission bool
}

type homeworkRankEntry struct {
	row           HomeworkRankRow
	achievedAt    time.Time
	hasSubmission bool
}

func buildHomeworkRankRows(
	refs []repository.HomeworkProblemRef,
	displayByID map[int64]string,
	members []model.TeamMember,
	users []model.User,
	best []repository.HomeworkBestScore,
) []HomeworkRankRow {
	byUser := make(map[int64]*homeworkRankAggregate, len(members))
	for _, member := range members {
		byUser[member.UserID] = &homeworkRankAggregate{scores: make(map[int64]int, len(refs))}
	}
	for _, score := range best {
		a, isMember := byUser[score.UserID]
		if !isMember {
			continue
		}
		a.hasSubmission = true
		a.scores[score.ProblemID] = score.Best
		a.total += score.Best
		if score.Best >= model.HomeworkProblemFullScore {
			a.solved++
		}
		if score.AchievedAt.After(a.achievedAt) {
			a.achievedAt = score.AchievedAt
		}
	}

	userByID := make(map[int64]model.User, len(users))
	for _, user := range users {
		userByID[user.ID] = user
	}
	entries := make([]homeworkRankEntry, 0, len(members))
	for _, member := range members {
		user, exists := userByID[member.UserID]
		if !exists {
			continue
		}
		a := byUser[member.UserID]
		row := HomeworkRankRow{
			UID:         user.UID,
			Username:    user.Username,
			StudentNo:   studentNoValue(user.StudentNo),
			RealName:    user.RealName,
			Avatar:      user.Avatar,
			Gender:      user.Gender,
			TotalScore:  a.total,
			SolvedCount: a.solved,
			Cells:       make([]RankCell, 0, len(refs)),
		}
		for _, ref := range refs {
			display, exists := displayByID[ref.ProblemID]
			if !exists {
				continue
			}
			score, submitted := a.scores[ref.ProblemID]
			row.Cells = append(row.Cells, RankCell{
				ProblemID: display,
				Score:     score,
				Solved:    score >= model.HomeworkProblemFullScore,
				Submitted: submitted,
			})
		}
		entries = append(entries, homeworkRankEntry{
			row:           row,
			achievedAt:    a.achievedAt,
			hasSubmission: a.hasSubmission,
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].row.TotalScore != entries[j].row.TotalScore {
			return entries[i].row.TotalScore > entries[j].row.TotalScore
		}
		// 同分时有提交者在完全未提交者之前；两人都有提交时再比较达成时间。
		if entries[i].hasSubmission != entries[j].hasSubmission {
			return entries[i].hasSubmission
		}
		if entries[i].hasSubmission && !entries[i].achievedAt.Equal(entries[j].achievedAt) {
			return entries[i].achievedAt.Before(entries[j].achievedAt)
		}
		if entries[i].row.StudentNo != entries[j].row.StudentNo {
			return entries[i].row.StudentNo < entries[j].row.StudentNo
		}
		return entries[i].row.UID < entries[j].row.UID
	})

	rows := make([]HomeworkRankRow, 0, len(entries))
	for i := range entries {
		entries[i].row.Rank = i + 1
		rows = append(rows, entries[i].row)
	}
	return rows
}

// Submissions 作业提交列表。
// 普通成员只看自己（忽略 uid 筛选）；团队管理员及以上可看全部并按成员筛选。
