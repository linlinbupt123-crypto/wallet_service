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
	WalletRepo *repository.Wallet
	SubscriptionRepo *repository.Subscription
	HDWalletDomain *HDWallet
}

// NewWalletService 整合 HDWalletService
func NewWalletService(repo *repository.Wallet) *Wallet {
	hd := NewHDWallet(repo)
	return &Wallet{
		WalletRepo:repository.NewWalletRepo(), 
		HDWalletDomain:hd,
		SubscriptionRepo:repository.NewSubscriptionRepo(),
	}
}

// 创建钱包
func (s *Wallet) CreateWallet(ctx context.Context, userID, passphrase string) (*entity.HDWallet, error) {
	return s.HDWalletDomain.CreateWallet(ctx, userID, passphrase)
}

// 添加资产
func (s *Wallet) AddAsset(ctx context.Context, walletID, symbol string) {
	asset := &entity.Asset{Symbol: symbol, Balance: 0}
	s.HDWalletDomain.AddAsset(walletID, asset)
}

// 查询资产
func (s *Wallet) GetAssets(ctx context.Context, walletID string) []*entity.Asset {
	return s.WalletRepo.GetAssets(walletID)
}

// 模拟转账
func (s *Wallet) Transfer(walletID, symbol string, amount float64) (string, error) {
	assets := s.WalletRepo.GetAssets(walletID)
	var a *entity.Asset
	for _, asset := range assets {
		if asset.Symbol == symbol {
			a = asset
			break
		}
	}
	if a == nil || a.Balance < amount {
		return "", fmt.Errorf("insufficient balance")
	}
	a.Balance -= amount
	txID := fmt.Sprintf("tx_%d", rand.Intn(1000000))
	return txID, nil
}

// 价格订阅
func (s *Wallet) SubscribePrice(userID, symbol string, threshold float64) {
	sub := &entity.Subscription{UserID: userID, Symbol: symbol, Threshold: threshold}
	s.repo.AddSubscription(sub)
}

// 异步价格监控
func (s *Wallet) StartPriceMonitor() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			for _, sub := range s.WalletRepo.ListSubscriptions() {
				price := 50 + rand.Float64()*100
				if price > sub.Threshold {
					fmt.Printf("[Notify] %s %s price %.2f > %.2f\n", sub.UserID, sub.Symbol, price, sub.Threshold)
				}
			}
		}
	}()
}
