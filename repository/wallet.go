/*
user_id → 唯一索引
XPub / EncryptedSeed / XPrvEncrypted / SaltHex / CreatedAt
*/
package repository

import (
	"context"

	"github.com/linlinbupt123-crypto/wallet_service/db"
	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Wallet struct {
	col *mongo.Collection
}

func NewWalletRepo() *Wallet {
	return &Wallet{col: db.MongoDB.WalletColl}
}

// Create wallet
func (r *Wallet) Create(ctx context.Context, w *entity.Wallet) (string, error) {
	res, err := r.col.InsertOne(ctx, w)
	if err != nil {
		return "", err
	}
	objectID := res.InsertedID.(primitive.ObjectID)
	return objectID.Hex(), nil
}

func (r *Wallet) GetByUserID(ctx context.Context, userID string) ([]*entity.Wallet, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var wallets []*entity.Wallet
	for cur.Next(ctx) {
		var w entity.Wallet
		if err := cur.Decode(&w); err != nil {
			continue // decode 错误跳过
		}
		wallets = append(wallets, &w)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// GetByID 根据 wallet ID 查找 Wallet
func (r *Wallet) GetByID(ctx context.Context, walletID string) (*entity.Wallet, error) {
	var w entity.Wallet
	oid, err := primitive.ObjectIDFromHex(walletID)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&w)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}
