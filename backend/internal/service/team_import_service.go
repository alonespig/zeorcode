package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"unicode/utf8"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/model"
	"zoj/internal/repository"
	"zoj/pkg/util"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const (
	studentImportSheet      = "学生名单"
	studentImportGuideSheet = "填写说明"
	studentImportMaxRows    = 500
)

var studentImportHeaders = []string{
	"邮箱",
	"学号",
	"姓名",
	"用户名",
}

type studentImportRow struct {
	Line         int
	Email        string
	StudentNo    string
	RealName     string
	Username     string
	PasswordHash string
}

// StudentImportTemplate 生成学生名单导入模板，仅团队所有者、团队管理员和站点管理员可下载。
func (s *TeamService) StudentImportTemplate(ctx context.Context, teamID, actorID int64, isSiteAdmin bool) ([]byte, error) {
	if err := s.requireManageAccess(ctx, teamID, actorID, isSiteAdmin); err != nil {
		return nil, err
	}
	data, err := buildStudentImportTemplate()
	if err != nil {
		return nil, errcode.ErrInternal.WithMsg("生成导入模板失败").Wrap(err)
	}
	return data, nil
}

// ImportStudents 按邮箱识别用户：邮箱存在则直接入队，不存在则按名单创建普通用户后入队。
// 整批创建账号和添加成员使用同一事务，任何一行失败都会全部回滚。
func (s *TeamService) ImportStudents(
	ctx context.Context,
	teamID, actorID int64,
	isSiteAdmin bool,
	reader io.Reader,
) (*dto.TeamStudentImportResp, error) {
	if err := s.requireManageAccess(ctx, teamID, actorID, isSiteAdmin); err != nil {
		return nil, err
	}
	rows, err := parseStudentImport(reader)
	if err != nil {
		return nil, err
	}

	// 密码哈希较耗时，先在事务外完成，避免长时间占用数据库连接和锁。
	emails := importEmails(rows)
	knownUsers, err := s.userRepo.FindByEmails(ctx, emails)
	if err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	knownByEmail := usersByEmail(knownUsers)
	for i := range rows {
		if _, exists := knownByEmail[rows[i].Email]; exists {
			continue
		}
		if err := prepareNewStudentAccount(&rows[i]); err != nil {
			return nil, err
		}
	}

	result := &dto.TeamStudentImportResp{Total: len(rows)}
	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		teamRepo := repository.NewTeamRepo(tx)
		userRepo := repository.NewUserRepo(tx)

		// 在事务内再次按邮箱读取，处理预检后恰好被其他请求创建的账号。
		existingUsers, err := userRepo.FindByEmails(ctx, emails)
		if err != nil {
			return err
		}
		userByEmail := usersByEmail(existingUsers)

		newRows := make([]*studentImportRow, 0, len(rows))
		for i := range rows {
			row := &rows[i]
			if _, exists := userByEmail[row.Email]; exists {
				continue
			}
			if row.PasswordHash == "" {
				return errcode.ErrOperationDenied.WithMsg(fmt.Sprintf("第 %d 行账号状态已变化，请重新导入", row.Line))
			}
			newRows = append(newRows, row)
		}
		if err := validateNewAccountUniqueness(ctx, userRepo, newRows); err != nil {
			return err
		}

		newUsers := make([]*model.User, 0, len(newRows))
		reservedUIDs := make(map[int64]struct{}, len(newRows))
		for _, row := range newRows {
			uid, err := genUniqueImportUID(ctx, userRepo, reservedUIDs)
			if err != nil {
				return err
			}
			reservedUIDs[uid] = struct{}{}
			newUsers = append(newUsers, &model.User{
				UID:       uid,
				Role:      model.RoleNormal,
				Username:  row.Username,
				StudentNo: studentNoPtr(row.StudentNo),
				RealName:  row.RealName,
				Password:  row.PasswordHash,
				Email:     emailPtr(row.Email),
			})
		}
		if err := userRepo.CreateUsersBatch(ctx, newUsers); err != nil {
			return err
		}
		for i, row := range newRows {
			userByEmail[row.Email] = newUsers[i]
		}

		members, err := teamRepo.ListMembers(ctx, teamID)
		if err != nil {
			return err
		}
		memberIDs := make(map[int64]struct{}, len(members))
		for _, member := range members {
			memberIDs[member.UserID] = struct{}{}
		}

		added := 0
		skipped := 0
		for _, row := range rows {
			user := userByEmail[row.Email]
			if _, exists := memberIDs[user.ID]; exists {
				skipped++
				continue
			}
			if err := teamRepo.AddMember(ctx, teamID, user.ID); err != nil {
				return err
			}
			memberIDs[user.ID] = struct{}{}
			added++
		}

		result.CreatedUsers = len(newUsers)
		result.AddedMembers = added
		result.SkippedMembers = skipped
		return nil
	})
	if err != nil {
		var appErr *errcode.AppErr
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	return result, nil
}

