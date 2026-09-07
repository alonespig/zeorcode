package handler

import (
	"net/http"
	"time"

	"zoj/internal/common/consts"
	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

type ContestController struct {
	contestSrv *service.ContestService
}

func NewContestController(contestSrv *service.ContestService) *ContestController {
	return &ContestController{contestSrv: contestSrv}
}

func (c *ContestController) Create(ctx *gin.Context) (any, error) {
	var form dto.CreateContestReq
	if err := ctx.ShouldBindJSON(&form); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return c.contestSrv.CreateContest(ctx.Request.Context(), &form)
}

func (c *ContestController) EditContest(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	return c.contestSrv.EditContest(ctx.Request.Context(), contestID)
}

func (c *ContestController) Update(ctx *gin.Context) (any, error) {
	var form dto.UpdateContestReq
	if err := ctx.ShouldBindJSON(&form); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	id, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	return c.contestSrv.UpdateContest(ctx.Request.Context(), id, &form)
}

func (c *ContestController) List(ctx *gin.Context) (any, error) {
	// 允许匿名浏览：未登录时 userID=0，isRegistered 一律 false
	var userID int64
	if v, exists := ctx.Get("userID"); exists {
		userID, _ = v.(int64)
	}

	var req dto.ContestListForm
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return c.contestSrv.ListContests(ctx.Request.Context(), userID, req.Page, req.PageSize, req.Keyword, req.Type, req.Status)
}

func (c *ContestController) JoinContest(ctx *gin.Context) (any, error) {
	var form dto.JoinContestReq
	if err := ctx.ShouldBindJSON(&form); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	contestID, err := c.contestSrv.ResolveID(ctx.Request.Context(), form.ContestID)
	if err != nil {
		return nil, err
	}
	if err := c.contestSrv.JoinContest(ctx.Request.Context(), userID, contestID, form.InviteCode); err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *ContestController) GetContestDetail(ctx *gin.Context) (any, error) {
	id, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	var userID int64
	if v, exists := ctx.Get("userID"); exists {
		userID, _ = v.(int64)
	}
	return c.contestSrv.GetContestDetail(ctx.Request.Context(), id, userID)
}

func (c *ContestController) GetContestDesc(ctx *gin.Context) (any, error) {
	id, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	contest, err := c.contestSrv.GetContestDesc(ctx.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	return contest, nil
}

func (c *ContestController) GetContestProblemList(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	roleVal, _ := ctx.Get("role")
	role, _ := roleVal.(int)
	return c.contestSrv.GetContestProblemList(ctx.Request.Context(), contestID, userID, role == 1)
}

func (c *ContestController) GetContestProblem(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	roleVal, _ := ctx.Get("role")
	role, _ := roleVal.(int)
	return c.contestSrv.GetContestProblemDetail(
		ctx.Request.Context(), contestID, ctx.Param("problemID"), userID, role == 1,
	)
}

func (c *ContestController) Submit(ctx *gin.Context) (any, error) {
	var form dto.ContestSubmitReq
	if err := ctx.ShouldBindJSON(&form); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, _ := userIDStr.(int64)
	contestID, err := c.contestSrv.ResolveID(ctx.Request.Context(), form.ContestID)
	if err != nil {
		return nil, err
	}
	form.ContestID = contestID
	subID, err := c.contestSrv.SubmitContestProblem(ctx.Request.Context(), &form, userID)
	if err != nil {
		return nil, err
	}
	return gin.H{"submissionID": subID}, nil
}

func (c *ContestController) GetContestSubmitInfo(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	roleVal, _ := ctx.Get("role")
	role, _ := roleVal.(int)
	return c.contestSrv.GetContestSubmitInfo(ctx.Request.Context(), contestID, userID, role == 1)
}

func (c *ContestController) GetContestSubmissions(ctx *gin.Context) (any, error) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, _ := userIDStr.(int64)
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	roleVal, _ := ctx.Get("role")
	role, _ := roleVal.(int)
	var q dto.ContestSubmissionQuery
	_ = ctx.ShouldBindQuery(&q)
	return c.contestSrv.GetContestSubmissions(ctx.Request.Context(), contestID, userID, role == 1, &q)
}

func (c *ContestController) GetContestRank(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	v, _ := ctx.Get("userID")
	userID, _ := v.(int64)
	return c.contestSrv.GetContestRank(ctx.Request.Context(), contestID, userID)
}

// RecomputeContest POST /api/admin/contest/:id/recompute
// 管理员从 submissions 明细整场重算比赛榜（重判后修正、或榜单漂移时手动重建）。
func (c *ContestController) RecomputeContest(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	return nil, c.contestSrv.RecomputeContest(ctx.Request.Context(), contestID)
}

// GetMyContestRank GET /api/contest/:id/myrank 当前用户在该比赛的名次
func (c *ContestController) GetMyContestRank(ctx *gin.Context) (any, error) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		return nil, err
	}
	v, _ := ctx.Get("userID")
	userID, _ := v.(int64)
	return c.contestSrv.GetMyContestRank(ctx.Request.Context(), contestID, userID)
}

type MatchEvent struct {
	Status consts.ContestStatus `json:"status"`
	Now    int64                `json:"now"`
}

// SSEventStream SSE 长连接，无法套 Wrap
func (c *ContestController) SSEventStream(ctx *gin.Context) {
	contestID, err := c.resolveContestID(ctx)
	if err != nil {
		ctx.String(400, "invalid contest id")
		return
	}
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	contest, err := c.contestSrv.GetContestDetail(ctx.Request.Context(), contestID, 0)
	if err != nil || contest == nil {
		ctx.String(500, "contest not found")
		return
	}
	flusher := ctx.Writer.(http.Flusher)

	ctx.SSEvent("init", gin.H{
		"startTime": contest.StartTime,
		"endTime":   contest.EndTime,
		"now":       time.Now().UnixMilli(),
	})
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case now := <-ticker.C:
			nowMs := now.UnixMilli()
			status := consts.ContestFinished
			if nowMs < contest.StartTime {
				status = consts.ContestNotStarted
			} else if nowMs < contest.EndTime {
				status = consts.ContestRunning
			}
			ctx.SSEvent("contest", MatchEvent{Status: status, Now: nowMs})
			flusher.Flush()
		}
	}
}

func (c *ContestController) resolveContestID(ctx *gin.Context) (int64, error) {
	publicID, err := parsePublicIDParam(ctx, "id", "比赛")
	if err != nil {
		return 0, err
	}
	return c.contestSrv.ResolveID(ctx.Request.Context(), publicID)
}
