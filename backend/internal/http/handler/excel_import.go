package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/gin-gonic/gin"
)

const maxExcelImportFileSize = 5 << 20

var errExcelImportTooLarge = errcode.New(errcode.InvalidParams, "Excel 文件大小不能超过 5MB")

func openExcelImport(c *gin.Context) (multipart.File, error) {
	// 为 multipart 边界和表单头留出空间。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxExcelImportFileSize+(1<<20))
	fh, err := c.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return nil, errExcelImportTooLarge.Wrap(err)
		}
		return nil, errcode.ErrInvalidParams.WithMsg("请选择要导入的 Excel 文件").Wrap(err)
	}
	if fh.Size == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("Excel 文件不能为空")
	}
	if fh.Size > maxExcelImportFileSize {
		return nil, errExcelImportTooLarge
	}
	if !strings.EqualFold(filepath.Ext(fh.Filename), ".xlsx") {
		return nil, errcode.ErrInvalidParams.WithMsg("仅支持 .xlsx 格式的 Excel 文件")
	}
	file, err := fh.Open()
	if err != nil {
		return nil, errcode.ErrFileUpload.WithMsg("无法读取 Excel 文件").Wrap(err)
	}
	return file, nil
}
