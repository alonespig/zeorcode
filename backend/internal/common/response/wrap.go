package response

import (
	"errors"

	"zoj/internal/common/errcode"
	"zoj/internal/common/logger"

	"github.com/gin-gonic/gin"
)

// HandlerFunc 业务 handler 的统一签名：返回数据 + 错误
// 路由用 Wrap 把它包成 gin.HandlerFunc，错误处理集中收口
type HandlerFunc func(c *gin.Context) (any, error)

// Wrap 把 (any, error) 风格 handler 转成 gin.HandlerFunc
// - err != nil：统一 Fail，并按错误级别记录日志
// - err == nil 且 data != nil：Success(data)
// - err == nil 且 data == nil：SuccessEmpty
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c)
		if err != nil {
			logAPIError(c, err)
			Fail(c, err)
			return
		}
		if data == nil {
			SuccessEmpty(c)
			return
		}
		Success(c, data)
	}
}

// logAPIError 按错误类型选日志级别
//   - AppErr 是预期的业务错误：Warn 级别
//   - 非 AppErr 是没归类的内部错误：Error 级别
func logAPIError(c *gin.Context, err error) {
	var appErr *errcode.AppErr
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	if errors.As(err, &appErr) {
		logger.Warnw("api fail",
			"path", path,
			"code", appErr.Code,
			"msg", appErr.Msg,
			"cause", appErr.Cause(),
		)
		return
	}
	logger.Errorw("api error",
		"path", path,
		"err", err,
	)
}
