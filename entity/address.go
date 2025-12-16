package entity

import (
	"time"
)

type Address struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	WalletID  string    `bson:"wallet_id" json:"wallet_id"`
	Chain     string    `bson:"chain" json:"chain"`     // btc / eth / solana
	Address   string    `bson:"address" json:"address"` // 主地址
	Index     uint32    `bson:"index" json:"index"`     // 派生索引
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
