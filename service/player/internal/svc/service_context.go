package svc

import (
	commonConf "fight-game/pkg/common/config"
	"fight-game/service/player/internal/config"
	"fight-game/service/player/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	RedisClient *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL
	db := commonConf.InitMySQL(&c.MySQL)

	// 自动迁移
	if err := db.AutoMigrate(
		&model.Player{},
		&model.PlayerCurrency{},
		&model.PlayerAsset{},
		&model.PlayerRank{},
	); err != nil {
		logx.Must(err)
	}

	// 初始化 Redis
	rds, err := redis.NewRedis(c.Redis)
	if err != nil {
		logx.Must(err)
	}

	logx.Info("Player service MySQL and Redis initialized successfully")
	return &ServiceContext{
		Config:      c,
		DB:          db,
		RedisClient: rds,
	}
}
