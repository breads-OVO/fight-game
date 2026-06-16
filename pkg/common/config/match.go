package config

// --- Match 配置 ---

type MatchConfig struct {
	GRPC         MatchGRPCConfig `mapstructure:"grpc"`          // gRPC
	ELORange     int             `mapstructure:"elo-range"`     // ELO 范围
	ScanInterval int             `mapstructure:"scan-interval"` // 扫描间隔
	Timeout      int             `mapstructure:"timeout"`       // 超时时间
}

type MatchGRPCConfig struct {
	Port int `mapstructure:"port"` // gRPC 端口
}
