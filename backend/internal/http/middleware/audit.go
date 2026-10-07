package middleware

import (
	"context"
	"strings"

	"zoj/internal/http/response"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

func AdminAudit(audit *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return
		}
		actorID, _ := c.Get("userID")
		id, _ := actorID.(int64)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		targets := make([]string, 0, len(c.Params))
		for _, param := range c.Params {
			targets = append(targets, param.Key+"="+param.Value)
		}
		success, code := response.Outcome(c)
		audit.Record(context.WithoutCancel(c.Request.Context()), service.AuditRecordParams{
			ActorID: id, Method: method, Path: path, Target: strings.Join(targets, ","),
			ClientIP: c.ClientIP(), Success: success, Code: code,
		})
	}
}
