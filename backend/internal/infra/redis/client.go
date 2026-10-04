package redis

import (
	"context"
	"log"

	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// NewRedisClient 按配置创建 go-redis 客户端并 Ping，返回客户端和关闭函数。
// 直接返回原生 *goredis.Client，不维护全局单例；由 DI 在组合根处创建并注入。
func NewRedisClient() (*goredis.Client, func()) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Fatal("failed to connect redis", err)
	}
	return client, func() { _ = client.Close() }
}
