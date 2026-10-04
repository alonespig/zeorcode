package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"zoj/internal/infra/cache"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/errcode"
	"zoj/pkg/judge"
	"zoj/pkg/scoring"
)

// contestRankTTL 排行榜缓存有效期：轮询期间命中缓存，最多滞后这么久
const contestRankTTL = 5 * time.Second

func markSelf(resp *ContestRank, viewerUID int64) {
	if viewerUID == 0 {
		return
	}
	for i := range resp.RankList {
		resp.RankList[i].IsSelf = resp.RankList[i].User.ID == viewerUID
	}
}

// uidOf 内部主键 → 对外用户号（0=未登录或查不到）
func (s *ContestService) uidOf(ctx context.Context, userPK int64) int64 {
	if userPK <= 0 {
		return 0
	}
	if u, err := s.userRepo.GetUserByID(ctx, userPK); err == nil {
		return u.UID
	}
	return 0
}

// GetMyContestRank 当前用户在某场比赛的名次（没参与返回 Found=false）
func (s *ContestService) GetMyContestRank(ctx context.Context, contestID, userID int64) (*MyContestRank, error) {
	rank, err := s.GetContestRank(ctx, contestID, userID)
	if err != nil {
		return nil, err
	}
	for _, item := range rank.RankList {
		if item.IsSelf { // GetContestRank 已按查看者标好 isSelf
			return &MyContestRank{
				Found:      true,
				Rank:       item.Rank,
				PassCount:  item.PassCount,
				Penalty:    item.Penalty,
				TotalScore: item.TotalScore,
				Rule:       rank.Contest.Rule,
			}, nil
		}
	}
	return &MyContestRank{Found: false, Rule: rank.Contest.Rule}, nil
}

type standingEntry struct {
	UserID int64
	Rank   int
}

// ratingSettleDue 保留一分钟结算缓冲；并发正确性由比赛行锁和锁内检查保证。

// contestStandings 计算比赛最终名次（仅有提交者，并列共享名次），供 rating 结算用。
// buildContestUCP 从一组「同一 user+problem、时间升序、已排除 Pending/CE」的提交明细，
// 重算出该 (contest,user,problem) 的聚合行。规则与 worker 的增量逻辑一一对应，保证结果一致：
//   - ACM / CF：Status/AcTime 取首个 AC；UnAcCount = 首 AC 之前的非 AC(非 CE)次数(罚时)；首 AC 后冻结。
//   - OI：取最后一次提交的 Score/Status/时间。
//   - IOI：取历史最高 Score 及其达成时间；Status 取最后一次。
//
// subs 不能为空。
func buildContestUCP(contestType model.ContestType, contestID, userID, problemID int64, subs []repository.RecomputeSubmissionRow) model.UserContestProblem {
	ucp := model.UserContestProblem{
		ContestID: contestID,
		UserID:    userID,
		ProblemID: problemID,
		SubCount:  len(subs),
	}
	switch contestType {
	case model.ContestOI:
		for _, r := range subs {
			if r.Status == judge.Accepted {
				ucp.AcCount++
			} else {
				ucp.UnAcCount++
			}
		}
		last := subs[len(subs)-1]
		ucp.Status = last.Status
		ucp.Score = last.Score
		t := last.CreatedAt
		ucp.AcTime = &t
	case model.ContestIOI:
		maxScore := -1
		var maxTime time.Time
		for _, r := range subs {
			if r.Status == judge.Accepted {
				ucp.AcCount++
			} else {
				ucp.UnAcCount++
			}
			if r.Score > maxScore {
				maxScore = r.Score
				maxTime = r.CreatedAt
			}
		}
		if maxScore < 0 {
			maxScore = 0
		}
		ucp.Score = maxScore
		ucp.Status = subs[len(subs)-1].Status
		mt := maxTime
		ucp.AcTime = &mt
	default: // ACM / CF
		for _, r := range subs {
			if ucp.Status == judge.Accepted {
				break // 首个 AC 后冻结：罚时/状态不再变
			}
			if r.Status == judge.Accepted {
				ucp.Status = judge.Accepted
				ucp.AcCount = 1
				t := r.CreatedAt
				ucp.AcTime = &t
			} else {
				ucp.Status = r.Status
				ucp.UnAcCount++
			}
		}
	}
	return ucp
}

