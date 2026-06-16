package main

import (
	"context"
	"fight-game/pkg/common/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("gateway")
	if err != nil {
		panic(err)
	}

	fx.New(
		fx.Provide(
			zap.NewProduction,
			func() *config.Config { return cfg },
		),
		fx.Invoke(
			func(lc fx.Lifecycle, logger *zap.Logger, c *config.Config) {
				gw := c.Gateway
				logger.Info("Gateway service starting",
					zap.Int("ws_port", gw.WS.Port),
					zap.Int("grpc_port", gw.GRPC.Port),
					zap.Int("http_port", gw.HTTP.Port),
				)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// TODO: start WebSocket server (gorilla/websocket)
						// TODO: start gRPC server (GatewayService)
						// TODO: start HTTP health check
						// TODO: etcd service registration
						// TODO: Auth gRPC client initialization
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("Gateway service shutting down")
						// TODO: stop accepting new WS connections
						// TODO: push SERVER_SHUTDOWN to active sessions
						// TODO: close WebSocket listener
						// TODO: close gRPC server
						// TODO: etcd deregistration
						return nil
					},
				})
			},
		),
	).Run()
}
