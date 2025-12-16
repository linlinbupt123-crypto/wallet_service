package chain

import (
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

type BTCChain struct {
	MainNet bool
}

func NewBTCChain(mainnet bool) *BTCChain {
	return &BTCChain{MainNet: mainnet}
}

func (b *BTCChain) netParams() *chaincfg.Params {
	if b.MainNet {
		return &chaincfg.MainNetParams
	}
	return &chaincfg.TestNet3Params
}

func (b *BTCChain) DeriveAddress(seed []byte, path string) (string, error) {
	master, err := hdkeychain.NewMaster(seed, b.netParams())
	if err != nil {
		return "", err
	}

	indices, err := parseDerivationPath(path)
	if err != nil {
		return "", err
	}

	key := master
	for _, idx := range indices {
		key, err = key.Derive(idx)
		if err != nil {
			return "", err
		}
	}

	addr, err := key.Address(b.netParams())
	if err != nil {
		return "", err
	}

	return addr.EncodeAddress(), nil
}
