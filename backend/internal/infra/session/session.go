// Package session 基于 Redis 的登录会话（token）存储，供鉴权中间件与登录/登出使用。
// 它是 Redis 的一层业务包装：只暴露 token 的存/取/删，调用方不碰 key 命名。
package session

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Session 登录态存储，底层是 Redis。
type Session struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *Session {
	return &Session{rdb: rdb}
}

func tokenKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

// Save 保存用户 token（带过期）。
func (s *Session) Save(ctx context.Context, userID int64, token string, ttl time.Duration) error {
	return s.rdb.Set(ctx, tokenKey(userID), token, ttl).Err()
}

// Get 取用户当前有效 token。
func (s *Session) Get(ctx context.Context, userID int64) (string, error) {
	return s.rdb.Get(ctx, tokenKey(userID)).Result()
}

// Delete 删除用户 token（登出）。
func (s *Session) Delete(ctx context.Context, userID int64) error {
	return s.rdb.Del(ctx, tokenKey(userID)).Err()
}
