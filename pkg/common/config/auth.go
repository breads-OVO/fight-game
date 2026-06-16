package config

// --- Auth 配置 ---

type AuthConfig struct {
	GRPC AuthGRPCConfig `mapstructure:"grpc"` // gRPC
}

type AuthGRPCConfig struct {
	Port int `mapstructure:"port"` // gRPC 端口
}
