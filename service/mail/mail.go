package main

import (
	"flag"

	"fight-game/pb/mail"
	"fight-game/service/mail/internal/config"
	"fight-game/service/mail/internal/server"
	"fight-game/service/mail/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/mail/etc/mail.yaml", "the config file")

func main() {
	// 加载配置
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	srv := server.NewMailServiceServer(ctx)

	s := zrpc.MustNewServer(c.RpcServer, func(grpcServer *grpc.Server) {
		mail.RegisterMailServiceServer(grpcServer, srv)
		if c.RpcServer.Mode == "dev" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Infof("Mail service starting on %s", c.RpcServer.ListenOn)
	s.Start()
}
