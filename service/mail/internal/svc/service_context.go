package svc

import (
	"context"
	commonConf "fight-game/pkg/common/config"
	"fight-game/service/mail/internal/config"
	"fight-game/service/mail/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	RedisClient *redis.Redis
	MongoDB     *mongo.Database
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL
	db := commonConf.InitMySQL(&c.MySQL)

	// 初始化 Redis
	rds, err := redis.NewRedis(c.Redis)
	if err != nil {
		logx.Must(err)
	}

	// 初始化 MongoDB
	mongoDB := commonConf.InitMongoDB(&c.MongoDB)
	// 创建索引
	if err := ensureIndexes(mongoDB); err != nil {
		logx.Must(err)
	}

	logx.Info("Mail service MySQL, Redis and MongoDB initialized successfully")
	return &ServiceContext{
		Config:      c,
		DB:          db,
		RedisClient: rds,
		MongoDB:     mongoDB,
	}
}

// 创建索引
func ensureIndexes(db *mongo.Database) error {
	ctx := context.Background()

	// MailBox 索引
	mailBoxColl := db.Collection(model.CollectionMailBox)
	_, err := mailBoxColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "player_id", Value: 1}, {Key: "mail_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "player_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "player_id", Value: 1}, {Key: "status", Value: 1}},
		},
	})
	if err != nil {
		return err
	}

	// MailBody 索引
	mailBodyColl := db.Collection(model.CollectionMailBody)
	_, err = mailBodyColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "mail_id", Value: 1}},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
