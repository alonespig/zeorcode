package service

import (
	"bytes"
	"strings"
	"testing"

	"zoj/pkg/util"

	"github.com/xuri/excelize/v2"
)

func TestBuildStudentImportTemplate(t *testing.T) {
	data, err := buildStudentImportTemplate()
	if err != nil {
		t.Fatalf("buildStudentImportTemplate() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("buildStudentImportTemplate() returned empty file")
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("template is not a valid xlsx: %v", err)
	}
	defer f.Close()
	if _, err := f.GetSheetIndex(studentImportSheet); err != nil {
		t.Fatalf("missing %q sheet: %v", studentImportSheet, err)
	}
	if _, err := f.GetSheetIndex(studentImportGuideSheet); err != nil {
		t.Fatalf("missing %q sheet: %v", studentImportGuideSheet, err)
	}
	for i, want := range studentImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		got, err := f.GetCellValue(studentImportSheet, cell)
		if err != nil {
			t.Fatalf("GetCellValue(%s) error = %v", cell, err)
		}
		if got != want {
			t.Errorf("header %s = %q, want %q", cell, got, want)
		}
	}
	if got, err := f.GetCellValue(studentImportSheet, "E1"); err != nil || got != "" {
		t.Fatalf("E1 = %q, err = %v; password column should not exist", got, err)
	}
	styleID, err := f.GetCellStyle(studentImportSheet, "A1")
	if err != nil {
		t.Fatalf("GetCellStyle(A1) error = %v", err)
	}
	if styleID != 0 {
		t.Fatalf("A1 style = %d, want default style 0", styleID)
	}
}

func TestParseStudentImport(t *testing.T) {
	data, err := buildStudentImportTemplate()
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	values := []any{" Student01@Example.edu.cn ", "20260001", "张三", "student01"}
	if err := f.SetSheetRow(studentImportSheet, "A2", &values); err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := parseStudentImport(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parseStudentImport() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	got := rows[0]
	if got.Line != 2 || got.Email != "student01@example.edu.cn" || got.StudentNo != "20260001" ||
		got.RealName != "张三" || got.Username != "student01" {
		t.Fatalf("parsed row = %#v", got)
	}
}

func TestParseStudentImportRejectsDuplicateEmail(t *testing.T) {
	data, err := buildStudentImportTemplate()
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	first := []any{"student01@example.edu.cn"}
	second := []any{"STUDENT01@example.edu.cn"}
	if err := f.SetSheetRow(studentImportSheet, "A2", &first); err != nil {
		t.Fatal(err)
	}
	if err := f.SetSheetRow(studentImportSheet, "A3", &second); err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	_, err = parseStudentImport(bytes.NewReader(buf.Bytes()))
	if err == nil || !strings.Contains(err.Error(), "邮箱与第 2 行重复") {
		t.Fatalf("parseStudentImport() error = %v", err)
	}
}

func TestPrepareNewStudentAccountUsesStudentNumberAsPassword(t *testing.T) {
	row := &studentImportRow{
		Line:      2,
		Email:     "student01@example.edu.cn",
		StudentNo: "20260001",
		RealName:  "张三",
		Username:  "student01",
	}
	if err := prepareNewStudentAccount(row); err != nil {
		t.Fatalf("prepareNewStudentAccount() error = %v", err)
	}
	if !util.CheckPassword(row.StudentNo, row.PasswordHash) {
		t.Fatal("password hash does not match student number")
	}
}
