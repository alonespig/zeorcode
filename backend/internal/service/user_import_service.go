package service

import (
	"context"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"unicode/utf8"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/model"

	"github.com/xuri/excelize/v2"
)

const (
	adminUserImportSheet      = "用户名单"
	adminUserImportGuideSheet = "填写说明"
	adminUserImportMaxRows    = 500
)

var adminUserImportHeaders = []string{
	"邮箱",
	"学号",
	"姓名",
	"用户名",
}

type adminUserImportRow struct {
	Line      int
	Email     string
	StudentNo string
	RealName  string
	Username  string
}

// UserImportTemplate 生成后台用户 Excel 导入模板。
func (u *UserService) UserImportTemplate() ([]byte, error) {
	data, err := buildAdminUserImportTemplate()
	if err != nil {
		return nil, errcode.ErrInternal.WithMsg("生成导入模板失败").Wrap(err)
	}
	return data, nil
}

// ImportUsers 从 Excel 批量创建普通用户，初始密码与学号相同。
func (u *UserService) ImportUsers(ctx context.Context, reader io.Reader) (*dto.BatchCreateUsersResp, error) {
	rows, err := parseAdminUserImport(reader)
	if err != nil {
		return nil, err
	}

	req := &dto.BatchCreateUsersReq{Users: make([]dto.BatchUserItem, 0, len(rows))}
	for _, row := range rows {
		req.Users = append(req.Users, dto.BatchUserItem{
			Username:  row.Username,
			StudentNo: row.StudentNo,
			RealName:  row.RealName,
			Email:     row.Email,
			Password:  row.StudentNo,
			Role:      model.RoleNormal,
		})
	}
	return u.BatchCreateUsers(ctx, req)
}

func parseAdminUserImport(reader io.Reader) ([]adminUserImportRow, error) {
	f, err := excelize.OpenReader(reader, excelize.Options{
		UnzipSizeLimit:    64 << 20,
		UnzipXMLSizeLimit: 16 << 20,
	})
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 文件无法解析，请重新下载模板填写").Wrap(err)
	}
	defer f.Close()

	rows, err := f.GetRows(adminUserImportSheet)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 中缺少“用户名单”工作表").Wrap(err)
	}
	if len(rows) == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 中缺少表头")
	}
	for i, want := range adminUserImportHeaders {
		if strings.TrimSpace(cellValue(rows[0], i)) != want {
			return nil, errcode.ErrInvalidParams.WithMsg("Excel 表头不匹配，请使用最新模板")
		}
	}

	result := make([]adminUserImportRow, 0, len(rows)-1)
	seenEmails := make(map[string]int, len(rows)-1)
	seenStudentNos := make(map[string]int, len(rows)-1)
	seenUsernames := make(map[string]int, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := adminUserImportRow{
			Line:      i + 1,
			Email:     normalizeEmail(cellValue(rows[i], 0)),
			StudentNo: strings.TrimSpace(cellValue(rows[i], 1)),
			RealName:  strings.TrimSpace(cellValue(rows[i], 2)),
			Username:  strings.TrimSpace(cellValue(rows[i], 3)),
		}
		if row.Email == "" && row.StudentNo == "" && row.RealName == "" && row.Username == "" {
			continue
		}
		if err := validateAdminUserImportRow(row); err != nil {
			return nil, err
		}
		if row.Email != "" {
			if firstLine, exists := seenEmails[row.Email]; exists {
				return nil, importRowError(row.Line, fmt.Sprintf("邮箱与第 %d 行重复", firstLine))
			}
			seenEmails[row.Email] = row.Line
		}
		if firstLine, exists := seenStudentNos[row.StudentNo]; exists {
			return nil, importRowError(row.Line, fmt.Sprintf("学号与第 %d 行重复", firstLine))
		}
		usernameKey := strings.ToLower(row.Username)
		if firstLine, exists := seenUsernames[usernameKey]; exists {
			return nil, importRowError(row.Line, fmt.Sprintf("用户名与第 %d 行重复", firstLine))
		}
		seenStudentNos[row.StudentNo] = row.Line
		seenUsernames[usernameKey] = row.Line
		result = append(result, row)
		if len(result) > adminUserImportMaxRows {
			return nil, errcode.ErrInvalidParams.WithMsg(fmt.Sprintf("一次最多导入 %d 个用户", adminUserImportMaxRows))
		}
	}
	if len(result) == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("用户名单中没有可导入的数据")
	}
	return result, nil
}

