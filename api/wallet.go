package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linlinbupt123-crypto/wallet_service/request"
	"github.com/linlinbupt123-crypto/wallet_service/service"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(ws *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: ws}
}

// CreateWallet, create HD wallet and main address
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	userID := c.Param("userID")
	var req request.CreateWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = userID

	wallet, addrs, err := h.walletService.CreateWalletAndAddresses(
		c.Request.Context(), req.UserID, req.Passphrase,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet":    wallet,
		"addresses": addrs,
	})
}

// GetAddresses, get all addresses of a user
func (h *WalletHandler) GetAddresses(c *gin.Context) {
	userID := c.Param("userID")

	ctx := context.Background()
	addrs, err := h.walletService.AddressRepo.GetByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, addrs)
}

// DeriveAddress, derived address
func (h *WalletHandler) DeriveAddress(c *gin.Context) {
	var req request.DeriveAddressRequst
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addr, err := h.walletService.DeriveNewAddress(
		c.Request.Context(),
		req.WalletID,
		req.UserID,
		req.Passphrase,
		req.ChainName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, addr)
}

func (h *WalletHandler) SendTransaction(c *gin.Context) {
	var req request.SendTxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	txHash, err := h.walletService.SendTransaction(
		c.Request.Context(),
		&req,
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"tx_hash": txHash,
	})
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID := c.Param("userID")

	var req request.GetBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	balance, err := h.walletService.GetBalance(
		c.Request.Context(),
		userID,
		req.Chain,
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"chain":   req.Chain,
		"balance": balance, // Wei
	})
}

type ImportWalletReq struct {
	WalletName string `json:"wallet_name" binding:"required"`
	Chain      string `json:"chain" binding:"required"` // 这里只支持 eth
	PrivateKey string `json:"private_key" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
}

func (h *WalletHandler) ImportWallet(c *gin.Context) {
	userID := c.Param("userID")
	var req ImportWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	wallet, addr, err := h.walletService.ImportETHPrivateKey(
		c.Request.Context(),
		userID,
		req.WalletName,
		req.PrivateKey,
		req.Passphrase,
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"wallet_id":   wallet.ID,
		"wallet_name": wallet.WalletName,
		"wallet_type": wallet.WalletType,
		"address":     addr.Address,
	})
}
