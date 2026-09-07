package handler

import (
	"strconv"

	"zoj/internal/common/errcode"
	"zoj/internal/common/publicid"

	"github.com/gin-gonic/gin"
)

func parsePublicIDParam(c *gin.Context, name, label string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || !publicid.Valid(id) {
		return 0, errcode.ErrInvalidParams.WithMsg("非法" + label + "编号")
	}
	return id, nil
}
