package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
)

type WalletRepo struct {
	wallets map[string]*entity.HDWallet
	assets  map[string]map[string]*entity.Asset  // walletID -> symbol -> asset
	subs    map[string]*entity.Subscription      // userID_symbol -> threshold
	mu      sync.RWMutex
}

func NewWalletRepo() *WalletRepo {
	return &WalletRepo{
		wallets: make(map[string]*entity.HDWallet),
		assets:  make(map[string]map[string]*entity.Asset),
		subs:    make(map[string]*entity.Subscription),
	}
}

// Wallet 操作
func (r *WalletRepo) SaveWallet(ctx context.Context, w *entity.HDWallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wallets[w.UserID] = w
	return nil
}
func (r *WalletRepo) GetWallet(ctx context.Context, userID string) (*entity.HDWallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.wallets[userID]
	if !ok {
		return nil, errors.New("wallet not found")
	}
	return w, nil
}

// Asset 操作
func (r *WalletRepo) AddAsset(walletID string, asset *entity.Asset) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.assets[walletID] == nil {
		r.assets[walletID] = make(map[string]*entity.Asset)
	}
	r.assets[walletID][asset.Symbol] = asset
}
func (r *WalletRepo) GetAssets(walletID string) []*entity.Asset {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*entity.Asset
	for _, a := range r.assets[walletID] {
		res = append(res, a)
	}
	return res
}

// Subscription 操作
func (r *WalletRepo) AddSubscription(sub *entity.Subscription) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := sub.UserID + "_" + sub.Symbol
	r.subs[key] = sub
}
func (r *WalletRepo) ListSubscriptions() []*entity.Subscription {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*entity.Subscription
	for _, s := range r.subs {
		res = append(res, s)
	}
	return res
}
