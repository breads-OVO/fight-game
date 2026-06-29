package config

import (
	"fight-game/pkg/common/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	RpcServer zrpc.RpcServerConf

	MySQL   config.MySQLConfig   // MySQL配置
	Redis   redis.RedisConf      // Redis配置
	MongoDB config.MongoDBConfig // MongoDB配置

	Player Player // 服务配置
}

type Player struct {
	Season        string
	RatingDefault int32
}
