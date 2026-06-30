package svc

import (
	"context"
	"time"

	commonConf "fight-game/pkg/common/config"
	"fight-game/service/friend/internal/config"
	"fight-game/service/friend/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"

	"fight-game/pb/gateway"
	"fight-game/pb/player"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	RedisClient  *redis.Redis
	MongoDB      *mongo.Database
	PushClient   gateway.PushServiceClient
	PlayerClient player.PlayerServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL
	db := commonConf.InitMySQL(&c.MySQL)

	// 自动迁移
	if err := db.AutoMigrate(
		&model.Friend{},
	); err != nil {
		logx.Must(err)
	}

	// 初始化 Redis
	rds, err := redis.NewRedis(c.Redis)
	if err != nil {
		logx.Must(err)
	}

	// 初始化 MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(c.MongoDB.Uri))
	if err != nil {
		logx.Must(err)
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		logx.Must(err)
	}
	mongoDB := mongoClient.Database(c.MongoDB.Database)

	// 创建聊天记录索引
	if err := ensureIndexes(mongoDB); err != nil {
		logx.Must(err)
	}

	// 初始化 Gateway Push 客户端
	pushRpc := zrpc.MustNewClient(c.GatewayRpc)
	pushClient := gateway.NewPushServiceClient(pushRpc.Conn())

	// 初始化 Player 客户端
	playerRpc := zrpc.MustNewClient(c.PlayerRpc)
	playerClient := player.NewPlayerServiceClient(playerRpc.Conn())

	logx.Info("Friend service MySQL, Redis and MongoDB initialized successfully")
	return &ServiceContext{
		Config:       c,
		DB:           db,
		RedisClient:  rds,
		MongoDB:      mongoDB,
		PushClient:   pushClient,
		PlayerClient: playerClient,
	}
}

func ensureIndexes(db *mongo.Database) error {
	ctx := context.Background()

	coll := db.Collection(model.CollectionChatMessages)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// 私聊查询索引（合并两个方向查询）
			Keys: bson.D{
				{Key: "chat_type", Value: 1},
				{Key: "participants", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			// TTL 过期索引：30天后自动删除
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 3600),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