// computeContestUserProblems 路①（对齐 HOJ）：从 submissions 明细现算某比赛的全部 (user,problem) 聚合，不落库。
// 榜单 / 结算 / 重算都以它为唯一真值来源 —— 天然反映最新判题结果（含重判后），无需任何"重算触发"。
func (s *ContestService) computeContestUserProblems(ctx context.Context, contest *model.Contest) ([]model.UserContestProblem, error) {
	rows, err := s.submitRepo.ListContestSubmissionsForRecompute(ctx, contest.ID)
	if err != nil {
		return nil, err
	}
	// 按 (user,problem) 分组（rows 已按 user,problem,time 升序）
	type key struct{ u, p int64 }
	groups := make(map[key][]repository.RecomputeSubmissionRow)
	order := make([]key, 0)
	for _, r := range rows {
		k := key{r.UserID, r.ProblemID}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], r)
	}
	ucps := make([]model.UserContestProblem, 0, len(order))
	for _, k := range order {
		ucps = append(ucps, buildContestUCP(contest.Type, contest.ID, k.u, k.p, groups[k]))
	}
	return ucps, nil
}

// RecomputeContest 把现算结果物化落库到 UserContestProblem，并清相关缓存。
// 路①下榜单已直接现算、不依赖此表；保留该接口用于手动重建/导出/排查（admin 触发）。
func (s *ContestService) RecomputeContest(ctx context.Context, contestID int64) error {
	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return err
	}
	ucps, err := s.computeContestUserProblems(ctx, contest)
	if err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	if err := s.repo.ReplaceContestUserProblems(ctx, contestID, ucps); err != nil {
		return errcode.ErrDatabase.Wrap(err)
	}
	s.invalidateContestCaches(ctx, contestID, ucps)
	return nil
}

// invalidateContestCaches 清掉比赛的 rank 缓存 + 各覆盖缓存，让读取回落到刚重算好的 DB。
func (s *ContestService) invalidateContestCaches(ctx context.Context, contestID int64, ucps []model.UserContestProblem) {
	_ = s.cache.Delete(ctx, cache.ContestRank(contestID))
	problems := make(map[int64]struct{})
	users := make(map[int64]struct{})
	for _, u := range ucps {
		problems[u.ProblemID] = struct{}{}
		users[u.UserID] = struct{}{}
		_ = s.cache.Delete(ctx, cache.ContestUserProblem(contestID, u.UserID, u.ProblemID))
	}
	for pid := range problems {
		_ = s.cache.Delete(ctx, cache.ContestProblemStats(contestID, pid))
		_ = s.cache.Delete(ctx, cache.ContestProblemSubmittedUsers(contestID, pid))
		_ = s.cache.Delete(ctx, cache.ContestProblemAcceptedUsers(contestID, pid))
	}
	for uid := range users {
		_ = s.cache.Delete(ctx, cache.ContestUserStats(contestID, uid))
		_ = s.cache.Delete(ctx, cache.ContestUserAcceptedProblems(contestID, uid))
	}
}

