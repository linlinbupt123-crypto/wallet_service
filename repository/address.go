package repository

import (
	"context"

	"github.com/linlinbupt123-crypto/wallet_service/db"
	"github.com/linlinbupt123-crypto/wallet_service/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AddressRepo struct {
	col *mongo.Collection
}

func NewAddressRepo() *AddressRepo {
	return &AddressRepo{col: db.MongoDB.AssetColl}
}

func (r *AddressRepo) Create(ctx context.Context, addr *entity.Address) error {
	_, err := r.col.InsertOne(ctx, addr)
	return err
}

func (r *AddressRepo) GetByUserID(ctx context.Context, userID string) ([]*entity.Address, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*entity.Address
	for cur.Next(ctx) {
		var a entity.Address
		if err := cur.Decode(&a); err == nil {
			out = append(out, &a)
		}
	}
	return out, nil
}

// 获取用户在某条链的最大 index (用于生成下一地址)
func (r *AddressRepo) GetMaxIndex(ctx context.Context, walletID string, chain string) (int, error) {
	opts := options.FindOne().SetSort(bson.M{"index": -1})

	var out entity.Address
	err := r.col.FindOne(ctx, bson.M{
		"wallet_id": walletID,
		"chain":     chain,
	}, opts).Decode(&out)

	if err == mongo.ErrNoDocuments {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}

	return int(out.Index), nil
}

// GetByAddrID 根据链上的地址查找 Address
func (r *AddressRepo) GetByAddrID(ctx context.Context, address string) (*entity.Address, error) {
	var addr entity.Address
	err := r.col.FindOne(ctx, bson.M{"address": address}).Decode(&addr)
	if err == mongo.ErrNoDocuments {
		return nil, nil // 找不到返回 nil
	}
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// GetByWalletID 根据钱包 ID 查找 Address
func (r *AddressRepo) GetByWalletID(ctx context.Context, walletID string) (*entity.Address, error) {
	var addr entity.Address
	err := r.col.FindOne(ctx, bson.M{"wallet_id": walletID}).Decode(&addr)
	if err == mongo.ErrNoDocuments {
		return nil, nil // 找不到返回 nil
	}
	if err != nil {
		return nil, err
	}
	return &addr, nil
}
