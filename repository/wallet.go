package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
)


type WalletRepo struct {
col *mongo.Collection
sub *mongo.Collection
}


func NewWalletRepo(db *mongo.Database) *WalletRepo {
return &WalletRepo{
col: db.Collection("wallets"),
sub: db.Collection("subscriptions"),
}
}


func (r *WalletRepo) Save(ctx context.Context, w *entity.HDWallet) error {
_, err := r.col.InsertOne(ctx, w)
return err
}


func (r *WalletRepo) GetByUserID(ctx context.Context, userID string) (*entity.HDWallet, error) {
var w entity.HDWallet
if err := r.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&w); err != nil {
return nil, err
}
return &w, nil
}


func (r *WalletRepo) AddSubscription(ctx context.Context, s *entity.Subscription) error {
_, err := r.sub.InsertOne(ctx, s)
return err
}


func (r *WalletRepo) ListSubscriptions(ctx context.Context) ([]*entity.Subscription, error) {
cur, err := r.sub.Find(ctx, bson.M{})
if err != nil {
return nil, err
}
defer cur.Close(ctx)
var out []*entity.Subscription
for cur.Next(ctx) {
var s entity.Subscription
if err := cur.Decode(&s); err == nil {
out = append(out, &s)
}
}
return out, nil
}


func (r *WalletRepo) AddAddress(ctx context.Context, addr *entity.Address) error {
_, err := r.col.UpdateOne(ctx, bson.M{"user_id": addr.UserID}, bson.M{"$push": bson.M{"addresses": addr}}, options.Update().SetUpsert(true))
return err
}