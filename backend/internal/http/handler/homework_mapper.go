package handler

import (
	"zoj/internal/http/dto"
	"zoj/internal/service"
)

const homeworkTimeLayout = "2006-01-02 15:04:05"

func toHomeworkListResp(r *service.HomeworkList) *dto.HomeworkListResp {
	items := make([]dto.HomeworkItemResp, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.HomeworkItemResp{
			ID:           it.ID,
			Title:        it.Title,
			Status:       it.Status,
			ProblemCount: it.ProblemCount,
			SolvedCount:  it.SolvedCount,
			StartTime:    it.StartTime.Format(homeworkTimeLayout),
			EndTime:      it.EndTime.Format(homeworkTimeLayout),
			Creator:      it.Creator,
			CanEdit:      it.CanEdit,
			CanDelete:    it.CanDelete,
		})
	}
	return &dto.HomeworkListResp{Total: int(r.Total), List: items}
}

func toHomeworkDetailResp(r *service.HomeworkDetail) *dto.HomeworkDetailResp {
	problems := make([]dto.HomeworkProblemResp, 0, len(r.Problems))
	for _, p := range r.Problems {
		problems = append(problems, dto.HomeworkProblemResp{
			ID:            p.ID,
			Name:          p.Name,
			Status:        p.Status,
			Difficulty:    p.Difficulty,
			Tags:          toTagItems(p.Tags),
			MyScore:       p.MyScore,
			AcceptedCount: p.AcceptedCount,
			SubmitCount:   p.SubmitCount,
		})
	}
	return &dto.HomeworkDetailResp{
		ID:            r.ID,
		TeamID:        r.TeamID,
		TeamName:      r.TeamName,
		Title:         r.Title,
		Description:   r.Description,
		Status:        r.Status,
		StartTime:     r.StartTime.Format(homeworkTimeLayout),
		EndTime:       r.EndTime.Format(homeworkTimeLayout),
		Creator:       r.Creator,
		CreatorAvatar: r.CreatorAvatar,
		ProblemCount:  r.ProblemCount,
		SolvedCount:   r.SolvedCount,
		MyScore:       r.MyScore,
		TotalScore:    r.TotalScore,
		Locked:        r.Locked,
		CanEdit:       r.CanEdit,
		CanSeeAll:     r.CanSeeAll,
		Problems:      problems,
	}
}

func toHomeworkRankResp(r *service.HomeworkRank) *dto.HomeworkRankResp {
	rows := make([]dto.HomeworkRankRow, 0, len(r.List))
	for _, row := range r.List {
		cells := make([]dto.RankCell, 0, len(row.Cells))
		for _, c := range row.Cells {
			cells = append(cells, dto.RankCell{ProblemID: c.ProblemID, Score: c.Score, Solved: c.Solved})
		}
		rows = append(rows, dto.HomeworkRankRow{
			Rank:        row.Rank,
			UID:         row.UID,
			Username:    row.Username,
			StudentNo:   row.StudentNo,
			RealName:    row.RealName,
			Avatar:      row.Avatar,
			TotalScore:  row.TotalScore,
			SolvedCount: row.SolvedCount,
			Cells:       cells,
		})
	}
	return &dto.HomeworkRankResp{
		StartTime:  r.StartTime.Format(homeworkTimeLayout),
		EndTime:    r.EndTime.Format(homeworkTimeLayout),
		ProblemIDs: r.ProblemIDs,
		TotalScore: r.TotalScore,
		List:       rows,
	}
}

func toHomeworkSubmissionListResp(r *service.HomeworkSubmissionList) *dto.HomeworkSubmissionListResp {
	items := make([]dto.HomeworkSubmissionItem, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.HomeworkSubmissionItem{
			ID:          it.ID,
			UID:         it.UID,
			Username:    it.Username,
			StudentNo:   it.StudentNo,
			RealName:    it.RealName,
			Avatar:      it.Avatar,
			ProblemID:   it.ProblemID,
			ProblemName: it.ProblemName,
			Status:      it.Status,
			Score:       it.Score,
			Language:    it.Language,
			TimeUsed:    it.TimeUsed,
			MemoryUsed:  it.MemoryUsed,
			InWindow:    it.InWindow,
			CreatedAt:   it.CreatedAt.Format(homeworkTimeLayout),
		})
	}
	return &dto.HomeworkSubmissionListResp{Total: int(r.Total), CanSeeAll: r.CanSeeAll, List: items}
}

func toHomeworkSubmissionDetailResp(r *service.HomeworkSubmissionDetail) *dto.HomeworkSubmissionDetailResp {
	return &dto.HomeworkSubmissionDetailResp{
		ID:            r.ID,
		UID:           r.UID,
		Username:      r.Username,
		StudentNo:     r.StudentNo,
		RealName:      r.RealName,
		Avatar:        r.Avatar,
		ProblemID:     r.ProblemID,
		ProblemName:   r.ProblemName,
		OJ:            r.OJ,
		Status:        r.Status,
		Score:         r.Score,
		Language:      r.Language,
		Code:          r.Code,
		CompileOutput: r.CompileOutput,
		TimeUsed:      r.TimeUsed,
		MemoryUsed:    r.MemoryUsed,
		CaseResults:   toSubmissionCaseResults(r.CaseResults),
		InWindow:      r.InWindow,
		CreatedAt:     r.CreatedAt.Format(homeworkTimeLayout),
	}
}

func toSaveHomeworkParams(req dto.SaveHomeworkReq) service.SaveHomeworkParams {
	return service.SaveHomeworkParams{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Problems:    req.Problems,
	}
}

func toHomeworkSubmitParams(req dto.HomeworkSubmitReq) service.HomeworkSubmitParams {
	return service.HomeworkSubmitParams{
		ProblemID: req.ProblemID,
		Code:      req.Code,
		Language:  req.Language,
	}
}

func toHomeworkSubmissionQueryParams(q dto.HomeworkSubmissionQuery) service.HomeworkSubmissionQueryParams {
	return service.HomeworkSubmissionQueryParams{
		Page:      q.Page,
		PageSize:  q.PageSize,
		UID:       q.UID,
		ProblemID: q.ProblemID,
		Status:    q.Status,
	}
}
