package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/linlinbupt123-crypto/wallet_service/api"
	"github.com/linlinbupt123-crypto/wallet_service/config"
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
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	walletService := service.NewWalletService(
		hdDomain,
		walletRepo,
		addressRepo,
		cfg.Eth,
	)

	// 3. Gin
	r := gin.Default()

	walletHandler := api.NewWalletHandler(walletService)

	// 1) 创建 HD 钱包 + 地址
	r.POST("/wallet/:userID", walletHandler.CreateWallet)

	// 2) 获取用户所有地址
	r.GET("/wallet/:userID/addresses", walletHandler.GetAddresses)

	// 3) 派生新的地址
	r.POST("/wallet/:userID/address/new", walletHandler.DeriveAddress) // ?chain=btc

	r.POST("/wallet/:userID/tx/send", walletHandler.SendTransaction)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
