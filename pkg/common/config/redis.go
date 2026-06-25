package config

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Host string // Redis 地址
	Type string `json:",default=node"` // Redis 类型
	Pass string // Redis 密码
	DB   int    // Redis 数据库
}

// InitRedis 初始化 Redis
func InitRedis(c *RedisConfig) redis.UniversalClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Host,
		Password: c.Pass,
		DB:       c.DB,
	})

	logx.Info("Redis connected successfully")
	return redisClient
}
