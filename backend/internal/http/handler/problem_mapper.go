package handler

import (
	"zoj/internal/http/dto"
	"zoj/internal/service"
)

const problemTimeLayout = "2006-01-02 15:04:05"

func toTagItems(tags []service.TagItem) []dto.TagItem {
	items := make([]dto.TagItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, dto.TagItem{ID: t.ID, Name: t.Name})
	}
	return items
}

func toTagItem(t *service.TagItem) *dto.TagItem {
	if t == nil {
		return nil
	}
	return &dto.TagItem{ID: t.ID, Name: t.Name}
}

func toAdminTagItems(tags []service.AdminTagItem) []dto.AdminTagItem {
	items := make([]dto.AdminTagItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, dto.AdminTagItem{ID: t.ID, Name: t.Name, ProblemCount: t.ProblemCount, ProblemSetCount: t.ProblemSetCount})
	}
	return items
}

func toProblemSamples(samples []service.ProblemSample) []dto.ProblemSample {
	items := make([]dto.ProblemSample, 0, len(samples))
	for _, s := range samples {
		items = append(items, dto.ProblemSample{Input: s.Input, Output: s.Output, Explain: s.Explain})
	}
	return items
}

func toServiceProblemSamples(samples []dto.ProblemSample) []service.ProblemSample {
	items := make([]service.ProblemSample, 0, len(samples))
	for _, s := range samples {
		items = append(items, service.ProblemSample{Input: s.Input, Output: s.Output, Explain: s.Explain})
	}
	return items
}

func toProblemDetailResp(r *service.ProblemDetail) *dto.ProblemDetailResp {
	return &dto.ProblemDetailResp{
		ID:              r.ID,
		Name:            r.Name,
		Difficulty:      r.Difficulty,
		TimeLimit:       r.TimeLimit,
		MemoryLimit:     r.MemoryLimit,
		Description:     r.Description,
		InputFormat:     r.InputFormat,
		OutputFormat:    r.OutputFormat,
		SubmitCount:     r.SubmitCount,
		AcceptedCount:   r.AcceptedCount,
		Hint:            r.Hint,
		Samples:         toProblemSamples(r.Samples),
		Tags:            toTagItems(r.Tags),
		OJ:              r.OJ,
		RemoteProblemID: r.RemoteProblemID,
		Hidden:          r.Hidden,
		HasTestData:     r.HasTestData,
	}
}

func toProblemItem(item service.ProblemItem) dto.ProblemItemResp {
	return dto.ProblemItemResp{
		ID:            item.ID,
		Name:          item.Name,
		Status:        item.Status,
		Difficulty:    item.Difficulty,
		Tags:          toTagItems(item.Tags),
		AcceptedCount: item.AcceptedCount,
		SubmitCount:   item.SubmitCount,
		CreatedAt:     item.CreatedAt.Format(problemTimeLayout),
		Hidden:        item.Hidden,
		OJ:            item.OJ,
	}
}

func toProblemListResp(r *service.ProblemList) *dto.ProblemListResp {
	items := make([]dto.ProblemItemResp, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, toProblemItem(it))
	}
	return &dto.ProblemListResp{Total: int(r.Total), List: items}
}

func toProblemForEdit(r *service.ProblemForEdit) *dto.ProblemModel {
	return &dto.ProblemModel{
		ID:           r.ID,
		Name:         r.Name,
		Difficulty:   r.Difficulty,
		TimeLimit:    r.TimeLimit,
		MemoryLimit:  r.MemoryLimit,
		Description:  r.Description,
		InputFormat:  r.InputFormat,
		OutputFormat: r.OutputFormat,
		Hint:         r.Hint,
		Samples:      toProblemSamples(r.Samples),
		Tags:         toTagItems(r.Tags),
		Hidden:       r.Hidden,
	}
}

func toCreateProblemParams(req dto.CreateProblemReq) service.CreateProblemParams {
	return service.CreateProblemParams{
		DisplayID:       req.DisplayID,
		Name:            req.Name,
		Difficulty:      req.Difficulty,
		TimeLimit:       req.TimeLimit,
		MemoryLimit:     req.MemoryLimit,
		Description:     req.Description,
		InputFormat:     req.InputFormat,
		OutputFormat:    req.OutputFormat,
		Hint:            req.Hint,
		Samples:         toServiceProblemSamples(req.Samples),
		TagsID:          req.TagsID,
		OJ:              req.OJ,
		RemoteProblemID: req.RemoteProblemID,
		Hidden:          req.Hidden,
	}
}

func toUpdateProblemParams(req dto.UpdateProblemReq) service.UpdateProblemParams {
	return service.UpdateProblemParams{
		ID:           req.ID,
		DisplayID:    req.DisplayID,
		Name:         req.Name,
		Difficulty:   req.Difficulty,
		TimeLimit:    req.TimeLimit,
		MemoryLimit:  req.MemoryLimit,
		Description:  req.Description,
		InputFormat:  req.InputFormat,
		OutputFormat: req.OutputFormat,
		Hint:         req.Hint,
		Samples:      toServiceProblemSamples(req.Samples),
		Tags:         req.Tags,
		Hidden:       req.Hidden,
	}
}

func toProblemListParams(req dto.ProblemListReq) service.ProblemListParams {
	return service.ProblemListParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		Order:      req.Order,
		Difficulty: req.Difficulty,
		Tags:       req.Tags,
	}
}
