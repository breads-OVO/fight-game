package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf                    // REST (HTTP/WS) 配置
	PushRpcServer zrpc.RpcServerConf // Push gRPC 服务端配置

	AuthRpc   zrpc.RpcClientConf // Auth gRPC 客户端配置
	MatchRpc  zrpc.RpcClientConf // Match gRPC 客户端配置
	PlayerRpc zrpc.RpcClientConf // Player gRPC 客户端配置
	MailRpc   zrpc.RpcClientConf // Mail gRPC 客户端配置
	FriendRpc zrpc.RpcClientConf // Friend gRPC 客户端配置
	WebSocket WebSocket          // websocket配置
}

type WebSocket struct {
	Port           int   // WebSocket服务端口
	ReadTimeout    int   // 读取超时时间
	WriteTimeout   int   // 写入超时时间
	PingInterval   int   // 心跳间隔
	MaxMessageSize int64 // 最大消息大小
	IdleTimeout    int   // 空闲超时时间
}