func validateAdminUserImportRow(row adminUserImportRow) error {
	switch {
	case row.Email != "" && !validImportEmail(row.Email):
		return importRowError(row.Line, "邮箱格式不正确")
	case row.StudentNo == "":
		return importRowError(row.Line, "学号不能为空")
	case utf8.RuneCountInString(row.StudentNo) > 32:
		return importRowError(row.Line, "学号不能超过 32 个字符")
	case len(row.StudentNo) > 72:
		return importRowError(row.Line, "学号作为初始密码时不能超过 72 个字节")
	case row.RealName == "":
		return importRowError(row.Line, "姓名不能为空")
	case utf8.RuneCountInString(row.RealName) > 64:
		return importRowError(row.Line, "姓名不能超过 64 个字符")
	case !usernamePattern.MatchString(row.Username):
		return importRowError(row.Line, "用户名只能包含字母、数字和下划线，长度为 2-20 位")
	default:
		return nil
	}
}

func validImportEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Address == email && utf8.RuneCountInString(email) <= 191
}

func buildAdminUserImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", adminUserImportSheet); err != nil {
		return nil, err
	}
	if _, err := f.NewSheet(adminUserImportGuideSheet); err != nil {
		return nil, err
	}
	if err := f.SetSheetRow(adminUserImportSheet, "A1", &adminUserImportHeaders); err != nil {
		return nil, err
	}

	textStyle, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(adminUserImportSheet, "A2", fmt.Sprintf("D%d", adminUserImportMaxRows+1), textStyle); err != nil {
		return nil, err
	}
	for _, col := range []struct {
		name  string
		width float64
	}{{"A", 30}, {"B", 24}, {"C", 22}, {"D", 25}} {
		if err := f.SetColWidth(adminUserImportSheet, col.name, col.name, col.width); err != nil {
			return nil, err
		}
	}
	if err := f.SetPanes(adminUserImportSheet, &excelize.Panes{
		Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft",
	}); err != nil {
		return nil, err
	}

	guideRows := [][]any{
		{"用户导入说明"},
		{"规则", "说明"},
		{"必填字段", "学号、姓名和用户名必须填写，邮箱可以留空。"},
		{"邮箱", "选填；填写时必须是有效邮箱，且不能与现有账号重复。"},
		{"用户角色", "导入后统一创建为普通用户，如需管理员权限可在用户列表中单独调整。"},
		{"用户名", "仅支持 2-20 位字母、数字和下划线。"},
		{"初始密码", "初始密码与学号相同，导入后请通知用户尽快修改。"},
		{"导入限制", fmt.Sprintf("每次最多 %d 人，仅支持 .xlsx 文件。", adminUserImportMaxRows)},
		{"事务规则", "任意一行校验失败时，本批次不会创建任何账号。"},
		{"填写示例", "student01@example.edu.cn | 20260001 | 张三 | student01（邮箱也可留空）"},
	}
	for i, row := range guideRows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(adminUserImportGuideSheet, cell, &row); err != nil {
			return nil, err
		}
	}
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "1E3A8A"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	guideHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "1E3A8A"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DBEAFE"}},
	})
	if err != nil {
		return nil, err
	}
	if err := f.MergeCell(adminUserImportGuideSheet, "A1", "B1"); err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(adminUserImportGuideSheet, "A1", "B1", titleStyle); err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(adminUserImportGuideSheet, "A2", "B2", guideHeaderStyle); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(adminUserImportGuideSheet, "A", "A", 18); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(adminUserImportGuideSheet, "B", "B", 78); err != nil {
		return nil, err
	}
	if err := f.SetRowHeight(adminUserImportGuideSheet, 1, 32); err != nil {
		return nil, err
	}
	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
