package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linlinbupt123-crypto/wallet_service/service"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(ws *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: ws}
}

// --- 请求结构 ---
type CreateWalletReq struct {
	UserID     string `json:"user_id" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
	ChainName  string `json:"chain_name" binding:"required"`
}

type DeriveAddressRequst struct {
	UserID     string `json:"user_id" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
	ChainName  string `json:"chain_name" binding:"required"`
}

// CreateWallet, create HD wallet and main address
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	userID := c.Param("userID")
	var req CreateWalletReq
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
	var req DeriveAddressRequst
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addr, err := h.walletService.DeriveNewAddress(
		c.Request.Context(),
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

type SendTxReq struct {
	Chain      string `json:"chain" binding:"required"`
	From       string `json:"from" binding:"required"`
	To         string `json:"to" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	Passphrase string `json:"passphrase" binding:"required"`
}

func (h *WalletHandler) SendTransaction(c *gin.Context) {
	userID := c.Param("userID")

	var req SendTxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	txHash, err := h.walletService.SendTransaction(
		c.Request.Context(),
		req.Chain,
		req.From,
		req.To,
		req.Amount,
		req.Passphrase,
		userID,
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"tx_hash": txHash,
	})
}
