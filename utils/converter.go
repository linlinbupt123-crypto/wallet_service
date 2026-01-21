package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"math/big"

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
func DeriveAESKey(passphrase, salt string) ([]byte, error) {
	// N=16384, r=8, p=1, keyLen=32 (AES-256)
	key, err := scrypt.Key([]byte(passphrase), []byte(salt), 16384, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecryptAES 解密 ciphertext
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) < 1 {
		return nil, errors.New("empty ciphertext")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(key) < block.BlockSize() {
		return nil, errors.New("key too short")
	}

	plainText := make([]byte, len(ciphertext))
	cfb := cipher.NewCFBDecrypter(block, key[:block.BlockSize()])
	cfb.XORKeyStream(plainText, ciphertext)

	return plainText, nil
}
