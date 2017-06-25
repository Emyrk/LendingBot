package balancer

import (
	//"fmt"
	//"sync"
	//"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	//"github.com/Emyrk/LendingBot/src/core/userdb"

	"go.uber.org/ratelimit"
)

type PoloniexAPIWithRateLimit struct {
	polo *poloniex.Poloniex

	limiter ratelimit.Limiter
}

func NewPoloniexAPIWithRateLimit() *PoloniexAPIWithRateLimit {
	p := new(PoloniexAPIWithRateLimit)
	p.limiter = ratelimit.New(6)
	p.polo = poloniex.StartPoloniex()
	return p
}

func (p *PoloniexAPIWithRateLimit) take() {
	//n := time.Now()
	p.limiter.Take()
	//PoloCallTakeWait.Observe(float64(time.Since(n).Nanoseconds()))
}

func (p *PoloniexAPIWithRateLimit) GetTicker() (map[string]poloniex.PoloniexTicker, error) {
	p.take()
	return p.polo.GetTicker()
}

func (p *PoloniexAPIWithRateLimit) GetLoanOrders(currency string) (*poloniex.PoloniexLoanOrders, error) {
	p.take()
	return p.polo.GetLoanOrders(currency)
}
