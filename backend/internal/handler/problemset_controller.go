package handler

import (
	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

// ProblemSetController 题单接口（前台只读 + 后台管理）
type ProblemSetController struct {
	setSrv *service.ProblemSetService
}

func NewProblemSetController(setSrv *service.ProblemSetService) *ProblemSetController {
	return &ProblemSetController{setSrv: setSrv}
}

// List GET /api/problemset 前台题单列表（只含已发布）
func (p *ProblemSetController) List(c *gin.Context) (any, error) {
	var req dto.ProblemSetListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return p.setSrv.List(c.Request.Context(), &req, optionalUserIDPtr(c), false)
}

// Detail GET /api/problemset/:id
func (p *ProblemSetController) Detail(c *gin.Context) (any, error) {
	id, err := p.resolveProblemSetID(c)
	if err != nil {
		return nil, err
	}
	return p.setSrv.Detail(c.Request.Context(), id, optionalUserIDPtr(c), isAdminFromCtx(c))
}

// Unlock POST /api/problemset/:id/unlock 提交邀请码解锁
func (p *ProblemSetController) Unlock(c *gin.Context) (any, error) {
	id, err := p.resolveProblemSetID(c)
	if err != nil {
		return nil, err
	}
	var req dto.UnlockProblemSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := p.setSrv.Unlock(c.Request.Context(), id, userID, req.InviteCode); err != nil {
		return nil, err
	}
	return nil, nil
}

// AdminList GET /api/admin/problemset 后台列表（含草稿）
func (p *ProblemSetController) AdminList(c *gin.Context) (any, error) {
	var req dto.ProblemSetListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return p.setSrv.List(c.Request.Context(), &req, optionalUserIDPtr(c), true)
}

// AdminCreate POST /api/admin/problemset
func (p *ProblemSetController) AdminCreate(c *gin.Context) (any, error) {
	var req dto.SaveProblemSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	id, err := p.setSrv.Save(c.Request.Context(), 0, &req, actorID)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": id}, nil
}

// AdminUpdate PUT /api/admin/problemset/:id
func (p *ProblemSetController) AdminUpdate(c *gin.Context) (any, error) {
	id, err := p.resolveProblemSetID(c)
	if err != nil {
		return nil, err
	}
	var req dto.SaveProblemSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	actorID, _, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if _, err := p.setSrv.Save(c.Request.Context(), id, &req, actorID); err != nil {
		return nil, err
	}
	return nil, nil
}

// AdminDelete DELETE /api/admin/problemset/:id
func (p *ProblemSetController) AdminDelete(c *gin.Context) (any, error) {
	id, err := p.resolveProblemSetID(c)
	if err != nil {
		return nil, err
	}
	if err := p.setSrv.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}
	return nil, nil
}

func (p *ProblemSetController) resolveProblemSetID(c *gin.Context) (int64, error) {
	publicID, err := parsePublicIDParam(c, "id", "题单")
	if err != nil {
		return 0, err
	}
	return p.setSrv.ResolveID(c.Request.Context(), publicID)
}

// optionalUserIDPtr 未登录返回 nil，让 service 跳过做题进度查询。
func optionalUserIDPtr(c *gin.Context) *int64 {
	userID := optionalCurrentUserID(c)
	if userID == 0 {
		return nil
	}
	return &userID
}

func isAdminFromCtx(c *gin.Context) bool {
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(int)
	return role == 1
}
