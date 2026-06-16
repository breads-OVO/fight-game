package main

import (
	"context"
	"fight-game/pkg/common/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("game")
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
				gm := c.Game
				logger.Info("Game service starting",
					zap.Int("grpc_port", gm.GRPC.Port),
					zap.Int("tick_rate", gm.TickRate),
					zap.Int("match_timeout", gm.MatchTimeout),
					zap.Int("reconnect_timeout", gm.ReconnectTimeout),
				)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// TODO: start gRPC server (GameService)
						// TODO: etcd service registration
						// TODO: GatewayService gRPC client initialization
						// TODO: room manager initialization (sync.Map)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("Game service shutting down")
						// TODO: etcd deregistration (stop accepting new rooms)
						// TODO: mark "shutting down" state
						// TODO: wait for active rooms to finish (shutdown.game-wait seconds)
						// TODO: force terminate remaining rooms
						// TODO: close gRPC server
						// TODO: close database / Redis connections
						return nil
					},
				})
			},
		),
	).Run()
}
