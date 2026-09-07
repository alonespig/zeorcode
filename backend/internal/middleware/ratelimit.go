package middleware

import (
	"fmt"
	"time"

	"zoj/internal/common/errcode"
	"zoj/internal/common/response"
	"zoj/internal/infra/cache"

	"github.com/gin-gonic/gin"
)

// RateLimit 基于 Redis 固定窗口的按 IP 限流：同一 IP 在 window 内最多 limit 次，
// 超过返回 429（ErrTooManyRequests）。scope 用于区分不同接口的计数桶（如 "global"/"login"）。
// cache 异常/不可用时 fail-open 放行——限流组件不应把正常请求也拦下。
func RateLimit(c *cache.Cache, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", scope, ctx.ClientIP())
		n, err := c.RateLimitIncr(ctx.Request.Context(), key, window)
		if err != nil {
			ctx.Next()
			return
		}
		if limit > 0 && n > int64(limit) {
			response.Fail(ctx, errcode.ErrTooManyRequests)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
