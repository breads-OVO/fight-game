package config

import (
	"fight-game/pkg/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	JwtSecret          string             // JWT密钥
	JwtExpire          int                // Access Token 过期小时数
	RefreshTokenExpire int                // Refresh Token 过期天数
	MySQL              config.MySQLConfig // MySQL配置
}
