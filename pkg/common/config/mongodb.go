package config

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBConfig struct {
	Uri      string
	Database string
}

// InitMongoDB 初始化MongoDB
func InitMongoDB(c *MongoDBConfig) *mongo.Database {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(c.Uri))
	if err != nil {
		logx.Must(err)
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		logx.Must(err)
	}
	mongoDB := mongoClient.Database(c.Database)

	logx.Info("MongoDB connected successfully")
	return mongoDB
}
