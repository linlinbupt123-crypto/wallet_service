package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/linlinbupt123-crypto/wallet_service/config"
	wrapErrors "github.com/linlinbupt123-crypto/wallet_service/errors"
)

type ETHChain struct {
	Rpc       string
	ChainID   *big.Int
	TestToken string
	MainNet   bool
}

func NewETHChain(cfg config.EthConfig) *ETHChain {
	return &ETHChain{
		Rpc:       cfg.RPC,
		ChainID:   big.NewInt(cfg.ChainID),
		TestToken: cfg.TestToken,
	}
}

func (e *ETHChain) SendETH(ctx context.Context, priv *ecdsa.PrivateKey, to string, amountWei *big.Int) (string, error) {
	client, err := ethclient.Dial(e.Rpc)
	if err != nil {
		return "", wrapErrors.WrapWithCode(wrapErrors.DailChain, "eth dial", err)
	}
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return "", wrapErrors.WrapWithCode(wrapErrors.GetchainIDErr, "get chainID", err)
	}
	e.ChainID = chainID

	fromAddr := crypto.PubkeyToAddress(priv.PublicKey)
	tmpaddr := fromAddr.Hex()
	fmt.Printf("from address: %s\n", tmpaddr)
	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", wrapErrors.WrapWithCode(wrapErrors.PendingNonceAt, "PendingNonceAt", err)
	}

	tip, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return "", err
	}

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return "", err
	}

	baseFee := header.BaseFee

	feeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)), // 留 buffer
		tip,
	)
	toAddr := common.HexToAddress(to)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   e.ChainID,
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		Gas:       21000,
		To:        &toAddr,
		Value:     amountWei,
	})

	signer := types.NewLondonSigner(e.ChainID)
	signedTx, err := types.SignTx(tx, signer, priv)
	if err != nil {
		return "", wrapErrors.WrapWithCode(wrapErrors.SignerErr, "SignTx", err)
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return "", wrapErrors.WrapWithCode(wrapErrors.SendTxErr, "SendTransaction", err)
	}

	return signedTx.Hash().Hex(), nil
}
