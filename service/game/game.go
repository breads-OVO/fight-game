package main

import (
	"flag"
	"fmt"

	"fight-game/pb/game"
	"fight-game/service/game/internal/config"
	"fight-game/service/game/internal/room"
	"fight-game/service/game/internal/server"
	"fight-game/service/game/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/game/etc/game.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// 启动 WS 直连服务器
	wsServer := room.NewWSServer(c.Game.WsAddr, ctx)
	go func() {
		fmt.Printf("Game WS server listening on %s\n", c.Game.WsAddr)
		if err := wsServer.Start(); err != nil {
			logx.Errorf("WS server error: %v", err)
		}
	}()

	// 启动 gRPC 服务
	srv := server.NewGameServiceServer(ctx)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		game.RegisterGameServiceServer(grpcServer, srv)
		if c.Mode == "dev" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Game gRPC server listening on %s\n", c.ListenOn)
	s.Start()
}
