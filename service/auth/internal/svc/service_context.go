package svc

import (
	commonConf "fight-game/pkg/common/config"
	"fight-game/service/auth/internal/config"
	"fight-game/service/auth/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	//初始化db
	db := commonConf.InitMySQL(&c.MySQL)

	// 自动迁移
	if err := db.AutoMigrate(
		&model.User{},
	); err != nil {
		logx.Must(err)
	}

	logx.Info("MySQL connected and migrated successfully")
	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
