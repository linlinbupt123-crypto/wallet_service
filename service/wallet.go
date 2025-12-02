package service

import (
	"context"
	"time"

	"github.com/linlinbupt123-crypto/wallet_service/domain"
	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
	"ygithub.com/linlinbupt123-crypto/wallet_service/chain"
)

type WalletService struct {
	hdSvc          *domain.HDWalletService
	walletRepo     *repository.WalletRepo
	addressRepo    *repository.AddressRepo
	btcChain       *chain.BTCChain
	ethChain       *chain.ETHChain
	solanaChain    *chain.SolanaChain
	useMainNet     bool
}

func NewWalletService(
	hdSvc *domain.HDWalletService,
	walletRepo *repository.WalletRepo,
	addressRepo *repository.AddressRepo,
	useMainNet bool,
) *WalletService {
	return &WalletService{
		hdSvc:       hdSvc,
		walletRepo:  walletRepo,
		addressRepo: addressRepo,
		btcChain:    chain.NewBTCChain(useMainNet),
		ethChain:    chain.NewETHChain(useMainNet),
		solanaChain: chain.NewSolanaChain(useMainNet),
		useMainNet:  useMainNet,
	}
}

// CreateWalletAndAddresses 创建 HD 钱包 + 主地址
func (s *WalletService) CreateWalletAndAddresses(ctx context.Context, userID, passphrase string) (*entity.HDWallet, map[string]string, error) {
	// 创建 HD 钱包对象
	wallet, err := s.hdSvc.CreateWallet(ctx, userID, passphrase)
	if err != nil {
		return nil, nil, err
	}

	// 存入数据库
	walletID, err := s.walletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, nil, err
	}
	wallet.ID = walletID

	// 4. 派生主地址
	seed := domain.DecryptSeedOrPanic(wallet, s.hdSvc) // 简化示例
	addresses := make(map[string]string)

	// BTC
	btcAddr, err := s.btcChain.DeriveAddress(seed, "m/44'/0'/0'/0/0")
	if err != nil {
		return nil, nil, err
	}
	addresses["btc"] = btcAddr
	s.addressRepo.Create(ctx, &entity.Address{
		UserID:   userID,
		WalletID: walletID,
		Chain:    "btc",
		Address:  btcAddr,
		Index:    0,
		CreatedAt: time.Now(),
	})

	// ETH
	ethAddr, err := s.ethChain.DeriveAddress(seed, "m/44'/60'/0'/0/0")
	if err != nil {
		return nil, nil, err
	}
	addresses["eth"] = ethAddr
	s.addressRepo.Create(ctx, &entity.Address{
		UserID:   userID,
		WalletID: walletID,
		Chain:    "eth",
		Address:  ethAddr,
		Index:    0,
		CreatedAt: time.Now(),
	})

	// Solana
	solAddr, err := s.solanaChain.DeriveAddress(seed, "m/44'/501'/0'/0'")
	if err != nil {
		return nil, nil, err
	}
	addresses["solana"] = solAddr
	s.addressRepo.Create(ctx, &entity.Address{
		UserID:   userID,
		WalletID: walletID,
		Chain:    "solana",
		Address:  solAddr,
		Index:    0,
		CreatedAt: time.Now(),
	})

	return wallet, addresses, nil
}
