package bitfinexlender

import (
	"github.com/eAndrius/bitfinex-go"
)

// BotConfig ...
type BotConfig struct {
	Bitfinex BitfinexConf
	Strategy StrategyConf

	API *bitfinex.API
}

// BotConfigs ...
type BotConfigs []BotConfig

// BitfinexConf ...
type BitfinexConf struct {
	APIKey          string
	APISecret       string
	ActiveWallet    string
	MaxActiveAmount float64
	MinLoanUSD      float64
}
