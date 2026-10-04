package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

const submissionTimeLayout = "2006-01-02 15:04:05"

func toSubmitCodeParams(req dto.SubmitCodeReq) service.SubmitCodeParams {
	return service.SubmitCodeParams{
		ProblemID: req.ProblemID,
		Code:      req.Code,
		Language:  req.Language,
	}
}

func toSubmissionCaseResults(cases []service.SubmissionCaseResult) []dto.SubmissionCaseResult {
	items := make([]dto.SubmissionCaseResult, 0, len(cases))
	for _, c := range cases {
		items = append(items, dto.SubmissionCaseResult{ID: c.ID, Status: c.Status, TimeUsed: c.TimeUsed, MemoryUsed: c.MemoryUsed})
	}
	return items
}

func toSubmissionItem(item service.SubmissionItem) dto.SubmissionItemResp {
	return dto.SubmissionItemResp{
		ID:          item.ID,
		ProblemID:   item.ProblemID,
		ProblemName: item.ProblemName,
		UserID:      item.UserID,
		UserName:    item.UserName,
		Rating:      item.Rating,
		Language:    item.Language,
		Result:      item.Result,
		TimeUsed:    item.TimeUsed,
		MemoryUsed:  item.MemoryUsed,
		CreatedAt:   item.CreatedAt.Format(submissionTimeLayout),
	}
}

func toSubmissionListResp(r *service.SubmissionList) *dto.SubmitListResp {
	items := make([]dto.SubmissionItemResp, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, toSubmissionItem(it))
	}
	return &dto.SubmitListResp{Total: r.Total, List: items}
}

func toSubmissionDetailResp(r *service.SubmissionDetail) *dto.SubmissionResp {
	return &dto.SubmissionResp{
		User: dto.UserDetail{
			ID:     r.User.ID,
			Name:   r.User.Name,
			Avatar: r.User.Avatar,
		},
		Problem: dto.SubmissionProblem{
			ID:          r.Problem.ID,
			Name:        r.Problem.Name,
			Description: r.Problem.Description,
			OJ:          r.Problem.OJ,
		},
		Submission: dto.SubmissionInfo{
			ID:            r.Submission.ID,
			Language:      r.Submission.Language,
			Code:          r.Submission.Code,
			CanViewCode:   r.Submission.CanViewCode,
			Status:        r.Submission.Status,
			TimeUsed:      r.Submission.TimeUsed,
			MemoryUsed:    r.Submission.MemoryUsed,
			CompileOutput: r.Submission.CompileOutput,
			CreatedAt:     r.Submission.CreatedAt.Format(submissionTimeLayout),
		},
		CaseResults: toSubmissionCaseResults(r.CaseResults),
	}
}

func toDailyAcceptedCountResp(r *service.DailyAcceptedCount) *dto.DailyAcceptedCountResp {
	return &dto.DailyAcceptedCountResp{Date: r.Date, Count: r.Count}
}
