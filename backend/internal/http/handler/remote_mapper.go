package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

func toRemoteProblemResp(r *service.RemoteProblem) *dto.RemoteProblemResp {
	return &dto.RemoteProblemResp{
		OJ:              r.OJ,
		RemoteProblemID: r.RemoteProblemID,
		Title:           r.Title,
		TimeLimit:       r.TimeLimit,
		MemoryLimit:     r.MemoryLimit,
		Description:     r.Description,
		InputFormat:     r.InputFormat,
		OutputFormat:    r.OutputFormat,
		Samples:         toProblemSamples(r.Samples),
		Hint:            r.Hint,
		Source:          r.Source,
	}
}

func toRemoteAccounts(items []service.RemoteAccount) []dto.RemoteAccountItem {
	result := make([]dto.RemoteAccountItem, 0, len(items))
	for _, it := range items {
		result = append(result, dto.RemoteAccountItem{
			ID:        it.ID,
			OJ:        it.OJ,
			AuthType:  it.AuthType,
			Username:  it.Username,
			HasSecret: it.HasSecret,
			Enabled:   it.Enabled,
			Valid:     it.Valid,
			Busy:      it.Busy,
			CreatedAt: it.CreatedAt,
		})
	}
	return result
}

func toCreateRemoteAccountParams(req dto.CreateRemoteAccountReq) service.CreateRemoteAccountParams {
	return service.CreateRemoteAccountParams{
		OJ:       req.OJ,
		AuthType: req.AuthType,
		Username: req.Username,
		Secret:   req.Secret,
		Enabled:  req.Enabled,
	}
}

func toUpdateRemoteAccountParams(req dto.UpdateRemoteAccountReq) service.UpdateRemoteAccountParams {
	return service.UpdateRemoteAccountParams{
		AuthType: req.AuthType,
		Username: req.Username,
		Secret:   req.Secret,
		Enabled:  req.Enabled,
	}
}
