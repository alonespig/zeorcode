package handler

import (
	"net/http"
	"strconv"

	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/http/dto"
	"github.com/alonespig/zeorcode/backend/internal/http/response"
	"github.com/alonespig/zeorcode/backend/internal/service"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/alonespig/zeorcode/backend/pkg/judge"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// AdminController 后台管理接口
type AdminController struct {
	userSrv      *service.UserService
	dashboardSrv *service.DashboardService
	auditSrv     *service.AuditService
	permSrv      *service.PermissionService
}

func NewAdminController(
	userSrv *service.UserService,
	dashboardSrv *service.DashboardService,
	auditSrv *service.AuditService,
	permSrv *service.PermissionService,
) *AdminController {
	return &AdminController{
		userSrv:      userSrv,
		dashboardSrv: dashboardSrv,
		auditSrv:     auditSrv,
		permSrv:      permSrv,
	}
}

// Dashboard GET /api/admin/dashboard 后台工作台聚合数据。
func (a *AdminController) Dashboard(c *gin.Context) (*dto.DashboardOverviewResp, error) {
	result, err := a.dashboardSrv.Overview(
		c.Request.Context(),
		hasAnyPermission(c, a.permSrv, authz.SubmissionRead),
	)
	if err != nil {
		return nil, err
	}
	return toDashboardOverviewResp(result), nil
}

func (a *AdminController) AuditLogs(c *gin.Context) (*dto.AuditLogListResp, error) {
	var req dto.AuditLogListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	result, err := a.auditSrv.List(c.Request.Context(), service.AuditQueryParams{
		Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword,
		Method: req.Method, Success: req.Success,
	})
	if err != nil {
		return nil, err
	}
	items := make([]dto.AuditLogItemResp, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, dto.AuditLogItemResp{
			ID: item.ID, ActorID: item.ActorID, ActorName: item.ActorName,
			Method: item.Method, Path: item.Path, Target: item.Target,
			ClientIP: item.ClientIP, Success: item.Success, Code: item.Code,
			CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &dto.AuditLogListResp{Total: result.Total, List: items}, nil
}

// ListUsers GET /api/admin/users?page=&pageSize=
func (a *AdminController) ListUsers(c *gin.Context) (*dto.AdminUserListResp, error) {
	var req dto.AdminUserListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	resp, err := a.userSrv.ListAllUsers(c.Request.Context(), service.AdminUserListParams{
		Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword, Role: req.Role, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return toAdminUserListResp(resp), nil
}

// DownloadUserImportTemplate GET /api/admin/users/import/template
func (a *AdminController) DownloadUserImportTemplate(c *gin.Context) {
	data, err := a.userSrv.UserImportTemplate()
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=user-import-template.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ImportUsers POST /api/admin/users/import，multipart 字段名为 file。
func (a *AdminController) ImportUsers(c *gin.Context) (*dto.BatchCreateUsersResp, error) {
	file, err := openExcelImport(c)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	resp, err := a.userSrv.ImportUsers(c.Request.Context(), file)
	if err != nil {
		return nil, err
	}
	return toBatchCreateUsersResult(resp), nil
}

// JudgeStatus GET /api/admin/judge/status 评测机健康面板：现场探活各实例
func (a *AdminController) JudgeStatus(c *gin.Context) (gin.H, error) {
	urls := viper.GetStringSlice("judge.urls")
	if len(urls) == 0 {
		if u := viper.GetString("judge.url"); u != "" {
			urls = []string{u}
		}
	}
	return gin.H{"list": judge.ProbeAll(urls)}, nil
}

// SetUserStatus PUT /api/admin/users/:id/status 封禁/解封用户（:id 为对外用户号）
func (a *AdminController) SetUserStatus(c *gin.Context) (gin.H, error) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	var req dto.SetUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	actorID, actorRole, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := a.userSrv.SetUserStatus(c.Request.Context(), actorID, actorRole, uid, req.Status); err != nil {
		return nil, err
	}
	return nil, nil
}

// SetUserRole PUT /api/admin/users/:id/role 修改用户角色（:id 为对外用户号）
func (a *AdminController) SetUserRole(c *gin.Context) (gin.H, error) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	var req dto.SetUserRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := a.userSrv.SetUserRole(c.Request.Context(), actorID, uid, req.Role); err != nil {
		return nil, err
	}
	return nil, nil
}

// PermissionCatalog GET /api/admin/permissions/catalog 返回权限点分组目录。
func (a *AdminController) PermissionCatalog(c *gin.Context) (*dto.PermissionCatalogResp, error) {
	return &dto.PermissionCatalogResp{Groups: authz.Catalog}, nil
}

// GetUserPermissions GET /api/admin/users/:id/permissions 查询某个管理员的权限。
func (a *AdminController) GetUserPermissions(c *gin.Context) (*dto.UserPermissionsResp, error) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	permissions, role, err := a.permSrv.GetByUID(c.Request.Context(), uid)
	if err != nil {
		return nil, err
	}
	return &dto.UserPermissionsResp{Role: role, Permissions: permissions}, nil
}

// SetUserPermissions PUT /api/admin/users/:id/permissions 保存某个管理员的权限。
func (a *AdminController) SetUserPermissions(c *gin.Context) (gin.H, error) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	var req dto.SetUserPermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if err := a.permSrv.SetByUID(c.Request.Context(), uid, req.Permissions); err != nil {
		return nil, err
	}
	return nil, nil
}
