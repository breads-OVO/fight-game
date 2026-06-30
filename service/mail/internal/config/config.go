package config

import (
	"fight-game/pkg/common/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	RpcServer zrpc.RpcServerConf

	MySQL     config.MySQLConfig   // MySQL配置
	Redis     redis.RedisConf      // Redis配置
	MongoDB   config.MongoDBConfig // MongoDB配置
	PlayerRpc zrpc.RpcClientConf   // Player gRPC 客户端配置

	Mail Mail // 邮件服务配置
}

type Mail struct {
	MailExpire int64 `json:",default=2592000"` // 邮件默认过期时间（秒），默认30天
}
