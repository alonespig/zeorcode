package response

import (
	"errors"
	"net/http"

	"zoj/internal/common/errcode"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// Success 业务成功
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: int(errcode.Success),
		Msg:  "success",
		Data: data,
	})
}

// SuccessEmpty 仅成功无数据
func SuccessEmpty(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: int(errcode.Success),
		Msg:  "success",
	})
}

// Fail 业务失败：从 *AppErr 自动解出 code/msg；非 AppErr 走未知错误
// 统一返回 HTTP 200，错误语义由 body.code 表达
func Fail(c *gin.Context, err error) {
	var appErr *errcode.AppErr
	if errors.As(err, &appErr) {
		c.JSON(http.StatusOK, Response{
			Code: int(appErr.Code),
			Msg:  appErr.Msg,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: int(errcode.UnknownError),
		Msg:  "未知错误",
	})
}
