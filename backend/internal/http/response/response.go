package response

import (
	"errors"
	"net/http"

	"zoj/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

const (
	outcomeSuccessKey = "response.outcome.success"
	outcomeCodeKey    = "response.outcome.code"
)

func setOutcome(c *gin.Context, success bool, code int) {
	c.Set(outcomeSuccessKey, success)
	c.Set(outcomeCodeKey, code)
}

func Outcome(c *gin.Context) (bool, int) {
	success, _ := c.Get(outcomeSuccessKey)
	code, _ := c.Get(outcomeCodeKey)
	result, _ := success.(bool)
	value, _ := code.(int)
	return result, value
}

// Success 业务成功
func Success(c *gin.Context, data any) {
	setOutcome(c, true, int(errcode.Success))
	c.JSON(http.StatusOK, Response{
		Code: int(errcode.Success),
		Msg:  "success",
		Data: data,
	})
}

// SuccessEmpty 仅成功无数据
func SuccessEmpty(c *gin.Context) {
	setOutcome(c, true, int(errcode.Success))
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
		setOutcome(c, false, int(appErr.Code))
		c.JSON(http.StatusOK, Response{
			Code: int(appErr.Code),
			Msg:  appErr.Msg,
		})
		return
	}
	setOutcome(c, false, int(errcode.UnknownError))
	c.JSON(http.StatusOK, Response{
		Code: int(errcode.UnknownError),
		Msg:  "未知错误",
	})
}
