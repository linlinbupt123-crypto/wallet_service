package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// MongoDB connection config
	uri := "mongodb://admin:password@localhost:27017/?authSource=admin"
	dbName := "wallet_service"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("MongoDB connect error:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("MongoDB disconnect error: %v", err)
		}
	}()

	db := client.Database(dbName)

	// 初始化所有 collection
	if err := initIndexes(ctx, db); err != nil {
		log.Fatal("Init indexes failed:", err)
	}

	fmt.Println("All indexes initialized successfully.")
}

// 安全创建索引函数
func createIndexSafe(ctx context.Context, col *mongo.Collection, index mongo.IndexModel) error {
	_, err := col.Indexes().CreateOne(ctx, index)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return nil // 忽略已存在索引
		}
		return err
	}
	return nil
}

// 初始化所有 collection 索引
func initIndexes(ctx context.Context, db *mongo.Database) error {
	// addresses
	addrCol := db.Collection("addresses")
	addrIndexes := []mongo.IndexModel{
		{Keys: bson.M{"address": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"user_id": 1}},
		{Keys: bson.M{"wallet_id": 1}},
		{Keys: bson.M{"chain": 1}},
		{Keys: bson.D{{Key: "wallet_id", Value: 1}, {Key: "chain", Value: 1}, {Key: "index", Value: -1}}},
	}
	for _, idx := range addrIndexes {
		if err := createIndexSafe(ctx, addrCol, idx); err != nil {
			return fmt.Errorf("addresses index error: %w", err)
		}
	}

	// subscriptions
	subCol := db.Collection("subscriptions")
	subIndexes := []mongo.IndexModel{
		{Keys: bson.M{"user_id": 1}},
		// 复合索引改为 bson.D
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "chain", Value: 1}, {Key: "symbol", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
	for _, idx := range subIndexes {
		if err := createIndexSafe(ctx, subCol, idx); err != nil {
			return fmt.Errorf("subscriptions index error: %w", err)
		}
	}

	// wallets
	walletCol := db.Collection("wallets")
	walletIndexes := []mongo.IndexModel{
		{Keys: bson.M{"user_id": 1}, Options: options.Index().SetUnique(true)},
	}
	for _, idx := range walletIndexes {
		if err := createIndexSafe(ctx, walletCol, idx); err != nil {
			return fmt.Errorf("wallets index error: %w", err)
		}
	}

	return nil
}
