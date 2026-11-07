package middleware

import (
	"fmt"
	"time"

	"github.com/alonespig/zeorcode/backend/pkg/errcode"
	"github.com/alonespig/zeorcode/backend/internal/http/response"
	"github.com/alonespig/zeorcode/backend/internal/infra/cache"

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

// RateLimitUserResource 按“用户 + 路径参数”限流，用于邀请码这类可被撞库的入口。
// 例如团队加入按 userID + 团队对外 ID 计数，避免短邀请码被同一账号反复尝试。
func RateLimitUserResource(c *cache.Cache, scope string, limit int, window time.Duration, param string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, _ := ctx.Get("userID")
		userID, _ := v.(int64)
		if userID <= 0 {
			ctx.Next()
			return
		}
		key := fmt.Sprintf("ratelimit:%s:%d:%s", scope, userID, ctx.Param(param))
		n, err := c.RateLimitIncr(ctx.Request.Context(), key, window)
		if err != nil {
			ctx.Next()
			return
		}
		if limit > 0 && n > int64(limit) {
			response.Fail(ctx, errcode.ErrTooManyRequests.WithMsg("尝试次数过多，请稍后再试"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
