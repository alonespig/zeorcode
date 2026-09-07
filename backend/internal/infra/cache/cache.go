package cache

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Cache struct {
	client *goredis.Client
}

func NewCache(client *goredis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	if c == nil || c.client == nil {
		return "", nil
	}
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	if c == nil || c.client == nil {
		return true, nil
	}
	return c.client.SetNX(ctx, key, value, ttl).Result()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil || len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	if c == nil || c.client == nil {
		return 0, nil
	}
	return c.client.Incr(ctx, key).Result()
}

func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *Cache) SAdd(ctx context.Context, key string, members ...any) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.SAdd(ctx, key, members...).Err()
}

func (c *Cache) SCard(ctx context.Context, key string) (int64, error) {
	if c == nil || c.client == nil {
		return 0, nil
	}
	return c.client.SCard(ctx, key).Result()
}

func (c *Cache) SIsMember(ctx context.Context, key string, member any) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}
	return c.client.SIsMember(ctx, key, member).Result()
}

func (c *Cache) HSet(ctx context.Context, key string, values ...any) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.HSet(ctx, key, values...).Err()
}

func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	return c.client.HGetAll(ctx, key).Result()
}

func (c *Cache) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	if c == nil || c.client == nil {
		return 0, nil
	}
	return c.client.HIncrBy(ctx, key, field, incr).Result()
}

func (c *Cache) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	value, err := c.GetString(ctx, key)
	if err != nil || value == "" {
		return false, err
	}
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, ttl)
}

// 固定窗口限流脚本：原子地 INCR，首次计数（==1）时给 key 设过期，避免
// INCR 与 EXPIRE 分开时中途失败导致 key 永不过期。
var rateLimitScript = goredis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return n
`)

// getAndDeleteScript 原子读取并删除一次性数据，兼容尚不支持 GETDEL 的 Redis 版本。
var getAndDeleteScript = goredis.NewScript(`
local value = redis.call('GET', KEYS[1])
if value then
  redis.call('DEL', KEYS[1])
end
return value
`)

var compareAndDeleteScript = goredis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// GetAndDelete 原子消费字符串值；key 不存在时返回 redis.Nil。
func (c *Cache) GetAndDelete(ctx context.Context, key string) (string, error) {
	if c == nil || c.client == nil {
		return "", goredis.Nil
	}
	return getAndDeleteScript.Run(ctx, c.client, []string{key}).Text()
}

// CompareAndDelete only releases a lock owned by value. A timed-out holder
// therefore cannot delete a newer holder's lock.
func (c *Cache) CompareAndDelete(ctx context.Context, key, value string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return compareAndDeleteScript.Run(ctx, c.client, []string{key}, value).Err()
}

// RateLimitIncr 固定窗口计数：返回该 key 在当前窗口内的累计次数。
// cache 不可用时返回 (0, nil)，交由调用方 fail-open 放行（限流组件挂掉不应拖垮接口）。
func (c *Cache) RateLimitIncr(ctx context.Context, key string, window time.Duration) (int64, error) {
	if c == nil || c.client == nil {
		return 0, nil
	}
	return rateLimitScript.Run(ctx, c.client, []string{key}, int(window.Seconds())).Int64()
}
