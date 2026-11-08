package response

import (
	"reflect"

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
