package config

import (
	"os"
)

type Config struct {
	MongoURI string
	EthRPC   string
	BTCRPC   string
	SolRPC   string
	Port     string
}

func LoadFromEnv() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		MongoURI: os.Getenv("MONGO_URI"),
		EthRPC:   os.Getenv("ETH_RPC"),
		BTCRPC:   os.Getenv("BTC_RPC"),
		SolRPC:   os.Getenv("SOL_RPC"),
		Port:     port,
	}
}
