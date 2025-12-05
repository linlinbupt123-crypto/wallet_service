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
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"

	// "github.com/btcsuite/btcutil/hdkeychain"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/btcsuite/btcd/chaincfg"

	"golang.org/x/crypto/pbkdf2"

	"github.com/linlinbupt123-crypto/wallet_service/entity"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
)

func createHDKey(seed []byte) (*bip32.Key, error) {
    return bip32.NewMasterKey(seed)
}
// ---------- Encryption helpers ----------
func deriveKey(passphrase string, salt []byte) []byte {
	 return pbkdf2.Key([]byte(passphrase), salt, 310_000, 32, sha256.New)
}
// clearBytes, clear keys
func clearBytes(b []byte) {
    for i := range b {
        b[i] = 0
    }
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
type HDWallet struct {
	WalletRepo *repository.Wallet
}

func NewHDWallet() *HDWallet {
	return &HDWallet{WalletRepo: repository.NewWalletRepo()}
}

/*
助记词 → Seed → Master Key(xprv/xpub)
             ↓
         加密 → 持久化
*/
// CreateWallet generates mnemonic, seed, xprv/xpub, encrypts and saves to MongoDB
func (s *HDWallet) CreateWallet(ctx context.Context, userID string, passphrase string) (*entity.HDWallet, error) {
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
    // 使用 bip32 生成主密钥
    masterKey, err := bip32.NewMasterKey(seed)
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
    // 加密存储助记词
    encMnemonic, err := encrypt([]byte(mnemonic), key)
    if err != nil {
        return nil, err
    }

	wallet := &entity.HDWallet{
		UserID:        userID,
        MnemonicEncrypted: encMnemonic,
		EncryptedSeed: encSeed,
		XPrvEncrypted: encXPrv,
		XPub:          xpub.String(),
		SaltHex:       hex.EncodeToString(salt),
		CreatedAt:     time.Now(),
	}
	_, err = s.WalletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *HDWallet) DecryptSeed(wallet *entity.HDWallet, passphrase string) ([]byte, error) {
	// 1. decode salt
	salt, err := hex.DecodeString(wallet.SaltHex)
	if err != nil {
		return nil, err
	}

	// 2. derive AES key
	key := deriveKey(passphrase, salt)

	// 3. decrypt seed
	seed, err := decrypt(wallet.EncryptedSeed, key)
	if err != nil {
		return nil, err
	}

	return seed, nil
}

// LoadWallet decrypts seed/xprv
func (s *HDWallet) LoadWallet(ctx context.Context, userID string, passphrase string) ([]byte, []byte, error) {
	wallet, err := s.WalletRepo.GetByUserID(ctx, userID)
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
func (s *HDWallet) DeriveETHAddress(seed []byte, path string) (string, error) {
    master, err := bip32.NewMasterKey(seed)
    if err != nil {
        return "", err
    }

    // parse path into uint32 indexes (non-hardened same, hardened add FirstHardenedChild)
    indices, err := parseDerivationPath(path)
    if err != nil {
        return "", err
    }

    key := master
    for _, idx := range indices {
        // go-bip32 expects hardened offset as bip32.FirstHardenedChild
        childIndex := idx
        if idx >= hdkeychain.HardenedKeyStart {
            // convert btcsuite hardened flag to bip32.FirstHardenedChild
            childIndex = (idx - hdkeychain.HardenedKeyStart) + bip32.FirstHardenedChild
        }
        key, err = key.NewChildKey(childIndex)
        if err != nil {
            return "", err
        }
    }

    // key.Key is the 32-byte private key; convert to ecdsa
    privKey, err := crypto.ToECDSA(key.Key)
    if err != nil {
        return "", err
    }

    addr := crypto.PubkeyToAddress(privKey.PublicKey)
    return addr.Hex(), nil
}

// Derive BTC address (P2PKH)
func (s *HDWallet) DeriveBTCAddress(seed []byte, path string) (string, error) {
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

    // 得到地址（btcsuite 会根据网络和 key 类型返回合适的地址）
    addr, err := key.Address(&chaincfg.MainNetParams)
    if err != nil {
        return "", err
    }
    return addr.EncodeAddress(), nil
}

// BTC : m/44'/0'/0'/0/0
// ETH : m/44'/60'/0'/0/0
// parseDerivationPath
func parseDerivationPath(path string) ([]uint32, error) {
    // accept "m/44'/60'/0'/0/0" or "44'/60'/0'/0/0"
    p := strings.TrimSpace(path)
    if strings.HasPrefix(p, "m/") || strings.HasPrefix(p, "M/") {
        p = p[2:]
    }

    if p == "" {
        return nil, errors.New("empty derivation path")
    }

    parts := strings.Split(p, "/")
    indices := make([]uint32, 0, len(parts))

    for _, part := range parts {
        if part == "" {
            return nil, errors.New("invalid path segment")
        }
        hardened := strings.HasSuffix(part, "'")
        if hardened {
            part = strings.TrimSuffix(part, "'")
        }
        v, err := strconv.ParseUint(part, 10, 32)
        if err != nil {
            return nil, errors.New("invalid derivation index")
        }
        idx := uint32(v)
        if hardened {
            idx += hdkeychain.HardenedKeyStart
        }
        indices = append(indices, idx)
    }
    return indices, nil
}

func (s *HDWallet) VerifyPassphrase(ctx context.Context, userID string, passphrase string) (bool, error) {
    wallet, err := s.WalletRepo.GetByUserID(ctx, userID)
    if err != nil {
        return false, err
    }
    
    // 尝试解密种子
    _, err = s.DecryptSeed(wallet, passphrase)
    if err != nil {
        // 可能是密码错误或数据损坏
        if err.Error() == "cipher: message authentication failed" {
            return false, nil // 密码错误
        }
        return false, err // 其他错误
    }
    
    return true, nil
}