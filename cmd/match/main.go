package main

import (
	"context"
	"fight-game/pkg/common/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("match")
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
				mt := c.Match
				logger.Info("Match service starting",
					zap.Int("grpc_port", mt.GRPC.Port),
					zap.Int("elo_range", mt.ELORange),
					zap.Int("scan_interval", mt.ScanInterval),
				)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// TODO: start gRPC server (MatchService)
						// TODO: etcd service registration
						// TODO: Redis client initialization (match queue)
						// TODO: start match scanning goroutine
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("Match service shutting down")
						// TODO: stop match scanning goroutine
						// TODO: close gRPC server
						// TODO: close Redis connection
						// TODO: etcd deregistration
						return nil
					},
				})
			},
		),
	).Run()
}
