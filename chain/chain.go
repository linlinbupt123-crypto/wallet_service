package chain

type ChainService interface {
    DeriveAddress(masterPubKey []byte, path string) (string, error)
    GetBalance(address string) (float64, error)
    BuildTransaction(fromAddress string, toAddress string, amount float64, opts map[string]interface{}) ([]byte, error)
}
