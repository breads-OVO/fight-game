package config

// --- Gateway 配置 ---

type GatewayConfig struct {
	WS   GatewayWSConfig   `mapstructure:"ws"`   // websocket
	GRPC GatewayGRPCConfig `mapstructure:"grpc"` // gRPC
	HTTP GatewayHTTPConfig `mapstructure:"http"` // HTTP
}

type GatewayWSConfig struct {
	Port         int `mapstructure:"port"`             // websocket 端口
	ReadTimeout  int `mapstructure:"read-timeout"`     // websocket 读超时
	WriteTimeout int `mapstructure:"write-timeout"`    // websocket 写超时
	PingInterval int `mapstructure:"ping-interval"`    // websocket ping 间隔
	MaxMsgSize   int `mapstructure:"max-message-size"` // websocket 最大消息大小
}

type GatewayGRPCConfig struct {
	Port int `mapstructure:"port"` // gRPC 端口
}

type GatewayHTTPConfig struct {
	Port int `mapstructure:"port"` // HTTP 端口
}
