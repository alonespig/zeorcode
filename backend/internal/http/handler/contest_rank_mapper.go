package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

func toContestRankResp(r *service.ContestRank) *dto.ContestRankResp {
	items := make([]dto.ContestRankItem, 0, len(r.RankList))
	for _, it := range r.RankList {
		problems := make([]dto.ContestProblemStatus, 0, len(it.Problems))
		for _, p := range it.Problems {
			problems = append(problems, dto.ContestProblemStatus{
				Label:      p.Label,
				Status:     p.Status,
				AcTime:     p.AcTime,
				Tries:      p.Tries,
				FirstBlood: p.FirstBlood,
				Score:      p.Score,
			})
		}
		items = append(items, dto.ContestRankItem{
			Rank: it.Rank,
			User: dto.User{
				ID:     it.User.ID,
				Name:   it.User.Name,
				School: it.User.School,
				Avatar: it.User.Avatar,
				Rating: it.User.Rating,
			},
			PassCount:   it.PassCount,
			Penalty:     it.Penalty,
			TotalScore:  it.TotalScore,
			Problems:    problems,
			IsSelf:      it.IsSelf,
			RatingDelta: it.RatingDelta,
			NewRating:   it.NewRating,
		})
	}
	problems := make([]dto.ProblemLabel, 0, len(r.Problems))
	for _, p := range r.Problems {
		problems = append(problems, dto.ProblemLabel{Label: p.Label, Color: p.Color, MaxScore: p.MaxScore})
	}
	return &dto.ContestRankResp{
		Contest: dto.Contest{
			ID:     r.Contest.ID,
			Title:  r.Contest.Title,
			Status: r.Contest.Status,
			Rule:   r.Contest.Rule,
			Frozen: r.Contest.Frozen,
		},
		RankList: items,
		Problems: problems,
	}
}

func toMyContestRankResp(r *service.MyContestRank) *dto.MyContestRankResp {
	return &dto.MyContestRankResp{
		Found:      r.Found,
		Rank:       r.Rank,
		PassCount:  r.PassCount,
		Penalty:    r.Penalty,
		TotalScore: r.TotalScore,
		Rule:       r.Rule,
	}
}
