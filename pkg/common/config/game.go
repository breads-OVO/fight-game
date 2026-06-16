package config

// --- Game 配置 ---

type GameConfig struct {
	GRPC             GameGRPCConfig `mapstructure:"grpc"`
	TickRate         int            `mapstructure:"tick-rate"`
	MatchTimeout     int            `mapstructure:"match-timeout"`
	ReconnectTimeout int            `mapstructure:"reconnect-timeout"`
}

type GameGRPCConfig struct {
	Port int `mapstructure:"port"`
}
