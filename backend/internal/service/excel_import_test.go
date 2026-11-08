package service

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/xuri/excelize/v2"
)

func TestReadExcelImportRows(t *testing.T) {
	const sheet = "导入名单"
	headers := []string{"学号", "姓名"}
	tests := []struct {
		name      string
		create    bool
		cells     map[string]string
		malformed bool
		wantMsg   string
	}{
		{name: "preserves original row numbers", create: true, cells: map[string]string{"A1": " 学号 ", "B1": "姓名", "A3": "1001", "B3": "张三"}},
		{name: "invalid workbook", malformed: true, wantMsg: "Excel 文件无法解析，请重新下载模板填写"},
		{name: "missing sheet", wantMsg: "Excel 中缺少“导入名单”工作表"},
		{name: "missing headers", create: true, wantMsg: "Excel 中缺少表头"},
		{name: "headers in the wrong order", create: true, cells: map[string]string{"A1": "姓名", "B1": "学号"}, wantMsg: "Excel 表头不匹配，请使用最新模板"},
		{name: "missing header column", create: true, cells: map[string]string{"A1": "学号"}, wantMsg: "Excel 表头不匹配，请使用最新模板"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.malformed {
				_, err := readExcelImportRows(strings.NewReader("not an Excel file"), sheet, headers)
				assertExcelImportError(t, err, tt.wantMsg)
				return
			}

			workbook := excelize.NewFile()
			t.Cleanup(func() { _ = workbook.Close() })
			if tt.create {
				if err := workbook.SetSheetName("Sheet1", sheet); err != nil {
					t.Fatal(err)
				}
			}
			for cell, value := range tt.cells {
				if err := workbook.SetCellValue(sheet, cell, value); err != nil {
					t.Fatal(err)
				}
			}
			buf, err := workbook.WriteToBuffer()
			if err != nil {
				t.Fatal(err)
			}
			rows, err := readExcelImportRows(buf, sheet, headers)
			if tt.wantMsg != "" {
				assertExcelImportError(t, err, tt.wantMsg)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 3 || len(rows[1]) != 0 || !slices.Equal(rows[2], []string{"1001", "张三"}) {
				t.Fatalf("expected data on Excel row 3, got %#v", rows)
			}
		})
	}
}

func assertExcelImportError(t *testing.T, err error, wantMsg string) {
	t.Helper()
	var appErr *errcode.AppErr
	if !errors.As(err, &appErr) || appErr.Code != errcode.InvalidParams || appErr.Msg != wantMsg {
		t.Fatalf("expected invalid params (%q), got %v", wantMsg, err)
	}
}
