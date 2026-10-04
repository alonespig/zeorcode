package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

func toAdminUserListResp(r *service.AdminUserList) *dto.AdminUserListResp {
	items := make([]dto.AdminUserItem, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.AdminUserItem{
			ID:        it.ID,
			Username:  it.Username,
			StudentNo: it.StudentNo,
			RealName:  it.RealName,
			Email:     it.Email,
			Role:      it.Role,
			Status:    it.Status,
			Signature: it.Signature,
			CreatedAt: it.CreatedAt.Unix(),
		})
	}
	return &dto.AdminUserListResp{Total: r.Total, List: items}
}

func toBatchCreateUsersResult(r *service.BatchCreateUsersResult) *dto.BatchCreateUsersResp {
	return &dto.BatchCreateUsersResp{Created: r.Created}
}
