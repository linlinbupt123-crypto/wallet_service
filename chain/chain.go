package chain

type ChainType string

const (
	BTC    ChainType = "btc"
	ETH    ChainType = "eth"
	SOLANA ChainType = "solana"
)

// WalletChain 定义统一接口
type WalletChain interface {
	DeriveAddress(seed []byte, path string) (string, error)
}
