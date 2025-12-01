package entity

import "time"

// Subscription 表示用户对某个币种价格的订阅
type Subscription struct {
	UserID    string    `bson:"userID"`    // 用户ID
	Symbol    string    `bson:"symbol"`    // 币种符号，例如 BTC、ETH
	Threshold float64   `bson:"threshold"` // 价格阈值，超过时触发通知
	Direction string    `bson:"direction"` // 可选: "above" / "below"，表示价格高于还是低于阈值
	CreatedAt time.Time `bson:"createdAt"` // 订阅创建时间
	NotifyURL string    `bson:"notifyURL"` // 可选: 回调通知 URL
	Active    bool      `bson:"active"`    // 订阅是否有效
}