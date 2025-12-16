package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	MongoURI string
	BTCRPC   string
	SolRPC   string
	Port     string
	Eth      EthConfig
}

type EthConfig struct {
	RPC       string `mapstructure:"rpc"`
	TestToken string `mapstructure:"test_token"`
	ChainID   int64  `mapstructure:"chain_id"`
	MainNet   bool   `mapstructure:"main_net"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// ENV 覆盖 YAML
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
