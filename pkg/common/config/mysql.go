package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLConfig MySQL配置
type MySQLConfig struct {
	DataSource   string
	MaxOpenConns int `json:",default=100"` // 最大打开连接数
	MaxIdleConns int `json:",default=10"`  // 最大空闲连接数
}

// InitMySQL 初始化MySQL
func InitMySQL(c *MySQLConfig) *gorm.DB {
	// 初始化 MySQL
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{
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
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	logx.Info("MySQL connected successfully")
	return db
}
