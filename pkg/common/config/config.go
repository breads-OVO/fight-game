package config

import (
	"fmt"

	"github.com/spf13/viper"
)

/*
Config 统一配置根结构
mapstructure 标签用于指定结构体字段在配置文件中的键名
*/
type Config struct {
	Service  string         `mapstructure:"service"`  // 服务名称
	Etcd     EtcdConfig     `mapstructure:"etcd"`     // etcd 配置
	Redis    RedisConfig    `mapstructure:"redis"`    // redis 配置
	MySQL    MySQLConfig    `mapstructure:"mysql"`    // mysql 配置
	Shutdown ShutdownConfig `mapstructure:"shutdown"` // 关闭服务时等待的时长

	// 以下是各服务的私有配置，根据 service 字段选择加载
	Gateway *GatewayConfig `mapstructure:"gateway"` // 网关服务配置
	Auth    *AuthConfig    `mapstructure:"auth"`    // 认证服务配置
	Match   *MatchConfig   `mapstructure:"match"`   // 匹配服务配置
	Game    *GameConfig    `mapstructure:"game"`    // 游戏服务配置
}

// Load 加载指定服务的配置文件。
// 加载顺序：default.yaml → {service}.yaml → 环境变量覆盖
func Load(service string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("default")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read default config: %w", err)
	}

	// 加载服务特定配置，合并到默认配置之上
	v.SetConfigName(service)
	if err := v.MergeInConfig(); err != nil {
		// 服务配置文件不存在时不报错，仅使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("merge %s config: %w", service, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.Service = service
	return &cfg, nil
}
