package main

import (
	"flag"

	"fight-game/pb/friend"
	"fight-game/service/friend/internal/config"
	"fight-game/service/friend/internal/server"
	"fight-game/service/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/friend/etc/friend.yaml", "the config file")

func main() {
	// 加载配置
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	srv := server.NewFriendServiceServer(ctx)

	s := zrpc.MustNewServer(c.RpcServer, func(grpcServer *grpc.Server) {
		friend.RegisterFriendServiceServer(grpcServer, srv)
		if c.RpcServer.Mode == "dev" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Infof("Friend service starting on %s", c.RpcServer.ListenOn)
	s.Start()
}
