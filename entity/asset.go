package entity

type Asset struct {
    Chain    string
    Symbol   string  // BTC / ETH / USDT …
    Balance  float64
    Address  string
}
