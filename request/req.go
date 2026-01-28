package request

type GetBalanceReq struct {
	Chain string `json:"chain" binding:"required"`
}

// --- 请求结构 ---
type CreateWalletReq struct {
	UserID     string `json:"user_id" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
	ChainName  string `json:"chain_name" binding:"required"`
}

type DeriveAddressRequst struct {
	WalletID   string `json:"wallet_id" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
	ChainName  string `json:"chain_name" binding:"required"`
}

type SendTxReq struct {
	Chain      string `json:"chain" binding:"required"`
	From       string `json:"from" binding:"required"`
	To         string `json:"to" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
}