func (s *ContestService) contestStandings(ctx context.Context, contest *model.Contest) ([]standingEntry, error) {
	contestUsers, err := s.repo.GetContestUserByContestID(ctx, contest.ID)
	if err != nil {
		return nil, err
	}
	records, err := s.computeContestUserProblems(ctx, contest)
	if err != nil {
		return nil, err
	}
	byUser := make(map[int64][]model.UserContestProblem)
	for _, r := range records {
		byUser[r.UserID] = append(byUser[r.UserID], r)
	}
	acm := contest.Type != model.ContestOI && contest.Type != model.ContestIOI && contest.Type != model.ContestCF
	isCF := contest.Type == model.ContestCF
	cfDur := int(contestEndTime(*contest).Sub(contest.StartTime).Minutes())
	cpScore := map[int64]int{} // CF：题目初始分 x
	if isCF {
		if cps, e := s.repo.GetContestProblemsByContestID(ctx, contest.ID); e == nil {
			for _, cp := range cps {
				cpScore[cp.ProblemID] = cp.Score
			}
		}
	}

	type row struct {
		userID   int64
		pass     int
		penalty  int
		score    int
		lastTime int64
	}
	rows := make([]row, 0, len(contestUsers))
	for _, cu := range contestUsers {
		recs := byUser[cu.UserID]
		if len(recs) == 0 {
			continue // 只算有提交者
		}
		rr := row{userID: cu.UserID}
		for _, rec := range recs {
			if acm {
				if rec.Status == judge.Accepted && rec.AcTime != nil {
					rr.pass++
					rr.penalty += int(rec.AcTime.Sub(contest.StartTime).Minutes()) + 20*rec.UnAcCount
				}
			} else if isCF {
				accepted := rec.Status == judge.Accepted && rec.AcTime != nil
				t := 0
				if rec.AcTime != nil {
					t = int(rec.AcTime.Sub(contest.StartTime).Minutes())
				}
				rr.score += scoring.CFScore(cpScore[rec.ProblemID], t, cfDur, rec.UnAcCount, accepted)
				if accepted && rec.AcTime.Unix() > rr.lastTime {
					rr.lastTime = rec.AcTime.Unix()
				}
			} else {
				rr.score += rec.Score
				if rec.AcTime != nil && rec.AcTime.Unix() > rr.lastTime {
					rr.lastTime = rec.AcTime.Unix()
				}
			}
		}
		rows = append(rows, rr)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if acm {
			if rows[i].pass != rows[j].pass {
				return rows[i].pass > rows[j].pass
			}
			return rows[i].penalty < rows[j].penalty
		}
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return rows[i].lastTime < rows[j].lastTime
	})

	out := make([]standingEntry, len(rows))
	for i := range rows {
		rank := i + 1
		if i > 0 {
			var eq bool
			if acm {
				eq = rows[i].pass == rows[i-1].pass && rows[i].penalty == rows[i-1].penalty
			} else {
				eq = rows[i].score == rows[i-1].score && rows[i].lastTime == rows[i-1].lastTime
			}
			if eq {
				rank = out[i-1].Rank
			}
		}
		out[i] = standingEntry{UserID: rows[i].userID, Rank: rank}
	}
	return out, nil
}

