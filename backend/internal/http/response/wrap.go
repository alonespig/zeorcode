package response

import (
	"errors"
	"reflect"

	"github.com/alonespig/zeorcode/backend/internal/infra/logger"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// HandlerFunc 业务 handler 的统一签名：返回具体类型数据 + 错误
// 路由用 Wrap 把它包成 gin.HandlerFunc，错误处理集中收口
type HandlerFunc[T any] func(c *gin.Context) (T, error)

// Wrap 把 (T, error) 风格 handler 转成 gin.HandlerFunc
//   - err != nil：统一 Fail，并按错误级别记录日志
//   - err == nil 且 data 为空：SuccessEmpty
//   - 其余：Success(data)
func Wrap[T any](h HandlerFunc[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c)
		if err != nil {
			logAPIError(c, err)
			Fail(c, err)
			return
		}
		if isNilData(data) {
			SuccessEmpty(c)
			return
		}
		Success(c, data)
	}
}

// isNilData 同时识别 interface nil 和 typed nil（如 (*dto.X)(nil)）
func isNilData(data any) bool {
	if data == nil {
		return true
	}
	value := reflect.ValueOf(data)
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return value.IsNil()
	default:
		return false
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
