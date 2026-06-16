package config

//通用配置

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"` // etcd 集群地址
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"` // redis 地址
	DB   int    `mapstructure:"db"`   // redis 数据库
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"` // mysql 数据源名称
}

type ShutdownConfig struct {
	GameWait    int `mapstructure:"game-wait"`    // 游戏服务关闭时等待的时长
	GatewayWait int `mapstructure:"gateway-wait"` // 网关服务关闭时等待的时长
	GRPCGrace   int `mapstructure:"grpc-grace"`   // gRPC 服务关闭时等待的时长
}
