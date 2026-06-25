package main

import (
	"flag"

	"fight-game/pb/match"
	"fight-game/service/match/internal/config"
	"fight-game/service/match/internal/server"
	"fight-game/service/match/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "service/match/etc/match.yaml", "the config file")

func main() {
	// 加载配置文件
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// 启动两个匹配扫描器（娱乐 + 竞技）
	ctx.EntertainmentScanner.Start()
	ctx.CompetitionScanner.Start()
	defer ctx.EntertainmentScanner.Stop()
	defer ctx.CompetitionScanner.Stop()

	// 启动RPC服务
	srv := server.NewMatchServiceServer(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		match.RegisterMatchServiceServer(grpcServer, srv)
		if c.Mode == "dev" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Infof("Match service starting on %s", c.ListenOn)
	s.Start()
}