func (s *TeamService) requireManageAccess(ctx context.Context, teamID, actorID int64, isSiteAdmin bool) error {
	if _, err := s.getTeam(ctx, teamID); err != nil {
		return err
	}
	access, err := s.Access(ctx, teamID, actorID, isSiteAdmin)
	if err != nil {
		return err
	}
	if !access.CanManage() {
		return errcode.ErrPermissionDenied.WithMsg("只有团队所有者和管理员能导入学生名单")
	}
	return nil
}

func parseStudentImport(reader io.Reader) ([]studentImportRow, error) {
	f, err := excelize.OpenReader(reader, excelize.Options{
		UnzipSizeLimit:    64 << 20,
		UnzipXMLSizeLimit: 16 << 20,
	})
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 文件无法解析，请重新下载模板填写").Wrap(err)
	}
	defer f.Close()

	rows, err := f.GetRows(studentImportSheet)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 中缺少“学生名单”工作表").Wrap(err)
	}
	if len(rows) == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 中缺少表头")
	}
	for i, want := range studentImportHeaders {
		if strings.TrimSpace(cellValue(rows[0], i)) != want {
			return nil, errcode.ErrInvalidParams.WithMsg("Excel 表头不匹配，请使用最新模板")
		}
	}

	result := make([]studentImportRow, 0, len(rows)-1)
	seenEmails := make(map[string]int, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := studentImportRow{
			Line:      i + 1,
			Email:     normalizeEmail(cellValue(rows[i], 0)),
			StudentNo: strings.TrimSpace(cellValue(rows[i], 1)),
			RealName:  strings.TrimSpace(cellValue(rows[i], 2)),
			Username:  strings.TrimSpace(cellValue(rows[i], 3)),
		}
		if row.Email == "" && row.StudentNo == "" && row.RealName == "" && row.Username == "" {
			continue
		}
		if row.Email == "" {
			return nil, importRowError(row.Line, "邮箱不能为空")
		}
		parsed, err := mail.ParseAddress(row.Email)
		if err != nil || parsed.Address != row.Email || utf8.RuneCountInString(row.Email) > 191 {
			return nil, importRowError(row.Line, "邮箱格式不正确")
		}
		if firstLine, exists := seenEmails[row.Email]; exists {
			return nil, importRowError(row.Line, fmt.Sprintf("邮箱与第 %d 行重复", firstLine))
		}
		seenEmails[row.Email] = row.Line
		result = append(result, row)
		if len(result) > studentImportMaxRows {
			return nil, errcode.ErrInvalidParams.WithMsg(fmt.Sprintf("一次最多导入 %d 名学生", studentImportMaxRows))
		}
	}
	if len(result) == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("学生名单中没有可导入的数据")
	}
	return result, nil
}

func prepareNewStudentAccount(row *studentImportRow) error {
	switch {
	case row.StudentNo == "":
		return importRowError(row.Line, "邮箱未注册，创建账号时学号不能为空")
	case utf8.RuneCountInString(row.StudentNo) > 32:
		return importRowError(row.Line, "学号不能超过 32 个字符")
	case len(row.StudentNo) > 72:
		return importRowError(row.Line, "学号作为初始密码时不能超过 72 个字节")
	case row.RealName == "":
		return importRowError(row.Line, "邮箱未注册，创建账号时姓名不能为空")
	case utf8.RuneCountInString(row.RealName) > 64:
		return importRowError(row.Line, "姓名不能超过 64 个字符")
	case !usernamePattern.MatchString(row.Username):
		return importRowError(row.Line, "用户名只能包含字母、数字和下划线，长度为 2-20 位")
	}
	hash, err := util.HashPassword(row.StudentNo)
	if err != nil {
		return errcode.ErrInternal.WithMsg(fmt.Sprintf("第 %d 行密码处理失败", row.Line)).Wrap(err)
	}
	row.PasswordHash = hash
	return nil
}

// genUniqueImportUID 同时避开数据库内 UID 和本批尚未落库的 UID。
func genUniqueImportUID(ctx context.Context, repo *repository.UserRepo, reserved map[int64]struct{}) (int64, error) {
	for i := 0; i < 10; i++ {
		uid, err := genUniqueUID(ctx, repo)
		if err != nil {
			return 0, err
		}
		if _, exists := reserved[uid]; !exists {
			return uid, nil
		}
	}
	return 0, errcode.ErrInternal.WithMsg("生成用户号失败，请重试")
}

