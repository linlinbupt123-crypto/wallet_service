package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/linlinbupt123-crypto/wallet_service/chain"
	"github.com/linlinbupt123-crypto/wallet_service/config"
	"github.com/linlinbupt123-crypto/wallet_service/domain"
	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
	"github.com/linlinbupt123-crypto/wallet_service/utils"
)

type WalletService struct {
	HDWalletDomain *domain.HDWallet
	WalletRepo     *repository.Wallet
	AddressRepo    *repository.AddressRepo
	EthChain       *chain.ETHChain
	UseMainNet     bool
}

func NewWalletService(
	hdSvc *domain.HDWallet,
	walletRepo *repository.Wallet,
	addressRepo *repository.AddressRepo,
	EthConfig config.EthConfig,
) *WalletService {
	return &WalletService{
		HDWalletDomain: hdSvc,
		WalletRepo:     walletRepo,
		AddressRepo:    addressRepo,
		EthChain:       chain.NewETHChain(EthConfig),
	}
}

// CreateWalletAndAddresses 创建 HD 钱包 + 主地址
func (s *WalletService) CreateWalletAndAddresses(ctx context.Context, userID, passphrase string) (*entity.HDWallet, map[string]string, error) {
	// 创建 HD 钱包对象
	wallet, err := s.HDWalletDomain.CreateWallet(ctx, userID, passphrase)
	if err != nil {
		return nil, nil, err
	}

	// 存入数据库
	walletID, err := s.WalletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, nil, err
	}
	wallet.ID = walletID

	// 派生主地址
	seed, err := s.HDWalletDomain.DecryptSeed(wallet, passphrase) // 简化示例
	if err != nil {
		return nil, nil, err
	}
	addresses := make(map[string]string)
	// ETH
	_, ethAddr, err := s.HDWalletDomain.DeriveETHKeyPair(seed, utils.ETH_DERIVATION_PATH_PREFIX+"0")
	if err != nil {
		return nil, nil, err
	}
	addresses["eth"] = ethAddr
	if err := s.AddressRepo.Create(ctx, &entity.Address{
		UserID:    userID,
		WalletID:  walletID,
		Chain:     "eth",
		Address:   ethAddr,
		Index:     0,
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, nil, err
	}

	return wallet, addresses, nil
}

// DeriveNewAddress 为用户在某条链派生下一个地址
func (s *WalletService) DeriveNewAddress(ctx context.Context, userID, chainName, passphrase string) (string, error) {
	// 1. 找到该用户的钱包（你可以按 userID 查，也可以传 walletID，这里假设 WalletRepo 有 GetByUserID）
	wallet, err := s.WalletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if wallet == nil {
		return "", errors.New("wallet not found for user")
	}

	// 2. 解密 seed
	seed, err := s.HDWalletDomain.DecryptSeed(wallet, passphrase)
	if err != nil {
		return "", err
	}

	// 3. 找该链目前最大的 index
	maxIndex, err := s.AddressRepo.GetMaxIndex(ctx, wallet.ID, chainName)
	if err != nil {
		return "", err
	}
	nextIndex := maxIndex + 1 // 如果没有记录，GetMaxIndex 会返回 -1，则 nextIndex=0

	// 4. 生成 BIP44 path
	var path string
	switch chainName {
	case "btc":
		path = generatePath(44, 0, 0, 0, nextIndex)
	case "eth":
		path = generatePath(44, 60, 0, 0, nextIndex)
	default:
		return "", errors.New("unsupported chain")
	}

	// 5. 派生地址
	var addr string
	switch chainName {
	case "eth":
		_, addr, err = s.HDWalletDomain.DeriveETHKeyPair(seed, path)
	}
	if err != nil {
		return "", err
	}

	// 6. 存数据库
	err = s.AddressRepo.Create(ctx, &entity.Address{
		UserID:    userID,
		WalletID:  wallet.ID,
		Chain:     chainName,
		Address:   addr,
		Index:     uint32(nextIndex),
		CreatedAt: time.Now(),
	})
	if err != nil {
		return "", err
	}

	return addr, nil
}

// BIP44 path helper: m / purpose' / coin_type' / account' / change / address_index
func generatePath(purpose, coinType, account, change, index int) string {
	// 简单 string 拼装，具体你也可以用专门 BIP32 库
	return fmt.Sprintf("m/%d'/%d'/%d'/%d/%d", purpose, coinType, account, change, index)
}

// GetBalance 查询余额（测试链）
func (s *WalletService) GetBalance(ctx context.Context, chainName, address string) (string, error) {
	switch chainName {
	case "eth":
	//	return s.EthChain.GetBalance(ctx, address)
	default:
		return "", errors.New("unsupported chain")
	}
	return "", errors.New("unsupported chain")
}

// SendTransaction 发起交易（fromAddress 对应你管理的地址）
func (s *WalletService) SendTransaction(ctx context.Context, chainName, toAddress, amount string, passphrase string, userID string) (string, error) {
	// 1. 找钱包
	wallet, err := s.WalletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if wallet == nil {
		return "", errors.New("wallet not found")
	}

	// 2. 解密 seed
	seed, err := s.HDWalletDomain.DecryptSeed(wallet, passphrase)
	if err != nil {
		return "", err
	}

	// 3. 根据地址找到 index（这里简单写法：先查所有地址再匹配）
	addrs, err := s.AddressRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	var index int32 = 0
	for _, a := range addrs {
		if a.Chain == chainName {
			index = int32(a.Index)
			break
		}
	}
	if index < 0 {
		return "", errors.New("address not found or not belongs to user")
	}

	// 4. 根据 index 重新 derive 出对应私钥/地址，让 chain 层去签名 & 广播
	switch chainName {
	case "btc":
		return "", nil
		// path 必须和你生成地址时保持一致
		//path := generatePath(44, 0, 0, 0, int(index))
		// return s.BtcChain.SendTransaction(ctx, seed, path, fromAddress, toAddress, amount)
	case "eth":
		path := generatePath(44, 60, 0, 0, int(index))

		privKey, _, err := s.HDWalletDomain.DeriveETHKeyPair(seed, path)
		if err != nil {
			return "", err
		}

		amountWei, _ := new(big.Int).SetString(amount, 10)

		return s.EthChain.SendETH(ctx, privKey, toAddress, amountWei)
	default:
		return "", errors.New("unsupported chain")
	}
}
