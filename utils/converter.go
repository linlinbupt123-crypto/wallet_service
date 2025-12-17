package utils

import (
	"math/big"
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
