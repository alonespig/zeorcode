package service

import (
	"fmt"
	"io"
	"strings"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/xuri/excelize/v2"
)

// readExcelImportRows 校验工作表和表头，保留原始行序供业务校验报告行号。
func readExcelImportRows(reader io.Reader, sheet string, headers []string) ([][]string, error) {
	f, err := excelize.OpenReader(reader, excelize.Options{
		UnzipSizeLimit:    64 << 20,
		UnzipXMLSizeLimit: 16 << 20,
	})
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 文件无法解析，请重新下载模板填写").Wrap(err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg(fmt.Sprintf("Excel 中缺少“%s”工作表", sheet)).Wrap(err)
	}
	if len(rows) == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 中缺少表头")
	}
	for i, want := range headers {
		if strings.TrimSpace(cellValue(rows[0], i)) != want {
			return nil, errcode.ErrInvalidParams.WithMsg("Excel 表头不匹配，请使用最新模板")
		}
	}
	return rows, nil
}