func (s *ContestService) GetContestRank(ctx context.Context, contestID, userID int64) (*ContestRank, error) {
	s.settleRatingIfNeeded(ctx, contestID)
	viewerUID := s.uidOf(ctx, userID) // 查看者的对外用户号，用于标记 isSelf
	cacheKey := cache.ContestRank(contestID)
	// 命中缓存直接返回（轮询期间不打 DB）；isSelf 按查看者临时标，不进缓存
	var cachedResp ContestRank
	if ok, err := s.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil && ok {
		markSelf(&cachedResp, viewerUID)
		return &cachedResp, nil
	}

	contest, err := s.getContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	contestProblems, err := s.repo.GetContestProblemsByContestID(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	problemIDs := make([]int64, 0, len(contestProblems))
	for _, problem := range contestProblems {
		problemIDs = append(problemIDs, problem.ProblemID)
	}

	sort.Slice(contestProblems, func(i, j int) bool {
		return contestProblems[i].Label < contestProblems[j].Label
	})

	contestUsers, err := s.repo.GetContestUserByContestID(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userIDs := make([]int64, 0, len(contestUsers))
	for _, user := range contestUsers {
		userIDs = append(userIDs, user.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	userMap := make(map[int64]model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	records, err := s.computeContestUserProblems(ctx, contest)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	recordMap := make(map[int64]map[int64]model.UserContestProblem, len(contestUsers))
	firstBloodMap := map[int64]int64{}
	firstBloodTime := map[int64]time.Time{}
	for _, record := range records {
		if recordMap[record.UserID] == nil {
			recordMap[record.UserID] = map[int64]model.UserContestProblem{}
		}
		recordMap[record.UserID][record.ProblemID] = record
		if record.Status == judge.Accepted && record.AcTime != nil {
			if t, ok := firstBloodTime[record.ProblemID]; !ok || record.AcTime.Before(t) {
				firstBloodTime[record.ProblemID] = *record.AcTime
				firstBloodMap[record.ProblemID] = record.UserID
			}
		}
	}

	resp := &ContestRank{
		Contest: ContestRankInfo{
			ID:     contest.PublicID,
			Title:  contest.Name,
			Status: int(contestStatus(*contest)),
		},
		Problems: make([]ContestProblemLabel, 0, len(contestProblems)),
		RankList: make([]ContestRankItem, 0, len(contestUsers)),
	}

	resp.Contest.Rule = contest.Type.String()
	for i := 0; i < len(contestProblems); i++ {
		resp.Problems = append(resp.Problems, ContestProblemLabel{
			Label:    contestProblems[i].Label,
			Color:    contestProblems[i].Color,
			MaxScore: contestProblems[i].Score,
		})
	}

	if contest.Type != model.ContestOI && contest.Type != model.ContestIOI && contest.Type != model.ContestCF {
		// ===== ACM（含未知/历史 type）：通过数 + 罚时 =====
		for _, user := range contestUsers {
			u := userMap[user.UserID]
			item := ContestRankItem{
				Problems: make([]ContestProblemStatus, 0, len(contestProblems)),
			}
			item.User.ID = u.UID // 对外用户号
			item.User.Name = u.Username
			item.User.Avatar = u.Avatar
			item.User.Rating = u.Rating

			for index, problem := range contestProblems {
				// 路①：records 已是从 submissions 现算的最新真值，无需再叠 Redis 覆盖缓存。
				userContestProblem, ok := recordMap[user.UserID][problem.ProblemID]
				if ok {
					item.Problems = append(item.Problems, ContestProblemStatus{
						Label:  problem.Label,
						Status: userContestProblem.Status,
						Tries:  userContestProblem.UnAcCount,
					})
					if userContestProblem.UserID == firstBloodMap[problem.ProblemID] {
						item.Problems[index].FirstBlood = 1
					}
					if userContestProblem.Status == judge.Accepted && userContestProblem.AcTime != nil {
						diff := userContestProblem.AcTime.Sub(contest.StartTime)
						hours := int(diff.Hours())
						minutes := int(diff.Minutes()) % 60
						timeStr := fmt.Sprintf("%02d:%02d", hours, minutes)
						item.PassCount++
						item.Problems[index].AcTime = &timeStr
						item.Penalty += int(userContestProblem.AcTime.Sub(contest.StartTime).Minutes()) + 20*userContestProblem.UnAcCount
					}
				} else {
					item.Problems = append(item.Problems, ContestProblemStatus{
						Label:  problem.Label,
						Status: 0,
						Tries:  0,
					})
				}
			}
			// PassCount/Penalty 直接用上面从 DB(recordMap) 算出的值；去掉 ContestUserStats 累加缓存覆盖(无 TTL、并发/重复消费会多加,漂移后永久污染榜单)。
			resp.RankList = append(resp.RankList, item)
		}

		sort.Slice(resp.RankList, func(i, j int) bool {
			if resp.RankList[i].PassCount != resp.RankList[j].PassCount {
				return resp.RankList[i].PassCount > resp.RankList[j].PassCount
			}
			return resp.RankList[i].Penalty < resp.RankList[j].Penalty
		})
	} else {
		// ===== OI / IOI：每题得分 + 总分（IOI 取最高分、OI 取最后一次，已在判题时落库）=====
		// OI 赛中封榜：只列参赛者、不公布分数，结束后再放开
		frozen := contest.Type == model.ContestOI && contestStatus(*contest) == model.ContestRunning
		resp.Contest.Frozen = frozen
		isCF := contest.Type == model.ContestCF
		cfDur := int(contestEndTime(*contest).Sub(contest.StartTime).Minutes()) // CF 动态分用的赛长（分钟）

		type scoreRow struct {
			item ContestRankItem
			last time.Time
		}
		rows := make([]scoreRow, 0, len(contestUsers))
		for _, user := range contestUsers {
			u := userMap[user.UserID]
			item := ContestRankItem{
				Problems: make([]ContestProblemStatus, 0, len(contestProblems)),
			}
			item.User.ID = u.UID // 对外用户号
			item.User.Name = u.Username
			item.User.Avatar = u.Avatar
			item.User.Rating = u.Rating

			var last time.Time
			for _, problem := range contestProblems {
				cell := ContestProblemStatus{Label: problem.Label}
				if !frozen {
					score := 0
					if ucp, ok := recordMap[user.UserID][problem.ProblemID]; ok {
						cell.Status = ucp.Status
						if isCF {
							accepted := ucp.Status == judge.Accepted && ucp.AcTime != nil
							t := 0
							if ucp.AcTime != nil {
								t = int(ucp.AcTime.Sub(contest.StartTime).Minutes())
							}
							score = scoring.CFScore(problem.Score, t, cfDur, ucp.UnAcCount, accepted)
							if accepted {
								// CF：分数下面展示 AC 时间（相对开赛 HH:MM）
								diff := ucp.AcTime.Sub(contest.StartTime)
								ts := fmt.Sprintf("%02d:%02d", int(diff.Hours()), int(diff.Minutes())%60)
								cell.AcTime = &ts
								if ucp.AcTime.After(last) {
									last = *ucp.AcTime
								}
							}
						} else {
							score = ucp.Score
							if ucp.AcTime != nil && ucp.AcTime.After(last) {
								last = *ucp.AcTime
							}
						}
					}
					cell.Score = &score
					item.TotalScore += score
				}
				item.Problems = append(item.Problems, cell)
			}
			rows = append(rows, scoreRow{item: item, last: last})
		}

		if !frozen {
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].item.TotalScore != rows[j].item.TotalScore {
					return rows[i].item.TotalScore > rows[j].item.TotalScore
				}
				return rows[i].last.Before(rows[j].last) // 同分早达者靠前
			})
		}
		for i := range rows {
			resp.RankList = append(resp.RankList, rows[i].item)
		}
	}

	for i := 0; i < len(resp.RankList); i++ {
		resp.RankList[i].Rank = i + 1
	}

	// rated 且已结算：给每行附上本场 rating 变化（±delta / 新分）
	if contest.Rated && contest.Settled {
		if changes, err := s.repo.GetRatingChangesByContest(ctx, contestID); err == nil {
			byUser := make(map[int64]model.RatingChange, len(changes))
			for _, c := range changes {
				byUser[c.UserID] = c
			}
			for i := range resp.RankList {
				if rc, ok := byUser[resp.RankList[i].User.ID]; ok {
					d, nr := rc.Delta, rc.NewRating
					resp.RankList[i].RatingDelta = &d
					resp.RankList[i].NewRating = &nr
				}
			}
		}
	}

	// 写缓存（短 TTL，存的是不含 isSelf 的基准榜）；失败不影响返回
	_ = s.cache.SetJSON(ctx, cacheKey, resp, contestRankTTL)

	markSelf(resp, viewerUID)
	return resp, nil
}
