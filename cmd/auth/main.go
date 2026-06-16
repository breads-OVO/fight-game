package main

import (
	"context"
	"fight-game/pkg/common/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("auth")
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
				logger.Info("Auth service starting",
					zap.Int("grpc_port", c.Auth.GRPC.Port),
				)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// TODO: start gRPC server (AuthService)
						// TODO: etcd service registration
						// TODO: GORM database initialization (accounts table)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("Auth service shutting down")
						// TODO: close gRPC server
						// TODO: close database connection
						// TODO: etcd deregistration
						return nil
					},
				})
			},
		),
	).Run()
}
