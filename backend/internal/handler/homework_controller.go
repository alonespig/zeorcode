package handler

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

// HomeworkController 团队作业接口
type HomeworkController struct {
	hwSrv *service.HomeworkService
}

func NewHomeworkController(hwSrv *service.HomeworkService) *HomeworkController {
	return &HomeworkController{hwSrv: hwSrv}
}

// ListByTeam GET /api/team/:id/homework
func (h *HomeworkController) ListByTeam(c *gin.Context) (any, error) {
	teamID, err := h.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	var page dto.PageForm
	if err := c.ShouldBindQuery(&page); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return h.hwSrv.ListByTeam(c.Request.Context(), teamID,
		optionalCurrentUserID(c), isAdminFromCtx(c), page.Page, page.PageSize)
}

// Create POST /api/team/:id/homework 布置作业
func (h *HomeworkController) Create(c *gin.Context) (any, error) {
	teamID, err := h.resolveTeamID(c)
	if err != nil {
		return nil, err
	}
	var req dto.SaveHomeworkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	id, err := h.hwSrv.Save(c.Request.Context(), teamID, 0, &req, userID, role == 1)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": id}, nil
}

// Detail GET /api/homework/:hid
func (h *HomeworkController) Detail(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	return h.hwSrv.Detail(c.Request.Context(), id, optionalCurrentUserID(c), isAdminFromCtx(c))
}

// Problem GET /api/homework/:hid/problem/:problemID
// 在作业权限与开始时间约束下返回单题题面。
func (h *HomeworkController) Problem(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	problemID := strings.TrimSpace(c.Param("problemID"))
	if problemID == "" {
		return nil, errcode.ErrInvalidParams.WithMsg("非法题号")
	}
	return h.hwSrv.ProblemDetail(c.Request.Context(), id, problemID,
		optionalCurrentUserID(c), isAdminFromCtx(c))
}

// Update PUT /api/homework/:hid 仅布置者本人可改
func (h *HomeworkController) Update(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	var req dto.SaveHomeworkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if _, err := h.hwSrv.Save(c.Request.Context(), 0, id, &req, userID, role == 1); err != nil {
		return nil, err
	}
	return nil, nil
}

// Delete DELETE /api/homework/:hid
func (h *HomeworkController) Delete(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	if err := h.hwSrv.Delete(c.Request.Context(), id, userID, role == 1); err != nil {
		return nil, err
	}
	return nil, nil
}

// Submit POST /api/homework/:hid/submit
func (h *HomeworkController) Submit(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	var req dto.HomeworkSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	userID, role, err := requireCurrentUser(c)
	if err != nil {
		return nil, err
	}
	subID, err := h.hwSrv.Submit(c.Request.Context(), id, &req, userID, role == 1)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": subID}, nil
}

// Rank GET /api/homework/:hid/rank IOI 排行榜
func (h *HomeworkController) Rank(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	return h.hwSrv.Rank(c.Request.Context(), id, optionalCurrentUserID(c), isAdminFromCtx(c))
}

// Submissions GET /api/homework/:hid/submission
// 普通成员只拿到自己的；团队管理员及以上拿到全部。
func (h *HomeworkController) Submissions(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	var q dto.HomeworkSubmissionQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	return h.hwSrv.Submissions(c.Request.Context(), id, optionalCurrentUserID(c), isAdminFromCtx(c), &q)
}

// SubmissionDetail GET /api/homework/:hid/submission/:sid
// 普通成员只能查看自己的代码，团队管理者可以查看全部成员代码。
func (h *HomeworkController) SubmissionDetail(c *gin.Context) (any, error) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		return nil, err
	}
	publicID, err := parsePublicIDParam(c, "sid", "提交")
	if err != nil {
		return nil, err
	}
	return h.hwSrv.SubmissionDetail(c.Request.Context(), id, publicID,
		optionalCurrentUserID(c), isAdminFromCtx(c))
}

// ExportRank GET /api/homework/:hid/rank/export 导出排行榜 CSV。
// 与 DownloadFiles 同款：直接写响应体，不套 response.Wrap（文件下载不走统一 JSON）。
func (h *HomeworkController) ExportRank(c *gin.Context) {
	id, err := h.resolveHomeworkID(c)
	if err != nil {
		c.String(400, "非法作业编号")
		return
	}
	rank, err := h.hwSrv.Rank(c.Request.Context(), id, optionalCurrentUserID(c), isAdminFromCtx(c))
	if err != nil {
		var appErr *errcode.AppErr
		if errors.As(err, &appErr) {
			c.JSON(200, gin.H{"code": int(appErr.Code), "msg": appErr.Msg})
		} else {
			c.JSON(200, gin.H{"code": int(errcode.UnknownError), "msg": "导出失败"})
		}
		return
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM，Excel 打开中文不乱码
	w := csv.NewWriter(&buf)
	header := []string{"名次", "学号", "姓名", "总分", "已通过"}
	header = append(header, rank.ProblemIDs...)
	_ = w.Write(header)
	for _, row := range rank.List {
		record := []string{strconv.Itoa(row.Rank), safeSpreadsheetCell(row.StudentNo), safeSpreadsheetCell(row.RealName), strconv.Itoa(row.TotalScore), strconv.Itoa(row.SolvedCount)}
		for _, cell := range row.Cells {
			record = append(record, strconv.Itoa(cell.Score))
		}
		_ = w.Write(record)
	}
	w.Flush()

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=rank.csv")
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

func safeSpreadsheetCell(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed == "" {
		return value
	}
	switch trimmed[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
}

func (h *HomeworkController) resolveTeamID(c *gin.Context) (int64, error) {
	publicID, err := parsePublicIDParam(c, "id", "团队")
	if err != nil {
		return 0, err
	}
	return h.hwSrv.ResolveTeamID(c.Request.Context(), publicID)
}

func (h *HomeworkController) resolveHomeworkID(c *gin.Context) (int64, error) {
	publicID, err := parsePublicIDParam(c, "hid", "作业")
	if err != nil {
		return 0, err
	}
	return h.hwSrv.ResolveID(c.Request.Context(), publicID)
}
