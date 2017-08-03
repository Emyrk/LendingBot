package bee

import (
	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
)

type IGlobalServer interface {
	GetPoloniexTicker(currency string) (poloniex.PoloniexTicker, bool)
	SetPoloniexTicker(currency string, t poloniex.PoloniexTicker)

	GetAmtForBTCValue(amount float64, currency string) float64
	GetBTCAmount(amount float64, currency string) float64
	GetLoanRate(exch int, currency string) (balancer.LoanRates, bool)
	SavePoloniexMonth(username, accesskey, secretkey string) bool
	SaveBitfinexMonth(username, accesskey, secretkey string) bool
}
