package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zoj/internal/common/errcode"
	"zoj/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type UploadController struct{}

const maxImageUploadSize = 5 << 20

func NewUploadController() *UploadController {
	return &UploadController{}
}

func (u *UploadController) UploadImage(c *gin.Context) (any, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageUploadSize+(1<<20))
	file, err := c.FormFile("file")
	if err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	if file.Size <= 0 || file.Size > maxImageUploadSize {
		return nil, errcode.ErrInvalidParams.WithMsg("图片大小不能超过5MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return nil, errcode.ErrInvalidParams.WithMsg("只能上传PNG、JPG或JPEG格式的图片")
	}

	src, err := file.Open()
	if err != nil {
		return nil, errcode.ErrFileUpload.Wrap(err)
	}
	defer src.Close()
	header := make([]byte, 512)
	n, readErr := src.Read(header)
	if readErr != nil && readErr != io.EOF {
		return nil, errcode.ErrFileUpload.Wrap(readErr)
	}
	mimeType := http.DetectContentType(header[:n])
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		return nil, errcode.ErrInvalidParams.WithMsg("文件内容不是有效的PNG或JPG图片")
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, errcode.ErrFileUpload.Wrap(err)
	}

	fileName := util.MD5(fmt.Sprintf("%s%x", file.Filename, time.Now().UnixNano())) + ext
	const uploadDir = "./uploads"
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		return nil, errcode.ErrFileUpload.Wrap(err)
	}
	savePath := filepath.Join(uploadDir, fileName)
	dst, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return nil, errcode.ErrFileUpload.Wrap(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(savePath)
		return nil, errcode.ErrFileUpload.Wrap(err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(savePath)
		return nil, errcode.ErrFileUpload.Wrap(err)
	}

	return gin.H{
		"url": viper.GetString("server.base_url") + "/uploads/" + fileName,
	}, nil
}
