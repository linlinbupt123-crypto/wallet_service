package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *MongoRepo

const (
	uri    = "mongodb://admin:password@localhost:27017/?authSource=admin"
	dbName = "wallet_service"
)

func InitMongo() {}
func init() {
	ctx := context.Background()
	var err error
	MongoDB, err = NewMongoRepo(ctx, uri, dbName)
	if err != nil {
		panic(err)
	}
}

type MongoRepo struct {
	Client     *mongo.Client
	DB         *mongo.Database
	WalletColl *mongo.Collection
	AssetColl  *mongo.Collection
	SubColl    *mongo.Collection
	AddrColl   *mongo.Collection
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
		Client:     client,
		DB:         db,
		WalletColl: db.Collection("wallets"),
		AssetColl:  db.Collection("assets"),
		SubColl:    db.Collection("subscriptions"),
		AddrColl:   db.Collection("addresses"),
	}, nil
}
