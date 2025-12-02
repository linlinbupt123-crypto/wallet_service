package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	client     *mongo.Client
	db         *mongo.Database
	walletColl *mongo.Collection
	assetColl  *mongo.Collection
	subColl    *mongo.Collection
}

func NewMongoRepo(ctx context.Context, uri, dbName string) (*MongoRepo, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	// ping
	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx2, nil); err != nil {
		return nil, err
	}
	db := client.Database(dbName)
	return &MongoRepo{
		client:     client,
		db:         db,
		walletColl: db.Collection("wallets"),
		assetColl:  db.Collection("assets"),
		subColl:    db.Collection("subscriptions"),
	}, nil
}
