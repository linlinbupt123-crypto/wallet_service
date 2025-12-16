package errors

type Code string

const (
	CodeChainRPC    Code = "CHAIN_RPC_ERROR"
	CodeGasEstimate Code = "GAS_ESTIMATE_ERROR"
	PendingNonceAt  Code = "PENDING_NONCE_AT_ERROR"
	DailChain       Code = "DIAL_CHAIN_ERROR"
	SignerErr       Code = "SIGNER_ERROR"
	SendTxErr       Code = "SEND_TX_ERROR"
	GetchainIDErr   Code = "GET_CHAIN_ID_ERROR"
)
