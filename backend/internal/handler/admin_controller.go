package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"zoj/internal/common/errcode"
	"zoj/internal/common/response"
	"zoj/internal/dto"
	"zoj/internal/service"
	"zoj/pkg/judge"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const maxAdminUserImportFileSize = 5 << 20

// AdminController 后台管理接口
type AdminController struct {
	userSrv *service.UserService
}

func NewAdminController(userSrv *service.UserService) *AdminController {
	return &AdminController{userSrv: userSrv}
}

// ListUsers GET /api/admin/users?page=&pageSize=
func (a *AdminController) ListUsers(c *gin.Context) (any, error) {
	var req dto.PageForm
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return a.userSrv.ListAllUsers(c.Request.Context(), req.Page, req.PageSize)
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
func (a *AdminController) ImportUsers(c *gin.Context) (any, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAdminUserImportFileSize+(1<<20))
	fh, err := c.FormFile("file")
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("请选择要导入的 Excel 文件").Wrap(err)
	}
	if fh.Size <= 0 || fh.Size > maxAdminUserImportFileSize {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 文件大小不能超过 5MB")
	}
	if strings.ToLower(filepath.Ext(fh.Filename)) != ".xlsx" {
		return nil, errcode.ErrInvalidParams.WithMsg("仅支持 .xlsx 格式的 Excel 文件")
	}
	file, err := fh.Open()
	if err != nil {
		return nil, errcode.ErrFileUpload.WithMsg("无法读取 Excel 文件").Wrap(err)
	}
	defer file.Close()

	return a.userSrv.ImportUsers(c.Request.Context(), file)
}

// JudgeStatus GET /api/admin/judge/status 评测机健康面板：现场探活各实例
func (a *AdminController) JudgeStatus(c *gin.Context) (any, error) {
	urls := viper.GetStringSlice("judge.urls")
	if len(urls) == 0 {
		if u := viper.GetString("judge.url"); u != "" {
			urls = []string{u}
		}
	}
	return gin.H{"list": judge.ProbeAll(urls)}, nil
}

// SetUserStatus PUT /api/admin/users/:id/status 封禁/解封用户（:id 为对外用户号）
func (a *AdminController) SetUserStatus(c *gin.Context) (any, error) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法用户号")
	}
	var req dto.SetUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if err := a.userSrv.SetUserStatus(c.Request.Context(), uid, req.Status); err != nil {
		return nil, err
	}
	return nil, nil
}

// SetUserRole PUT /api/admin/users/:id/role 修改用户角色（:id 为对外用户号）
func (a *AdminController) SetUserRole(c *gin.Context) (any, error) {
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
