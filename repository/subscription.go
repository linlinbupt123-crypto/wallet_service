package repository

import (
	"context"

	"github.com/linlinbupt123-crypto/wallet_service/db"
	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Subscription struct {
	col *mongo.Collection
}

func NewSubscriptionRepo() *Subscription {
	return &Subscription{col: db.MongoDB.SubColl}
}

func (r *Subscription) Create(ctx context.Context, s *entity.Subscription) error {
	_, err := r.col.InsertOne(ctx, s)
	return err
}

func (r *Subscription) List(ctx context.Context, userID string) ([]*entity.Subscription, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID})
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
