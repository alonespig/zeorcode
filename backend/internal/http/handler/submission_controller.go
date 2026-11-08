package handler

import (
	"fmt"
	"net/http"

	"github.com/alonespig/zeorcode/backend/internal/authz"
	"github.com/alonespig/zeorcode/backend/internal/http/dto"
	"github.com/alonespig/zeorcode/backend/internal/infra/mq"
	"github.com/alonespig/zeorcode/backend/internal/service"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type SubmissionController struct {
	subSrv      *service.SubmissionService
	runSrv      *service.CodeRunService
	contestSrv  *service.ContestService
	mq          *mq.MQ
	permissions *service.PermissionService
}

func NewSubmissionController(
	s *service.SubmissionService,
	runSrv *service.CodeRunService,
	contestSrv *service.ContestService,
	m *mq.MQ,
	permissions *service.PermissionService,
) *SubmissionController {
	return &SubmissionController{
		subSrv: s, runSrv: runSrv, contestSrv: contestSrv, mq: m,
		permissions: permissions,
	}
}

func (s *SubmissionController) Create(c *gin.Context) (gin.H, error) {
	var req dto.SubmitCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userIDStr, exists := c.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	subID, err := s.subSrv.CreateSubmission(
		c.Request.Context(), toSubmitCodeParams(req), userID,
		hasAnyPermission(c, s.permissions, authz.ProblemRead),
	)
	if err != nil {
		return nil, err
	}
	return gin.H{"submissionID": subID}, nil
}

func (s *SubmissionController) Run(c *gin.Context) (*dto.RunCodeResp, error) {
	const maxRequestBytes = 8 * 1024 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)

	var req dto.RunCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userIDValue, exists := c.Get("userID")
	if !exists {
		return nil, errcode.ErrUnauthorized
	}
	userID, ok := userIDValue.(int64)
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("用户ID类型错误")
	}
	result, err := s.runSrv.Run(
		c.Request.Context(), toCodeRunParams(req), userID,
		hasAnyPermission(c, s.permissions, authz.ProblemRead),
	)
	if err != nil {
		return nil, err
	}
	return toCodeRunResp(result), nil
}

func (s *SubmissionController) GetSubmissionList(c *gin.Context) (*dto.SubmitListResp, error) {
	var req dto.SubmissionListForm
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	// 经 JWTAuthOptional 注入：只有拥有提交查看权限的管理员才能看隐藏题的提交
	resp, err := s.subSrv.List(
		c.Request.Context(), req.Page, req.PageSize, req.Status, req.Username,
		req.UserID, req.ProblemID,
		hasAnyPermission(c, s.permissions, authz.SubmissionRead),
	)
	if err != nil {
		return nil, err
	}
	return toSubmissionListResp(resp), nil
}

func (s *SubmissionController) GetSubmissionByID(c *gin.Context) (*dto.SubmissionResp, error) {
	publicID, err := parsePublicIDParam(c, "id", "提交")
	if err != nil {
		return nil, err
	}
	// 经 JWTAuthOptional 注入：未登录时取不到，requesterID=0、无权查看他人提交
	uidVal, _ := c.Get("userID")
	requesterID, _ := uidVal.(int64)
	resp, err := s.subSrv.GetSubmissionByPublicID(
		c.Request.Context(), publicID, requesterID,
		hasAnyPermission(c, s.permissions, authz.SubmissionRead),
	)
	if err != nil {
		return nil, err
	}
	return toSubmissionDetailResp(resp), nil
}

// Rejudge POST /api/admin/submission/:id/rejudge 单条重判（超管）
func (s *SubmissionController) Rejudge(c *gin.Context) (gin.H, error) {
	publicID, err := parsePublicIDParam(c, "id", "提交")
	if err != nil {
		return nil, err
	}
	n, err := s.subSrv.RejudgeSubmission(c.Request.Context(), publicID)
	if err != nil {
		return nil, err
	}
	return gin.H{"rejudged": n}, nil
}

// RejudgeContest POST /api/admin/contest/:id/rejudge 整场重判（超管）
func (s *SubmissionController) RejudgeContest(c *gin.Context) (gin.H, error) {
	publicID, err := parsePublicIDParam(c, "id", "比赛")
	if err != nil {
		return nil, err
	}
	cid, err := s.contestSrv.ResolveAdminID(c.Request.Context(), publicID)
	if err != nil {
		return nil, err
	}
	n, err := s.subSrv.RejudgeContest(c.Request.Context(), cid)
	if err != nil {
		return nil, err
	}
	return gin.H{"rejudged": n}, nil
}

// RejudgeContestProblem POST /api/admin/contest/:id/problem/:pid/rejudge 某比赛某题全部重判（超管）
func (s *SubmissionController) RejudgeContestProblem(c *gin.Context) (gin.H, error) {
	publicID, err := parsePublicIDParam(c, "id", "比赛")
	if err != nil {
		return nil, err
	}
	cid, err := s.contestSrv.ResolveAdminID(c.Request.Context(), publicID)
	if err != nil {
		return nil, err
	}
	pid, err := s.contestSrv.ResolveProblemID(c.Request.Context(), c.Param("pid"))
	if err != nil {
		return nil, err
	}
	n, err := s.subSrv.RejudgeContestProblem(c.Request.Context(), cid, pid)
	if err != nil {
		return nil, err
	}
	return gin.H{"rejudged": n}, nil
}

// GetSubmissionStream SSE 长连接，直接往 c.Writer 推数据，无法套 Wrap
func (s *SubmissionController) GetSubmissionStream(c *gin.Context) {
	publicID, err := parsePublicIDParam(c, "id", "提交")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	requesterVal, _ := c.Get("userID")
	requesterID, _ := requesterVal.(int64)
	if _, err := s.subSrv.GetSubmissionByPublicID(
		c.Request.Context(), publicID, requesterID,
		hasAnyPermission(c, s.permissions, authz.SubmissionRead),
	); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	subID, err := s.subSrv.ResolveSubmissionID(c.Request.Context(), publicID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}
	events, cleanup := s.mq.SubscribeSubmission(c, subID)
	defer cleanup()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			if ev.Done {
				fmt.Fprintf(c.Writer, "event: done\n")
				fmt.Fprintf(c.Writer, "data: %s\n\n", ev.Payload)
				flusher.Flush()
				return
			}
			fmt.Fprintf(c.Writer, "event: judging\n")
			fmt.Fprintf(c.Writer, "data: %s\n\n", ev.Payload)
			flusher.Flush()
		}
	}
}
