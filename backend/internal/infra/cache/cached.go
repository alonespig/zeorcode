package cache

import (
	"context"
	"time"
)

// Cached 通用 cache-aside 封装：
//   - 命中缓存直接反序列化返回；
//   - 未命中则调用 loader 取数据，成功后写回缓存再返回；
//   - 缓存不可用（c 为 nil / Redis 报错）时自动降级为直接调 loader，不影响主流程；
//   - loader 出错不写缓存。
//
// 用法：
//
//	resp, err := cache.Cached(ctx, c, rediskey, 5*time.Minute, func() (*dto.Xxx, error) {
//	    return svc.loadFromDB(...)
//	})
func Cached[T any](ctx context.Context, c *Cache, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	if c != nil {
		var cached T
		if ok, err := c.GetJSON(ctx, key, &cached); err == nil && ok {
			return cached, nil
		}
	}

	value, err := loader()
	if err != nil {
		var zero T
		return zero, err
	}

	if c != nil {
		_ = c.SetJSON(ctx, key, value, ttl)
	}
	return value, nil
}
