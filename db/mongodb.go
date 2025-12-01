package db

import (
	"context"

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

func NewMongoRepo(uri, dbName string) (*MongoRepo, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
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


