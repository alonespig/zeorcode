package handler

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/gin-gonic/gin"
)

func TestOpenExcelImport(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		size         int
		wantMsg      string
		wantMaxBytes bool
	}{
		{name: "accepts case insensitive extension at file limit", filename: "users.XLSX", size: 5 << 20},
		{name: "missing file", wantMsg: "请选择要导入的 Excel 文件"},
		{name: "empty file", filename: "users.xlsx", wantMsg: "Excel 文件不能为空"},
		{name: "unsupported extension", filename: "users.csv", size: 1, wantMsg: "仅支持 .xlsx 格式的 Excel 文件"},
		{name: "above file limit", filename: "users.xlsx", size: (5 << 20) + 1, wantMsg: "Excel 文件大小不能超过 5MB"},
		{name: "above request limit", filename: "users.xlsx", size: 8 << 20, wantMsg: "Excel 文件大小不能超过 5MB", wantMaxBytes: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			if tt.filename != "" {
				part, err := writer.CreateFormFile("file", tt.filename)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := part.Write(make([]byte, tt.size)); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/import", &body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			file, err := openExcelImport(c)
			if c.Request.MultipartForm != nil {
				t.Cleanup(func() { _ = c.Request.MultipartForm.RemoveAll() })
			}
			if file != nil {
				defer file.Close()
			}
			if tt.wantMsg != "" {
				var appErr *errcode.AppErr
				if file != nil || !errors.As(err, &appErr) || appErr.Code != errcode.InvalidParams || appErr.Msg != tt.wantMsg {
					t.Fatalf("expected no file and invalid params (%q), got file=%v err=%v", tt.wantMsg, file, err)
				}
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) != tt.wantMaxBytes {
					t.Fatalf("unexpected request size error cause: %v", err)
				}
				return
			}
			if err != nil || file == nil {
				t.Fatalf("expected an open file, got %v", err)
			}
			if size, err := io.Copy(io.Discard, file); err != nil || size != int64(tt.size) {
				t.Fatalf("read %d bytes, err=%v, want %d bytes", size, err, tt.size)
			}
		})
	}
}
