package handler

import (
	"strconv"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type RemoteController struct {
	remoteSrv *service.RemoteService
}

func NewRemoteController(srv *service.RemoteService) *RemoteController {
	return &RemoteController{remoteSrv: srv}
}

// ListOJ GET /api/remote/oj  支持的远程 OJ 列表
func (c *RemoteController) ListOJ(ctx *gin.Context) (any, error) {
	return gin.H{"list": c.remoteSrv.ListOJ(ctx.Request.Context())}, nil
}

// GetRemoteProblem GET /api/remote/problem?oj=HDU&pid=1000
func (c *RemoteController) GetRemoteProblem(ctx *gin.Context) (any, error) {
	oj := ctx.Query("oj")
	pid := ctx.Query("pid")
	if oj == "" || pid == "" {
		return nil, errcode.ErrInvalidParams.WithMsg("缺少 oj 或 pid 参数")
	}
	return c.remoteSrv.CrawlProblem(ctx.Request.Context(), oj, pid)
}

// ListAccounts GET /api/admin/remote-account
func (c *RemoteController) ListAccounts(ctx *gin.Context) (any, error) {
	list, err := c.remoteSrv.ListAccounts(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	return gin.H{"list": list}, nil
}

// CreateAccount POST /api/admin/remote-account
func (c *RemoteController) CreateAccount(ctx *gin.Context) (any, error) {
	var req dto.CreateRemoteAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	id, err := c.remoteSrv.CreateAccount(ctx.Request.Context(), &req)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": id}, nil
}

// UpdateAccount PUT /api/admin/remote-account/:id
func (c *RemoteController) UpdateAccount(ctx *gin.Context) (any, error) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法 id")
	}
	var req dto.UpdateRemoteAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return nil, c.remoteSrv.UpdateAccount(ctx.Request.Context(), id, &req)
}

// DeleteAccount DELETE /api/admin/remote-account/:id
func (c *RemoteController) DeleteAccount(ctx *gin.Context) (any, error) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("非法 id")
	}
	return nil, c.remoteSrv.DeleteAccount(ctx.Request.Context(), id)
}
