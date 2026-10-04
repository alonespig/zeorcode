package service

import (
	"context"
	"fmt"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/infra/logger"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/rating"

	"gorm.io/gorm"
)

// initialRating 首次参加 rated 比赛的起算基线（rating 为 0=未定级时按此计算）
const initialRating = 1200

// ratingSettleGrace 比赛结束后到允许 rating 结算的宽限期，见 ratingSettleDue
const ratingSettleGrace = time.Minute

func ratingSettleDue(contest model.Contest, now time.Time) bool {
	return !now.Before(contestEndTime(contest).Add(ratingSettleGrace))
}

// settleRatingIfNeeded 惰性结算：rated + 已结束超过宽限期 + 未结算 + 比赛内无 Pending 提交时，
// 按最终名次算 CF rating 落库（幂等）。结算只做一次，所以必须等所有提交都判完、名次定下来再算，
// 否则还在评测中的提交会被 contestStandings 排除，且之后无法补算。
func (s *ContestService) settleRatingIfNeeded(ctx context.Context, contestID int64) {
	contest, err := s.repo.GetContestByID(ctx, contestID)
	if err != nil || !contest.Rated || contest.Settled {
		return
	}
	if !ratingSettleDue(*contest, time.Now()) {
		return
	}
	var settled []ratingSettlement
	err = s.repo.SettleTx(ctx, contestID, func(tx *gorm.DB, c *model.Contest) error {
		var err error
		settled, err = settleContestRating(ctx, tx, c)
		return err
	})
	if err != nil {
		logger.S().Warnw("settle rating failed", "contestID", contestID, "err", err)
		return
	}
	if len(settled) == 0 {
		return
	}
	_ = s.cache.Delete(ctx, cache.ContestRank(contestID))

	// 结算后给每个参赛者发 rating 变化通知
	for _, n := range settled {
		s.notify.Notify(ctx, n.userID, 0, "rating",
			"Rating 结算 · "+contest.Name,
			fmt.Sprintf("%d → %d (%+d)", n.old, n.nw, n.delta),
			fmt.Sprintf("/contest/%d/rank", contest.PublicID), "contest", contestID)
	}
}

type ratingSettlement struct {
	userID  int64
	old, nw int
	delta   int
}

// settleContestRating 必须在持有比赛行锁的事务内调用。
// 检查 Pending、读取名次、更新 rating 使用同一个事务快照。
func settleContestRating(ctx context.Context, tx *gorm.DB, c *model.Contest) ([]ratingSettlement, error) {
	contestID := c.ID
	var settled []ratingSettlement
	if c.Settled || !c.Rated || !ratingSettleDue(*c, time.Now()) {
		return nil, nil
	}
	// 所有查询必须绑定当前事务，不能拿着行锁用事务外连接读榜单。
	reader := ContestService{repo: repository.NewContestRepo(tx), submitRepo: repository.NewSubmissionRepo(tx)}
	pending, err := reader.submitRepo.HasPendingContestSubmission(ctx, contestID)
	if err != nil || pending {
		return nil, err
	}
	ranks, err := reader.contestStandings(ctx, c)
	if err != nil || len(ranks) < 2 {
		return nil, err
	}
	userIDs := make([]int64, 0, len(ranks))
	for _, e := range ranks {
		userIDs = append(userIDs, e.UserID)
	}
	var users []model.User
	if err := tx.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	curRating := make(map[int64]int, len(users))
	curMax := make(map[int64]int, len(users))
	for _, u := range users {
		curRating[u.ID] = u.Rating
		curMax[u.ID] = u.MaxRating
	}

	players := make([]rating.Player, len(ranks))
	for i, e := range ranks {
		r := curRating[e.UserID]
		if r == 0 {
			r = initialRating
		}
		players[i] = rating.Player{Rating: r, Rank: e.Rank}
	}
	results := rating.Calculate(players)

	for i, e := range ranks {
		old := players[i].Rating
		nw := results[i].NewRating
		if err := tx.Create(&model.RatingChange{
			ContestID: contestID, UserID: e.UserID, Rank: e.Rank,
			OldRating: old, NewRating: nw, Delta: results[i].Delta,
		}).Error; err != nil {
			return nil, err
		}
		maxr := nw
		if curMax[e.UserID] > maxr {
			maxr = curMax[e.UserID]
		}
		if err := tx.Model(&model.User{}).Where("id = ?", e.UserID).
			Updates(map[string]any{"rating": nw, "max_rating": maxr}).Error; err != nil {
			return nil, err
		}
		settled = append(settled, ratingSettlement{userID: e.UserID, old: old, nw: nw, delta: results[i].Delta})
	}
	if err := tx.Model(&model.Contest{}).Where("id = ?", contestID).Update("settled", true).Error; err != nil {
		return nil, err
	}
	return settled, nil
}
