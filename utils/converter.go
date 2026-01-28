package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/scrypt"
)

func WeiToETH(wei *big.Int) string {
	f := new(big.Float).SetInt(wei)
	f.Quo(f, big.NewFloat(1e18))
	return f.Text('f', 18) // 18 位小数
}

func ETHToWei(eth string) (*big.Int, error) {
	f, _, err := big.ParseFloat(eth, 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, err
	}
	f.Mul(f, big.NewFloat(1e18))
	wei := new(big.Int)
	f.Int(wei)
	return wei, nil
}

// DeriveAESKey 通过 passphrase + salt 派生 AES key
func DeriveAESKey(passphrase, saltHex string) ([]byte, error) {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return nil, err
	}
	// N=16384, r=8, p=1, keyLen=32 (AES-256)
	key, err := scrypt.Key([]byte(passphrase), salt, 16384, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecryptAES 解密 ciphertext
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
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

	nonce := ciphertext[:nonceSize]
	encrypted := ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		// 包含：密钥错误 / 数据被篡改 / tag 校验失败
		return nil, errors.New("decrypt failed or data corrupted")
	}

	return plaintext, nil
}
func EncryptAES(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 返回：nonce + ciphertext
	return append(nonce, ciphertext...), nil
}

// NormalizeETHAddress
// - 接受任意合法 ETH 地址（小写 / 大写 / checksum）
// - 返回标准 checksum 地址
func NormalizeETHAddress(addr string) (string, error) {
	if !common.IsHexAddress(addr) {
		return "", errors.New("invalid ethereum address")
	}
	return common.HexToAddress(addr).Hex(), nil
}
