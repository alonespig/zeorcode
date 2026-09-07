package service

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestBuildAdminUserImportTemplate(t *testing.T) {
	data, err := buildAdminUserImportTemplate()
	if err != nil {
		t.Fatalf("buildAdminUserImportTemplate() error = %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("template is not a valid xlsx: %v", err)
	}
	defer f.Close()

	for i, want := range adminUserImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		got, err := f.GetCellValue(adminUserImportSheet, cell)
		if err != nil || got != want {
			t.Fatalf("header %s = %q, err = %v, want %q", cell, got, err, want)
		}
	}
	if got, err := f.GetCellValue(adminUserImportSheet, "E1"); err != nil || got != "" {
		t.Fatalf("E1 = %q, err = %v; password column should not exist", got, err)
	}
	styleID, err := f.GetCellStyle(adminUserImportSheet, "A1")
	if err != nil || styleID != 0 {
		t.Fatalf("A1 style = %d, err = %v, want default style 0", styleID, err)
	}
}

func TestParseAdminUserImport(t *testing.T) {
	data, err := buildAdminUserImportTemplate()
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	values := []any{" Student01@Example.edu.cn ", "20260001", "张三", "student01"}
	if err := f.SetSheetRow(adminUserImportSheet, "A2", &values); err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := parseAdminUserImport(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parseAdminUserImport() error = %v", err)
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

func TestParseAdminUserImportAllowsEmptyEmail(t *testing.T) {
	data, err := buildAdminUserImportTemplate()
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	first := []any{"", "20260001", "张三", "student01"}
	second := []any{"", "20260002", "李四", "student02"}
	if err := f.SetSheetRow(adminUserImportSheet, "A2", &first); err != nil {
		t.Fatal(err)
	}
	if err := f.SetSheetRow(adminUserImportSheet, "A3", &second); err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := parseAdminUserImport(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parseAdminUserImport() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Email != "" || rows[1].Email != "" {
		t.Fatalf("empty emails were not preserved: %#v", rows)
	}
}

func TestParseAdminUserImportRejectsDuplicateUsernameIgnoringCase(t *testing.T) {
	data, err := buildAdminUserImportTemplate()
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	first := []any{"one@example.edu.cn", "20260001", "张三", "Student01"}
	second := []any{"two@example.edu.cn", "20260002", "李四", "student01"}
	if err := f.SetSheetRow(adminUserImportSheet, "A2", &first); err != nil {
		t.Fatal(err)
	}
	if err := f.SetSheetRow(adminUserImportSheet, "A3", &second); err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	_, err = parseAdminUserImport(bytes.NewReader(buf.Bytes()))
	if err == nil || !strings.Contains(err.Error(), "用户名与第 2 行重复") {
		t.Fatalf("parseAdminUserImport() error = %v", err)
	}
}
