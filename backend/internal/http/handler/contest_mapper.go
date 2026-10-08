package handler

import (
	"zoj/internal/http/dto"
	"zoj/internal/service"
)

func toContestProblemInputs(inputs []dto.ContestProblemInput) []service.ContestProblemInput {
	items := make([]service.ContestProblemInput, 0, len(inputs))
	for _, in := range inputs {
		items = append(items, service.ContestProblemInput{ProblemID: in.ProblemID, Color: in.Color, Score: in.Score})
	}
	return items
}

func toCreateContestParams(req dto.CreateContestReq) service.CreateContestParams {
	return service.CreateContestParams{
		Name:        req.Name,
		ContestDate: req.ContestDate,
		ContestTime: req.ContestTime,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Duration:    req.Duration,
		Type:        req.Type,
		Problems:    toContestProblemInputs(req.Problems),
		InviteCode:  req.InviteCode,
		Rated:       req.Rated,
	}
}

func toUpdateContestParams(req dto.UpdateContestReq) service.UpdateContestParams {
	return service.UpdateContestParams{
		Name:        req.Name,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Type:        req.Type,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Problems:    toContestProblemInputs(req.Problems),
		Rated:       req.Rated,
	}
}

func toContestSubmitParams(req dto.ContestSubmitReq) service.ContestSubmitParams {
	return service.ContestSubmitParams{
		ContestID: req.ContestID,
		Label:     req.Label,
		Code:      req.Code,
		Language:  req.Language,
	}
}

func toContestSubmissionQueryParams(q dto.ContestSubmissionQuery) service.ContestSubmissionQueryParams {
	return service.ContestSubmissionQueryParams{
		Username:  q.Username,
		ProblemID: q.ProblemID,
		Status:    q.Status,
	}
}

func toContestListResp(r *service.ContestList) *dto.ContestListResp {
	items := make([]dto.ContestItem, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.ContestItem{
			ID:             it.ID,
			Name:           it.Name,
			Description:    it.Description,
			CoverURL:       it.CoverURL,
			StartTime:      it.StartTime,
			EndTime:        it.EndTime,
			Type:           it.Type,
			Status:         it.Status,
			IsRegistered:   it.IsRegistered,
			Rated:          it.Rated,
			NeedInviteCode: it.NeedInviteCode,
			Duration:       it.Duration,
			Participants:   it.Participants,
			Archived:       it.Archived,
		})
	}
	return &dto.ContestListResp{Total: r.Total, List: items}
}

func toContestDescResp(r *service.ContestDesc) *dto.ContestDesc {
	return &dto.ContestDesc{
		Name:        r.Name,
		StartTime:   r.StartTime,
		EndTime:     r.EndTime,
		Description: r.Description,
	}
}

func toContestDetailResp(r *service.ContestDetail) *dto.ContestDetailResp {
	return &dto.ContestDetailResp{
		Name:           r.Name,
		StartTime:      r.StartTime.UnixMilli(),
		EndTime:        r.EndTime.UnixMilli(),
		IsRegistered:   r.IsRegistered,
		NeedInviteCode: r.NeedInviteCode,
		Rule:           r.Rule,
		Rated:          r.Rated,
		Settled:        r.Settled,
	}
}

func toContestProblemListResp(r *service.ContestProblemList) *dto.ContestProblemListResp {
	items := make([]dto.ContestProblem, 0, len(r.ProblemList))
	for _, p := range r.ProblemList {
		items = append(items, dto.ContestProblem{
			ProblemID:     p.ProblemID,
			Name:          p.Name,
			Label:         p.Label,
			Color:         p.Color,
			Status:        p.Status,
			TimeLimit:     p.TimeLimit,
			MemoryLimit:   p.MemoryLimit,
			TotalCount:    p.TotalCount,
			AcceptedCount: p.AcceptedCount,
		})
	}
	return &dto.ContestProblemListResp{ProblemList: items}
}

func toContestProblemDetailResp(r *service.ContestProblemDetail) *dto.ContestProblemResp {
	return &dto.ContestProblemResp{
		Name:         r.Name,
		TimeLimit:    r.TimeLimit,
		MemoryLimit:  r.MemoryLimit,
		Description:  r.Description,
		InputFormat:  r.InputFormat,
		OutputFormat: r.OutputFormat,
		Hint:         r.Hint,
		Samples:      toProblemSamples(r.Samples),
		HasTestData:  r.HasTestData,
	}
}

func toContestSubmitInfoResp(r *service.ContestSubmitInfo) *dto.ContestSubmitResp {
	items := make([]dto.ContestProblemItem, 0, len(r.ProblemList))
	for _, p := range r.ProblemList {
		items = append(items, dto.ContestProblemItem{Name: p.Name, Label: p.Label})
	}
	return &dto.ContestSubmitResp{ProblemList: items}
}

func toContestSubmissionListResp(r *service.ContestSubmissionList) *dto.ContestSubmissionListResp {
	items := make([]dto.ContestSubmissionItem, 0, len(r.Submissions))
	for _, s := range r.Submissions {
		items = append(items, dto.ContestSubmissionItem{
			ID:          s.ID,
			ProblemID:   s.ProblemID,
			ProblemName: s.ProblemName,
			UserID:      s.UserID,
			UserName:    s.UserName,
			Rating:      s.Rating,
			Language:    s.Language,
			Result:      s.Result,
			TimeUsed:    s.TimeUsed,
			MemoryUsed:  s.MemoryUsed,
			CreatedAt:   s.CreatedAt,
		})
	}
	return &dto.ContestSubmissionListResp{Total: r.Total, Submissions: items}
}

func toEditContestResp(r *service.ContestEditInfo) *dto.EditContestResp {
	items := make([]dto.EasyInfo, 0, len(r.ProblemList))
	for _, p := range r.ProblemList {
		items = append(items, dto.EasyInfo{ID: p.ID, Name: p.Name, Color: p.Color, Score: p.Score})
	}
	return &dto.EditContestResp{
		Name:        r.Name,
		Description: r.Description,
		CoverURL:    r.CoverURL,
		Type:        r.Type,
		StartTime:   r.StartTime.UnixMilli(),
		EndTime:     r.EndTime.UnixMilli(),
		ProblemList: items,
		Rated:       r.Rated,
	}
}

func toCreateContestResult(r *service.CreateContestResult) *dto.CreateContestResp {
	return &dto.CreateContestResp{ID: r.ID}
}
