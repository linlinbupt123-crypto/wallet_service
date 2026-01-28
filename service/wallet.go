package service

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/linlinbupt123-crypto/wallet_service/chain"
	"github.com/linlinbupt123-crypto/wallet_service/config"
	"github.com/linlinbupt123-crypto/wallet_service/domain"
	"github.com/linlinbupt123-crypto/wallet_service/entity"
	walletErr "github.com/linlinbupt123-crypto/wallet_service/errors"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
	"github.com/linlinbupt123-crypto/wallet_service/request"
	"github.com/linlinbupt123-crypto/wallet_service/utils"
	"golang.org/x/crypto/scrypt"
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
func (s *WalletService) CreateWalletAndAddresses(ctx context.Context, userID, passphrase string) (*entity.Wallet, map[string]string, error) {
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
func (s *WalletService) DeriveNewAddress(ctx context.Context, userID, walletID, chainName, passphrase string) (string, error) {
	// 1. find wallet
	wallet, err := s.WalletRepo.GetByID(ctx, walletID)
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

// SendTransaction 发起交易（fromAddress 对应你管理的地址）
func (s *WalletService) SendTransaction(ctx context.Context, req *request.SendTxReq) (string, error) {
	fromAddr, err := utils.NormalizeETHAddress(req.From)
	if err != nil {
		return "", err
	}
	toAddr, err := utils.NormalizeETHAddress(req.To)
	if err != nil {
		return "", err
	}
	// 1. address → walletID
	addr, err := s.AddressRepo.GetByAddrID(ctx, fromAddr)
	if err != nil {
		return "", err
	}
	if addr == nil {
		return "", fmt.Errorf("address: %s not found or not belongs to user", fromAddr)
	}
	if addr.Chain != req.Chain {
		return "", errors.New("address not found or not belongs to user")
	}
	index := int32(addr.Index)
	if index < 0 {
		return "", errors.New("address not found or not belongs to user")
	}
	wallet, err := s.WalletRepo.GetByID(ctx, addr.WalletID)
	if err != nil {
		return "", err
	}
	if wallet == nil {
		return "", errors.New("wallet not found")
	}
	// 根据 index 重新 derive 出对应私钥/地址，让 chain 层去签名 & 广播
	switch req.Chain {
	case "btc":
		return "", nil
	case "eth":
		return s.sendTransactionByAddress(ctx, wallet, addr, toAddr, req.Amount, req.Passphrase)
	default:
		return "", errors.New("unsupported chain")
	}
}

func (s *WalletService) sendTransactionByAddress(
	ctx context.Context,
	wallet *entity.Wallet,
	addr *entity.Address,
	toAddress string,
	amount string,
	passphrase string,
) (string, error) {
	var privKey *ecdsa.PrivateKey
	switch wallet.WalletType {
	case "hd":
		seed, err := s.HDWalletDomain.DecryptSeed(wallet, passphrase)
		if err != nil {
			return "", err
		}
		// HD 派生 index 就是 Address 表里的 Index
		path := generatePath(44, 60, 0, 0, int(addr.Index))
		privKey, _, err = s.HDWalletDomain.DeriveETHKeyPair(seed, path)
		if err != nil {
			return "", walletErr.WrapWithCode(walletErr.DeriveErr, "DeriveETHKeyPair", err)
		}
	case "imported":

		key, err := utils.DeriveAESKey(passphrase, wallet.SaltHex)
		if err != nil {
			return "", err
		}

		privKeyBytes, err := utils.DecryptAES(wallet.CipherKey, key)
		if err != nil {
			return "", err
		}

		privKeyHex := strings.TrimPrefix(string(privKeyBytes), "0x")

		privKey, err = crypto.HexToECDSA(privKeyHex)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("unsupported wallet type")
	}
	amountWei, err := utils.ETHToWei(amount)
	if err != nil {
		return "", err
	}

	return s.EthChain.SendETH(ctx, privKey, toAddress, amountWei)
}

func (s *WalletService) GetBalance(
	ctx context.Context,
	userID string,
	chain string,
) (string, error) {

	// 1. 找用户地址（这里简单：取 index = 0 的主地址）
	addrs, err := s.AddressRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	var address string
	for _, a := range addrs {
		if a.Chain == chain && a.Index == 0 {
			address = a.Address
			break
		}
	}
	if address == "" {
		return "", errors.New("address not found")
	}

	// 2. 查链上余额
	switch chain {
	case "eth":
		balanceWei, err := s.EthChain.GetBalance(ctx, address)
		if err != nil {
			return "", err
		}
		return utils.WeiToETH(balanceWei), nil
	default:
		return "", errors.New("unsupported chain")
	}
}

func (s *WalletService) ImportETHPrivateKey(
	ctx context.Context,
	userID, walletName, privKeyHex, passphrase string,
) (*entity.Wallet, *entity.Address, error) {
	// 1. 解码私钥
	privKey, err := crypto.HexToECDSA(privKeyHex[2:]) // 去掉 0x
	if err != nil {
		return nil, nil, errors.New("invalid private key")
	}

	// 2. 生成加密 key
	salt := []byte(userID + time.Now().String()) // 可用随机数也行
	key, _ := scrypt.Key([]byte(passphrase), salt, 16384, 8, 1, 32)

	// 3. 加密私钥
	cipherPriv, err := utils.EncryptAES([]byte(privKeyHex), key)
	if err != nil {
		return nil, nil, err
	}

	// 4. 创建 Wallet
	wallet := &entity.Wallet{
		UserID:     userID,
		WalletName: walletName,
		WalletType: "imported",
		CipherKey:  cipherPriv,
		SaltHex:    hex.EncodeToString(salt),
		CreatedAt:  time.Now(),
	}

	walletID, err := s.WalletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, nil, err
	}
	wallet.ID = walletID

	// 5. 创建 Address
	addr := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	addressEntity := &entity.Address{
		UserID:    userID,
		WalletID:  walletID,
		Chain:     "eth",
		Address:   addr,
		Index:     0,
		Source:    "imported",
		CreatedAt: time.Now(),
	}

	if err := s.AddressRepo.Create(ctx, addressEntity); err != nil {
		return nil, nil, err
	}

	return wallet, addressEntity, nil
}
