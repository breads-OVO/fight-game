package config

import (
	"fight-game/pkg/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	MySQL      config.MySQLConfig // MySQL配置
	RedisCache config.RedisConfig // Redis配置
	Match      MatchConfig        // 匹配配置
}

type MatchConfig struct {
	QueueTimeout    int `json:",default=30"`  // 匹配超时时间（秒）
	RatingRange     int `json:",default=100"` // 初始匹配分差
	RatingRangeMax  int `json:",default=500"` // 匹配范围最大值
	RatingRangeStep int `json:",default=50"`  // 匹配分差步进（随时间扩大）
}
