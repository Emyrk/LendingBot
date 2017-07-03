package balancer_test

import (
	"testing"

	. "github.com/Emyrk/LendingBot/balancer"
)

func TestRateCalc(t *testing.T) {
	var c [32]byte
	b := NewBalancer(c, "mongodb://localhost:27017", "", "")
	b.RateCalculator.UpdateExchangeStats(PoloniexExchange)
	b.RateCalculator.UpdateTicker()
	err := b.RateCalculator.CalculateLoanRate(PoloniexExchange, "FCT")
	if err != nil {
		t.Error(err)
	}
}
