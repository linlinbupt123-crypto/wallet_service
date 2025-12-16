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
func (r *Wallet) Create(ctx context.Context, w *entity.HDWallet) (string, error) {
	res, err := r.col.InsertOne(ctx, w)
	if err != nil {
		return "", err
	}
	objectID := res.InsertedID.(primitive.ObjectID)
	return objectID.Hex(), nil
}

func (r *Wallet) GetByUserID(ctx context.Context, userID string) (*entity.HDWallet, error) {
	var w entity.HDWallet
	err := r.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w, nil
}
