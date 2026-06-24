package svc

import (
	"fight-game/service/auth/internal/config"
	"fight-game/service/auth/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	//初始化db
	db, err := gorm.Open(mysql.Open(c.MySQL.DataSource), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		logx.Must(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logx.Must(err)
	}
	sqlDB.SetMaxOpenConns(c.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MySQL.MaxIdleConns)

	if err := db.AutoMigrate(&model.User{}); err != nil {
		logx.Must(err)
	}

	logx.Info("MySQL connected and migrated successfully")
	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
