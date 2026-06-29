package main

import (
	"flag"

	"fight-game/pb/player"
	"fight-game/service/player/internal/config"
	"fight-game/service/player/internal/server"
	"fight-game/service/player/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/player/etc/player.yaml", "the config file")

func main() {
	// 加载配置
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	srv := server.NewPlayerServiceServer(ctx)

	s := zrpc.MustNewServer(c.RpcServer, func(grpcServer *grpc.Server) {
		player.RegisterPlayerServiceServer(grpcServer, srv)
		if c.RpcServer.Mode == "dev" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Infof("Player service starting on %s", c.RpcServer.ListenOn)
	s.Start()
}
