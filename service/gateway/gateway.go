package main

import (
	"fight-game/service/gateway/internal/config"
	"fight-game/service/gateway/internal/handler"
	"fight-game/service/gateway/internal/router"
	"fight-game/service/gateway/internal/svc"
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "service/gateway/etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)
	svcCtx.Router = initRouter(svcCtx)

	handler.RegisterRoutes(server, svcCtx)

	logx.Infof("Gateway starting on WebSocket port %d, gRPC port %d", c.WebSocket.Port, c.Port)
	server.Start()
}

func initRouter(svcCtx *svc.ServiceContext) *router.Router {
	r := router.NewRouter()

	// 按模块批量注册，每个模块一个文件、一次注册
	r.RegisterModule(handler.NewSystemHandler().Routes())
	r.RegisterModule(handler.NewAuthHandler(svcCtx).Routes())

	return r
}
