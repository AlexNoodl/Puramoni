package db

import (
	"backend/internal/logger"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type MongoClient struct {
	*mongo.Client
}

func ConnectMongo(uri string, log *logger.Logger) (*MongoClient, error) {
	clientOptions := options.Client().ApplyURI(uri).SetConnectTimeout(10 * time.Second)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	log.Info("Connected to MongoDB")
	return &MongoClient{client}, nil
}

func (c *MongoClient) Close() {
	if err := c.Disconnect(context.Background()); err != nil {
		zap.L().Error("Error disconnecting from MongoDB", zap.Error(err))
	}
}
