package handler

import (
	"zoj/internal/http/dto"
	"zoj/internal/service"
)

const problemSetTimeLayout = "2006-01-02 15:04:05"

func toProblemSetListParams(req dto.ProblemSetListReq) service.ProblemSetListParams {
	return service.ProblemSetListParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		Tags:       req.Tags,
		Visibility: req.Visibility,
	}
}

func toProblemSetListResp(r *service.ProblemSetList) *dto.ProblemSetListResp {
	items := make([]dto.ProblemSetItemResp, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.ProblemSetItemResp{
			ID:           it.ID,
			Title:        it.Title,
			Tags:         toTagItems(it.Tags),
			ProblemCount: it.ProblemCount,
			SolvedCount:  it.SolvedCount,
			Visibility:   it.Visibility,
			Published:    it.Published,
			Author:       it.Author,
			UpdatedAt:    it.UpdatedAt.Format(problemSetTimeLayout),
		})
	}
	return &dto.ProblemSetListResp{Total: int(r.Total), List: items}
}

func toProblemSetDetailResp(r *service.ProblemSetDetail) *dto.ProblemSetDetailResp {
	problems := make([]dto.ProblemSetProblemResp, 0, len(r.Problems))
	for _, p := range r.Problems {
		problems = append(problems, dto.ProblemSetProblemResp{
			ID:            p.ID,
			Name:          p.Name,
			Status:        p.Status,
			Difficulty:    p.Difficulty,
			Tags:          toTagItems(p.Tags),
			AcceptedCount: p.AcceptedCount,
			SubmitCount:   p.SubmitCount,
		})
	}
	return &dto.ProblemSetDetailResp{
		ID:           r.ID,
		Title:        r.Title,
		Description:  r.Description,
		Tags:         toTagItems(r.Tags),
		Visibility:   r.Visibility,
		Published:    r.Published,
		ProblemCount: r.ProblemCount,
		SolvedCount:  r.SolvedCount,
		Locked:       r.Locked,
		Author:       r.Author,
		UpdatedAt:    r.UpdatedAt.Format(problemSetTimeLayout),
		Problems:     problems,
	}
}

func toSaveProblemSetParams(req dto.SaveProblemSetReq) service.SaveProblemSetParams {
	return service.SaveProblemSetParams{
		Title:       req.Title,
		Description: req.Description,
		Published:   req.Published,
		Visibility:  req.Visibility,
		InviteCode:  req.InviteCode,
		TagIDs:      req.TagIDs,
		Problems:    req.Problems,
	}
}
