package balancer

import (
	"fmt"
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
	p.limiter = ratelimit.New(4)
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

func (p *PoloniexAPIWithRateLimit) PoloniexGetBalances(accessKey, secret string) (*poloniex.PoloniexBalance, error) {
	p.take()
	return p.polo.GetBalances(accessKey, secret)
}

// PoloniexGetAvailableBalances returns:
//		map[string] :: key = "exchange", "lending", "margin"
//		|-->	map[string] :: key = currency
func (p *PoloniexAPIWithRateLimit) PoloniexGetAvailableBalances(accessKey, secret string) (map[string]map[string]float64, error) {
	p.take()
	return p.polo.GetAvilableBalances(accessKey, secret)
}

// PoloniexCreateLoanOffer creates a loan offer
func (p *PoloniexAPIWithRateLimit) PoloniexCreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, accessKey, secret string) (int64, error) {
	p.take()
	return p.polo.CreateLoanOffer(currency, amount, rate, duration, autoRenew, accessKey, secret)
}

// PoloniexGetInactiveLoans returns your current loans that are not taken
func (p *PoloniexAPIWithRateLimit) PoloniexGetInactiveLoans(accessKey, secret string) (map[string][]poloniex.PoloniexLoanOffer, error) {
	p.take()
	return p.polo.GetOpenLoanOffers(accessKey, secret)
}

// PoloniexGetActiveLoans returns your current loans that are taken
func (p *PoloniexAPIWithRateLimit) PoloniexGetActiveLoans(accessKey, secret string) (*poloniex.PoloniexActiveLoans, error) {
	p.take()
	return p.polo.GetActiveLoans(accessKey, secret)
}

func (p *PoloniexAPIWithRateLimit) PoloniexCancelLoanOffer(currency string, orderNumber int64, accessKey, secret string) (bool, error) {
	p.take()
	return p.polo.CancelLoanOffer(currency, orderNumber, accessKey, secret)
}

func (p *PoloniexAPIWithRateLimit) PoloniexGetLoanOrders(currency string) (*poloniex.PoloniexLoanOrders, error) {
	p.take()
	PoloPublicCalls.Inc()
	return p.polo.GetLoanOrders(currency)
}

func (p *PoloniexAPIWithRateLimit) PoloniexSingleAuthenticatedTradeHistory(currency, accessKey, secret, start, end string) (resp poloniex.PoloniexAuthenticatedTradeHistoryResponse, err error) {
	if currency == "all" {
		return resp, fmt.Errorf("Cannot be 'all'")
	}

	p.take()
	respNonCast, err := p.polo.GetAuthenticatedTradeHistory(currency, start, end, accessKey, secret)
	resp = respNonCast.(poloniex.PoloniexAuthenticatedTradeHistoryResponse)
	return
}

func (p *PoloniexAPIWithRateLimit) PoloniexAllAuthenticatedTradeHistory(accessKey, secret, start, end string) (resp poloniex.PoloniexAuthenticatedTradeHistoryAll, err error) {
	p.take()
	respNonCast, err := p.polo.GetAuthenticatedTradeHistory("", start, end, accessKey, secret)
	resp = respNonCast.(poloniex.PoloniexAuthenticatedTradeHistoryAll)
	return
}

func (p *PoloniexAPIWithRateLimit) PoloniexAuthenticatedLendingHistory(accessKey, secret, start, end, limit string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
	p.take()
	resp, err = p.polo.GetAuthenticatedLendingHistory(start, end, limit, accessKey, secret)
	return
}
