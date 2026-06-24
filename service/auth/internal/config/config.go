package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	JwtSecret          string      // JWT密钥
	JwtExpire          int         // Access Token 过期小时数
	RefreshTokenExpire int         // Refresh Token 过期天数
	MySQL              MySQLConfig // MySQL配置
}

type MySQLConfig struct {
	DataSource   string
	MaxOpenConns int `json:",default=100"` // 最大打开连接数
	MaxIdleConns int `json:",default=10"`  // 最大空闲连接数
}
