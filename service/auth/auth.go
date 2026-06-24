package main

import (
	"flag"

	"fight-game/pb/auth"
	"fight-game/service/auth/internal/config"
	"fight-game/service/auth/internal/server"
	"fight-game/service/auth/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/auth/etc/auth.yaml", "the config file")

func main() {

	// 加载配置
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务
	ctx := svc.NewServiceContext(c)
	srv := server.NewAuthServiceServer(ctx)

	// 启动 gRPC 服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		auth.RegisterAuthServiceServer(grpcServer, srv)
		if c.Mode == "dev" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Infof("Auth service starting on %s", c.ListenOn)
	s.Start()
}
