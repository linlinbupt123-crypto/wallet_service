package main

import (
	"github.com/gin-gonic/gin"
	"github.com/linlinbupt123-crypto/wallet_service/api"
	"github.com/linlinbupt123-crypto/wallet_service/db"
	"github.com/linlinbupt123-crypto/wallet_service/domain"
	"github.com/linlinbupt123-crypto/wallet_service/repository"
	"github.com/linlinbupt123-crypto/wallet_service/service"
)

func main() {
	// 1. 初始化 MongoDB
	db.InitMongo()

	// 2. 初始化依赖
	hdDomain := domain.NewHDWallet()
	walletRepo := repository.NewWalletRepo()
	addressRepo := repository.NewAddressRepo()

	// false = testnet（BTC testnet, ETH sepolia/goerli 等）
	walletService := service.NewWalletService(
		hdDomain,
		walletRepo,
		addressRepo,
		false,
	)

	// 3. Gin
	r := gin.Default()

	walletHandler := api.NewWalletHandler(walletService)

	// 1) 创建 HD 钱包 + 地址
	r.POST("/wallet/create", walletHandler.CreateWallet)

	// 2) 获取用户所有地址
	r.GET("/wallet/:userID/addresses", walletHandler.GetAddresses)

	// 3) 派生新的地址
	r.POST("/wallet/:userID/address/new", walletHandler.DeriveNewAddress) // ?chain=btc

	// 4) 查询余额
	r.GET("/wallet/balance", walletHandler.GetBalance) // ?chain=eth&address=xxx

	// 5) 发起交易（测试链）
	r.POST("/wallet/send", walletHandler.SendTransaction)

	r.Run(":8080")
}
