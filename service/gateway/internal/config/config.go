package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf           // rest配置
	WebSocket     WebSocket // websocket配置
	Auth          Auth      // 认证配置
}

type WebSocket struct {
	Port           int   // WebSocket服务端口
	ReadTimeout    int   // 读取超时时间
	WriteTimeout   int   // 写入超时时间
	PingInterval   int   // 心跳间隔
	MaxMessageSize int64 // 最大消息大小
	IdleTimeout    int   // 空闲超时时间
}

type Auth struct {
	JwtSecret string // JWT密钥
	JwtExpire int    // JWT过期时间
}
