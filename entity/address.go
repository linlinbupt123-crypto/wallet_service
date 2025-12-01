package entity

import (
	"time"
)

type Address struct {
    UserID      string
    Chain       string // "BTC" | "ETH" | "SOL"
    Address     string
    DerivePath  string // m/44'/60'/0'/0/0
    CreatedAt   time.Time
}
