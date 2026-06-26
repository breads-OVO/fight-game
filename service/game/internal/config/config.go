package config

import (
	commonConf "fight-game/pkg/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	MySQL      commonConf.MySQLConfig // MySQL配置
	RedisCache commonConf.RedisConfig // Redis配置
	Game       GameConfig             // 游戏配置
}

type GameConfig struct {
	WsAddr       string `json:",default=:9005"` // WebSocket 直连地址
	PickTimeout  int    `json:",default=60"`    // 选人超时（秒）
	BanTimeout   int    `json:",default=30"`    // 禁用超时（秒）
	FightTimeout int    `json:",default=60"`    // 每回合战斗超时（秒）
	Characters   int    `json:",default=3"`     // 每人可选角色数
}
