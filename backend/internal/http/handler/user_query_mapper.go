package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

func toUserRankListResp(r *service.UserRankList) *dto.UserRank {
	items := make([]dto.UserInfo, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.UserInfo{
			ID:          it.ID,
			Index:       it.Index,
			Username:    it.Username,
			Avatar:      it.Avatar,
			PassCount:   it.PassCount,
			SubmitCount: it.SubmitCount,
			Signature:   it.Signature,
			Rating:      it.Rating,
			MaxRating:   it.MaxRating,
		})
	}
	return &dto.UserRank{Total: r.Total, UserRank: items}
}

func toRatingHistoryResp(r *service.RatingHistory) *dto.RatingHistoryResp {
	items := make([]dto.RatingHistoryItem, 0, len(r.History))
	for _, it := range r.History {
		items = append(items, dto.RatingHistoryItem{
			ContestID:   it.ContestID,
			ContestName: it.ContestName,
			Rank:        it.Rank,
			OldRating:   it.OldRating,
			NewRating:   it.NewRating,
			Delta:       it.Delta,
			Time:        it.Time.UnixMilli(),
		})
	}
	return &dto.RatingHistoryResp{Rating: r.Rating, MaxRating: r.MaxRating, History: items}
}

func toContestHistoryResp(r *service.ContestHistory) *dto.ContestHistoryResp {
	items := make([]dto.ContestHistoryItem, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.ContestHistoryItem{
			ContestID:   it.ContestID,
			ContestName: it.ContestName,
			Type:        it.Type,
			Time:        it.Time.Unix(),
			Status:      it.Status,
			Rank:        it.Rank,
			Total:       it.Total,
			OldRating:   it.OldRating,
			NewRating:   it.NewRating,
			Delta:       it.Delta,
		})
	}
	return &dto.ContestHistoryResp{List: items}
}

func toUserProblems(items []service.UserProblem) []dto.ProblemSimple {
	result := make([]dto.ProblemSimple, 0, len(items))
	for _, p := range items {
		result = append(result, dto.ProblemSimple{ID: p.ID, Name: p.Name})
	}
	return result
}

func toUserInfoResp(r *service.UserInfo) *dto.UserInfoResp {
	return &dto.UserInfoResp{
		ID:          r.ID,
		Name:        r.Name,
		Avatar:      r.Avatar,
		Email:       r.Email,
		Gender:      r.Gender,
		School:      r.School,
		Signature:   r.Signature,
		CreatedAt:   r.CreatedAt.Unix(),
		SolveItem:   toUserProblems(r.SolveItem),
		UnsolveItem: toUserProblems(r.UnsolveItem),
	}
}

func toProfileResp(r *service.Profile) *dto.UserProfile {
	return &dto.UserProfile{
		User: dto.UserDetail{
			ID:        r.User.ID,
			Name:      r.User.Name,
			Avatar:    r.User.Avatar,
			Gender:    r.User.Gender,
			School:    r.User.School,
			Signature: r.User.Signature,
			CreatedAt: r.User.CreatedAt.Format("2006-6-6"),
		},
		SolveItem:   toUserProblems(r.SolveItem),
		UnsolveItem: toUserProblems(r.UnsolveItem),
	}
}
