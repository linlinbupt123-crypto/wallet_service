package domain

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
)

type Wallet struct {
	Ctx              context.Context
	WalletRepo       *repository.Wallet
	SubscriptionRepo *repository.Subscription
	HDWalletDomain   *HDWallet
}

// NewWalletService 整合 HDWalletService
func NewWalletService() *Wallet {
	hd := NewHDWallet()
	return &Wallet{
		WalletRepo:       repository.NewWalletRepo(),
		HDWalletDomain:   hd,
		SubscriptionRepo: repository.NewSubscriptionRepo(),
	}
}

// 创建钱包
func (s *Wallet) CreateWallet(userID, passphrase string) (*entity.Wallet, error) {
	return s.HDWalletDomain.CreateWallet(s.Ctx, userID, passphrase)
}

// 价格订阅
func (s *Wallet) SubscribePrice(userID, symbol string, threshold float64) error {
	sub := &entity.Subscription{UserID: userID, Symbol: symbol, Threshold: threshold}
	return s.SubscriptionRepo.Create(s.Ctx, sub)
}

// 异步价格监控
func (s *Wallet) StartPriceMonitor(ctx context.Context, userID string) {
	// 用 ticker 替代 sleep，可通过 ctx 停止
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Price monitor stopped for:", userID)
				return

			case <-ticker.C:
				subscriptions, err := s.SubscriptionRepo.List(ctx, userID)
				if err != nil {
					fmt.Println("query err:", err)
					continue
				}

				for _, sub := range subscriptions {
					price := 50 + rand.Float64()*100
					if price > sub.Threshold {
						fmt.Printf("[Notify] %s %s price %.2f > %.2f\n",
							sub.UserID, sub.Symbol, price, sub.Threshold)
					}
				}
			}
		}
	}()
}
