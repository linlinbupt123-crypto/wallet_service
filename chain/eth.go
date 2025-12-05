package chain

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto"
)

type ETHChain struct {
	MainNet bool
}

func NewETHChain(mainnet bool) *ETHChain {
	return &ETHChain{MainNet: mainnet}
}

func (e *ETHChain) DeriveAddress(seed []byte, path string) (string, error) {
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams) // hdkeychain 不区分 eth 网络
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

	priv, err := key.ECPrivKey()
	if err != nil {
		return "", err
	}

	ecdsaKey, err := crypto.ToECDSA(priv.Serialize())
	if err != nil {
		return "", err
	}

	address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)
	return address.Hex(), nil
}

