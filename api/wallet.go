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
}

// --- 创建钱包 + 主地址 ---
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req CreateWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	wallet, addrs, err := h.walletService.CreateWalletAndAddresses(
		ctx, req.UserID, req.Passphrase,
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

// 获取用户全部地址
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