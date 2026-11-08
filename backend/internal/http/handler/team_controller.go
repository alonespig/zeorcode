package handler

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/alonespig/zeorcode/backend/internal/http/dto"
	"github.com/alonespig/zeorcode/backend/internal/http/response"
	"github.com/alonespig/zeorcode/backend/internal/service"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

const teamTimeLayout = "2006-01-02 15:04:05"

// TeamController 团队与成员接口
type TeamController struct {
	teamSrv     *service.TeamService
	permissions *service.PermissionService
}

func NewTeamController(teamSrv *service.TeamService, permissions *service.PermissionService) *TeamController {
	return &TeamController{teamSrv: teamSrv, permissions: permissions}
}

// List GET /api/team 团队列表（mine=1 只看我加入的，visibility=0/1 筛选公开度）
func (t *TeamController) List(c *gin.Context) (*dto.TeamListResp, error) {
	var req dto.TeamListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	resp, err := t.teamSrv.List(c.Request.Context(), service.TeamListParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		Mine:       req.Mine,
		Visibility: req.Visibility,
	}, optionalCurrentUserID(c))
	if err != nil {
		return nil, err
	}
	return toTeamListResp(resp), nil
}

// Detail GET /api/team/:id
func (t *TeamController) Detail(c *gin.Context) (*dto.TeamDetailResp, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	resp, err := t.teamSrv.Detail(c.Request.Context(), id, optionalCurrentUserID(c), isSiteTeamManager(c, t.permissions))
	if err != nil {
		return nil, err
	}
	return toTeamDetailResp(resp), nil
}

// Create POST /api/team
func (t *TeamController) Create(c *gin.Context) (gin.H, error) {
	var req dto.SaveTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	id, err := t.teamSrv.Create(c.Request.Context(), toSaveTeamParams(req), userID)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": id}, nil
}

// Update PUT /api/team/:id
func (t *TeamController) Update(c *gin.Context) (gin.H, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	var req dto.SaveTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := t.teamSrv.Update(c.Request.Context(), id, toSaveTeamParams(req), userID, isSiteTeamManager(c, t.permissions)); err != nil {
		return nil, err
	}
	return nil, nil
}

// Delete DELETE /api/team/:id 解散团队
func (t *TeamController) Delete(c *gin.Context) (gin.H, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := t.teamSrv.Delete(c.Request.Context(), id, userID, isSiteTeamManager(c, t.permissions)); err != nil {
		return nil, err
	}
	return nil, nil
}

// Join POST /api/team/:id/join
func (t *TeamController) Join(c *gin.Context) (gin.H, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	var req dto.JoinTeamReq
	// 公开团队不需要邀请码，请求体可以为空，绑定失败不算错
	_ = c.ShouldBindJSON(&req)
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := t.teamSrv.Join(c.Request.Context(), id, userID, req.InviteCode); err != nil {
		return nil, err
	}
	return nil, nil
}

// Quit POST /api/team/:id/quit
func (t *TeamController) Quit(c *gin.Context) (gin.H, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := t.teamSrv.Quit(c.Request.Context(), id, userID); err != nil {
		return nil, err
	}
	return nil, nil
}

// ListMembers GET /api/team/:id/member
func (t *TeamController) ListMembers(c *gin.Context) (*dto.TeamMemberListResp, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	resp, err := t.teamSrv.ListMembers(c.Request.Context(), id, optionalCurrentUserID(c), isSiteTeamManager(c, t.permissions))
	if err != nil {
		return nil, err
	}
	return toTeamMemberListResp(resp), nil
}

// DownloadStudentImportTemplate GET /api/team/:id/member/import/template
func (t *TeamController) DownloadStudentImportTemplate(c *gin.Context) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	data, err := t.teamSrv.StudentImportTemplate(c.Request.Context(), id, actorID, isSiteTeamManager(c, t.permissions))
	if err != nil {
		response.Fail(c, err)
		return
	}
	filename := url.PathEscape("学生名单导入模板.xlsx")
	c.Header("Content-Disposition", "attachment; filename=student-import-template.xlsx; filename*=UTF-8''"+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ImportStudents POST /api/team/:id/member/import，multipart 字段名为 file。
func (t *TeamController) ImportStudents(c *gin.Context) (*dto.TeamStudentImportResp, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}

	file, err := openExcelImport(c)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	resp, err := t.teamSrv.ImportStudents(c.Request.Context(), id, actorID, isSiteTeamManager(c, t.permissions), file)
	if err != nil {
		return nil, err
	}
	return toStudentImportResp(resp), nil
}

