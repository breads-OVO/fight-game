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
	MongoDB config.MongoDBConfig // MongoDB配置（聊天记录）

	GatewayRpc zrpc.RpcClientConf // Gateway 推送服务客户端配置
	PlayerRpc  zrpc.RpcClientConf // Player 服务客户端配置
}
