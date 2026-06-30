package main

import (
	"flag"

	"fight-game/pb/gateway"
	"fight-game/service/gateway/internal/config"
	"fight-game/service/gateway/internal/handler"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/server"
	"fight-game/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/gateway/etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)
	svcCtx.Router = initRouter(svcCtx)

	// 启动 REST (HTTP/WS) 服务
	restServer := rest.MustNewServer(c.RestConf)
	defer restServer.Stop()
	handler.RegisterRoutes(restServer, svcCtx)

	// 启动 PushService gRPC 服务（供 Friend/Mail 等服务调用推送）
	pushSrv := server.NewPushServiceServer(svcCtx)
	grpcServer := zrpc.MustNewServer(c.PushRpcServer, func(grpc *grpc.Server) {
		gateway.RegisterPushServiceServer(grpc, pushSrv)
		if c.Mode == "dev" {
			reflection.Register(grpc)
		}
	})
	defer grpcServer.Stop()

	logx.Infof("Gateway starting — HTTP/WS port: %d, Push gRPC port: %s, REST gRPC port: %d",
		c.WebSocket.Port, c.PushRpcServer.ListenOn, c.Port)

	// 同时启动 REST 和 gRPC 服务
	go func() {
		grpcServer.Start()
	}()
	restServer.Start()
}

func initRouter(svcCtx *svc.ServiceContext) *router.Router {
	r := router.NewRouter()

	// 按模块批量注册，每个模块一个文件、一次注册
	r.RegisterModule(handler.NewSystemHandler().Routes())
	r.RegisterModule(handler.NewAuthHandler(svcCtx).Routes())
	r.RegisterModule(handler.NewMatchHandler(svcCtx).Routes())
	r.RegisterModule(handler.NewPlayerHandler(svcCtx).Routes())
	r.RegisterModule(handler.NewMailHandler(svcCtx).Routes())
	r.RegisterModule(handler.NewFriendHandler(svcCtx).Routes())

	return r
}