// ImportStudentsManual POST /api/team/:id/member/import/manual，手动粘贴学生名单。
func (t *TeamController) ImportStudentsManual(c *gin.Context) (*dto.TeamStudentImportResp, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	var req dto.TeamStudentManualImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	resp, err := t.teamSrv.ImportStudentsManual(c.Request.Context(), id, actorID, isSiteTeamManager(c, t.permissions), req.Text)
	if err != nil {
		return nil, err
	}
	return toStudentImportResp(resp), nil
}

// SetMemberRole PUT /api/team/:id/member/:uid/role 设置/取消团队管理员（:uid 为对外用户号）
func (t *TeamController) SetMemberRole(c *gin.Context) (gin.H, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	uid, err := parseUIDParam(c)
	if err != nil {
		return nil, err
	}
	var req dto.SetMemberRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := t.teamSrv.SetMemberRole(c.Request.Context(), id, uid, actorID, isSiteTeamManager(c, t.permissions), req.Role); err != nil {
		return nil, err
	}
	return nil, nil
}

// RemoveMember DELETE /api/team/:id/member/:uid（:uid 为对外用户号）
func (t *TeamController) RemoveMember(c *gin.Context) (gin.H, error) {
	id, err := t.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	uid, err := parseUIDParam(c)
	if err != nil {
		return nil, err
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := t.teamSrv.RemoveMember(c.Request.Context(), id, uid, actorID, isSiteTeamManager(c, t.permissions)); err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *TeamController) resolveTeamID(c *gin.Context) (int64, error) {
	publicID, err := parsePublicIDParam(c, "id", "团队")
	if err != nil {
		return 0, err
	}
	return t.teamSrv.ResolveID(c.Request.Context(), publicID)
}

func parseUIDParam(c *gin.Context) (int64, error) {
	uid, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil || uid <= 0 {
		return 0, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	return uid, nil
}

func toSaveTeamParams(req dto.SaveTeamReq) service.SaveTeamParams {
	return service.SaveTeamParams{
		Name:        req.Name,
		CoverURL:    req.CoverURL,
		Description: req.Description,
		Visibility:  req.Visibility,
		InviteCode:  req.InviteCode,
	}
}

func toTeamListResp(r *service.TeamListResult) *dto.TeamListResp {
	items := make([]dto.TeamItemResp, 0, len(r.List))
	for _, it := range r.List {
		item := dto.TeamItemResp{
			ID:          it.ID,
			Name:        it.Name,
			CoverURL:    it.CoverURL,
			Visibility:  it.Visibility,
			MemberCount: it.MemberCount,
			Owner:       it.Owner,
			OwnerAvatar: it.OwnerAvatar,
			MyRole:      it.MyRole,
			CreatedAt:   it.CreatedAt.Format(teamTimeLayout),
		}
		items = append(items, item)
	}
	return &dto.TeamListResp{Total: int(r.Total), List: items}
}

func toTeamDetailResp(r *service.TeamDetail) *dto.TeamDetailResp {
	return &dto.TeamDetailResp{
		ID:            r.ID,
		Name:          r.Name,
		CoverURL:      r.CoverURL,
		Description:   r.Description,
		Visibility:    r.Visibility,
		MemberCount:   r.MemberCount,
		HomeworkCount: r.HomeworkCount,
		Owner:         r.Owner,
		OwnerUID:      r.OwnerUID,
		MyRole:        r.MyRole,
		CanManage:     r.CanManage,
		InviteCode:    r.InviteCode,
		CreatedAt:     r.CreatedAt.Format(teamTimeLayout),
	}
}

func toTeamMemberListResp(r *service.TeamMemberListResult) *dto.TeamMemberListResp {
	items := make([]dto.TeamMemberItem, 0, len(r.List))
	for _, it := range r.List {
		items = append(items, dto.TeamMemberItem{
			UID:       it.UID,
			Username:  it.Username,
			StudentNo: it.StudentNo,
			RealName:  it.RealName,
			Gender:    it.Gender,
			Avatar:    it.Avatar,
			Role:      it.Role,
			JoinedAt:  it.JoinedAt.Format(teamTimeLayout),
		})
	}
	return &dto.TeamMemberListResp{Total: r.Total, List: items}
}

func toStudentImportResp(r *service.StudentImportResult) *dto.TeamStudentImportResp {
	return &dto.TeamStudentImportResp{
		Total:          r.Total,
		CreatedUsers:   r.CreatedUsers,
		AddedMembers:   r.AddedMembers,
		SkippedMembers: r.SkippedMembers,
	}
}
