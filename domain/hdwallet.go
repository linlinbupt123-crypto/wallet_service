package domain

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/btcsuite/btcutil/hdkeychain"
	bip39 "github.com/tyler-smith/go-bip39"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/btcsuite/btcd/chaincfg"

	"golang.org/x/crypto/pbkdf2"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
)

// ---------- Encryption helpers ----------
func deriveKey(passphrase string, salt []byte) []byte {
    return pbkdf2.Key([]byte(passphrase), salt, 100_000, 32, sha256.New)
}

func encrypt(data []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, gcm.NonceSize())
    _, err = io.ReadFull(rand.Reader, nonce)
    if err != nil {
        return nil, err
    }
    return append(nonce, gcm.Seal(nil, nonce, data, nil)...), nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }
    nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
    return gcm.Open(nil, nonce, ct, nil)
}

// ---------- Wallet service ----------
type HDWalletService struct {
    repo *repository.WalletRepo
}

func NewHDWalletService(repo *repository.WalletRepo) *HDWalletService {
    return &HDWalletService{repo: repo}
}
/*
助记词 → Seed → Master Key(xprv/xpub)
             ↓
         加密 → 持久化
*/
// CreateWallet generates mnemonic, seed, xprv/xpub, encrypts and saves to MongoDB
func (s *HDWalletService) CreateWallet(ctx context.Context, userID string, passphrase string) (*entity.HDWallet, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
    if err != nil {
        return nil, err
    }
    seed := bip39.NewSeed(mnemonic, "")
	// build private master key by seed
    masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
    if err != nil {
        return nil, err
    }
	// xpub 可以生成地址（Bitcoin/Ethereum/其它链）
    // xpub 无法签名（没有私钥）
	// Derive the xpub (public extended key) from the xprv (private extended key), without exposing any private key.
    xpub, err := masterKey.Neuter()
    if err != nil {
        return nil, err
    }
	xprvStr := masterKey.String() // the master private extended key (xprv)
	
	salt := make([]byte, 16)
_, err = rand.Read(salt)
if err != nil {
    return nil, err
}

key := deriveKey(passphrase, salt)

encSeed, err := encrypt(seed, key)
if err != nil {
    return nil, err
}

encXPrv, err := encrypt([]byte(xprvStr), key)
if err != nil {
    return nil, err
}

    wallet := &entity.HDWallet{
        UserID:        userID,
        EncryptedSeed: encSeed,
        XPrvEncrypted: encXPrv,
        XPub:          xpub.String(),
        SaltHex:       hex.EncodeToString(salt),
        CreatedAt:     time.Now(),
    }
    err = s.repo.Save(ctx, wallet)
    if err != nil {
        return nil, err
    }
    return wallet, nil
}

// LoadWallet decrypts seed/xprv
func (s *HDWalletService) LoadWallet(ctx context.Context, userID string, passphrase string) ([]byte, []byte, error) {
    wallet, err := s.repo.GetByUserID(ctx, userID)
    if err != nil {
        return nil, nil, err
    }
    salt, _ := hex.DecodeString(wallet.SaltHex)
    key := deriveKey(passphrase, salt)
    seed, err := decrypt(wallet.EncryptedSeed, key)
    if err != nil {
        return nil, nil, err
    }
    xprv, err := decrypt(wallet.XPrvEncrypted, key)
    if err != nil {
        return nil, nil, err
    }
    return seed, xprv, nil
}

// Derive ETH address (BIP32/BIP44)
func (s *HDWalletService) DeriveETHAddress(seed []byte, path string) (string, error) {
    // 1. master key
    master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
    if err != nil {
        return "", err
    }

    // 2. 解析 "m/44'/60'/0'/0/0" 这种路径
    indices, err := parseDerivationPath(path)
    if err != nil {
        return "", err
    }

    // 3. 依次派生子 key
    key := master
    for _, idx := range indices {
        key, err = key.Child(idx)
        if err != nil {
            return "", err
        }
    }

    // 4. 获取私钥
    priv, err := key.ECPrivKey()
    if err != nil {
        return "", err
    }

    // 5. 转换为 go-ethereum 的 ECDSA 格式
    ecdsaKey, err := crypto.ToECDSA(priv.Serialize())
    if err != nil {
        return "", err
    }

    // 6. 得到 ETH address
    address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)
    return address.Hex(), nil
}

// Derive BTC address (P2PKH)
func (s *HDWalletService) DeriveBTCAddress(seed []byte, path string) (string, error) {
  master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
    if err != nil {
        return "", err
    }

    indices, err := parseDerivationPath(path)
    if err != nil {
        return "", err
    }

    key := master
    for _, idx := range indices {
        key, err = key.Child(idx)
        if err != nil {
            return "", err
        }
    }

    addr, err := key.Address(&chaincfg.MainNetParams)
    if err != nil {
        return "", err
    }

    return addr.EncodeAddress(), nil
}

// Parse derivation path like m/44'/60'/0'/0/0
func parseDerivationPath(path string) ([]uint32, error) {
    // 这里用之前的 parsePath 方法
    return []uint32{44 + hdkeychain.HardenedKeyStart, 60 + hdkeychain.HardenedKeyStart, 0 + hdkeychain.HardenedKeyStart, 0, 0}, nil
}