func validateNewAccountUniqueness(ctx context.Context, repo *repository.UserRepo, rows []*studentImportRow) error {
	if len(rows) == 0 {
		return nil
	}
	names := make([]string, 0, len(rows))
	studentNos := make([]string, 0, len(rows))
	nameLine := make(map[string]int, len(rows))
	studentNoLine := make(map[string]int, len(rows))
	for _, row := range rows {
		nameKey := strings.ToLower(row.Username)
		if firstLine, exists := nameLine[nameKey]; exists {
			return importRowError(row.Line, fmt.Sprintf("用户名与第 %d 行重复", firstLine))
		}
		if firstLine, exists := studentNoLine[row.StudentNo]; exists {
			return importRowError(row.Line, fmt.Sprintf("学号与第 %d 行重复", firstLine))
		}
		nameLine[nameKey] = row.Line
		studentNoLine[row.StudentNo] = row.Line
		names = append(names, row.Username)
		studentNos = append(studentNos, row.StudentNo)
	}
	existing, err := repo.CountByUsernames(ctx, names)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return errcode.ErrUserExists.WithMsg("用户名已存在，且属于其他邮箱: " + existing[0])
	}
	existing, err = repo.CountByStudentNos(ctx, studentNos)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return errcode.ErrUserExists.WithMsg("学号已存在，且属于其他邮箱: " + existing[0])
	}
	return nil
}

func buildStudentImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", studentImportSheet); err != nil {
		return nil, err
	}
	if _, err := f.NewSheet(studentImportGuideSheet); err != nil {
		return nil, err
	}
	if err := f.SetSheetRow(studentImportSheet, "A1", &studentImportHeaders); err != nil {
		return nil, err
	}

	textStyle, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(studentImportSheet, "A2", fmt.Sprintf("D%d", studentImportMaxRows+1), textStyle); err != nil {
		return nil, err
	}
	for _, col := range []struct {
		name  string
		width float64
	}{{"A", 30}, {"B", 24}, {"C", 22}, {"D", 25}} {
		if err := f.SetColWidth(studentImportSheet, col.name, col.name, col.width); err != nil {
			return nil, err
		}
	}
	if err := f.SetPanes(studentImportSheet, &excelize.Panes{
		Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft",
	}); err != nil {
		return nil, err
	}
	guideRows := [][]any{
		{"学生名单导入说明"},
		{"规则", "说明"},
		{"账号判断", "系统只按邮箱判断账号是否存在。"},
		{"已有账号", "只需填写邮箱；其他列不会覆盖该账号的已有资料。"},
		{"新账号", "填写邮箱、学号、姓名和用户名后，系统会创建普通账号。"},
		{"用户名", "仅支持 2-20 位字母、数字和下划线。"},
		{"账号密码", "新账号的初始密码与学号相同，导入后请通知学生尽快修改。"},
		{"导入限制", fmt.Sprintf("每次最多 %d 人，仅支持 .xlsx 文件。", studentImportMaxRows)},
		{"事务规则", "任意一行校验失败时，本批次不会创建账号，也不会添加成员。"},
		{"新账号示例", "student01@example.edu.cn | 20260001 | 张三 | student01"},
	}
	for i, row := range guideRows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(studentImportGuideSheet, cell, &row); err != nil {
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
	if err := f.MergeCell(studentImportGuideSheet, "A1", "B1"); err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(studentImportGuideSheet, "A1", "B1", titleStyle); err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(studentImportGuideSheet, "A2", "B2", guideHeaderStyle); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(studentImportGuideSheet, "A", "A", 18); err != nil {
		return nil, err
	}
	if err := f.SetColWidth(studentImportGuideSheet, "B", "B", 78); err != nil {
		return nil, err
	}
	if err := f.SetRowHeight(studentImportGuideSheet, 1, 32); err != nil {
		return nil, err
	}
	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func importEmails(rows []studentImportRow) []string {
	emails := make([]string, 0, len(rows))
	for _, row := range rows {
		emails = append(emails, row.Email)
	}
	return emails
}

func usersByEmail(users []model.User) map[string]*model.User {
	result := make(map[string]*model.User, len(users))
	for i := range users {
		result[normalizeEmail(emailValue(users[i].Email))] = &users[i]
	}
	return result
}

func cellValue(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func importRowError(line int, message string) error {
	return errcode.ErrInvalidParams.WithMsg(fmt.Sprintf("第 %d 行：%s", line, message))
}
