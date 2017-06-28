package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	//"github.com/Emyrk/LendingBot/src/core/userdb"

	"go.uber.org/ratelimit"
)

var limiter ratelimit.Limiter

func init() {
	limiter = ratelimit.New(5)
}

var _ = fmt.Println

func Take() {
	take()
}

func take() {
	n := time.Now()
	limiter.Take()
	PoloCallTakeWait.Observe(float64(time.Since(n).Nanoseconds()))
}

type Lockable interface {
	Lock()
	Unlock()
}

type PoloniexAccessCache struct {
	Cache map[string]*PoloniexAccessStruct
	sync.RWMutex

	Remove chan string
}

func NewPoloniexAccessCache() *PoloniexAccessCache {
	p := new(PoloniexAccessCache)
	p.Cache = make(map[string]*PoloniexAccessStruct)

	return p
}

type PoloniexAccessStruct struct {
	Username string
	APIKey   string
	Secret   string

	LastStatsUpdate time.Time
	sync.Mutex
}

func (p *PoloniexAccessCache) shouldRecordStats(username string) bool {
	p.Lock()
	defer p.Unlock()
	if v, ok := p.Cache[username]; ok {
		if time.Since(v.LastStatsUpdate).Seconds() > 10*60 {
			v.LastStatsUpdate = time.Now()
			p.Cache[username] = v
			return true
		}
	}
	return false
}

func (s *State) removeFromPoloniexCache(username string) {
	s.poloniexCache.Remove <- username
}

func (s *State) updateCache() {
	for {
		select {
		case u := <-s.poloniexCache.Remove:
			s.poloniexCache.Lock()
			delete(s.poloniexCache.Cache, u)
			s.poloniexCache.Unlock()
		default:
			return
		}
	}
}

// getAccessAndSecret will return the access struct, but will also LOCK it. Be sure to unlock it
func (s *State) getAccessAndSecret(username string) (*PoloniexAccessStruct, error) {
	s.updateCache()
	s.poloniexCache.RLock()
	c, ok := s.poloniexCache.Cache[username]
	s.poloniexCache.RUnlock()
	if ok {
		c.Lock()
		return c, nil
	}

	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return nil, err
	}

	ck := u.GetCipherKey(s.CipherKey)
	accessKey, err := u.PoloniexKeys.DecryptAPIKeyString(ck)
	if err != nil {
		return nil, err
	}

	secretKey, err := u.PoloniexKeys.DecryptAPISecretString(ck)
	if err != nil {
		return nil, err
	}
	s.poloniexCache.Lock()
	tmp := new(PoloniexAccessStruct)
	tmp.Username = username
	tmp.APIKey = accessKey
	tmp.Secret = secretKey
	s.poloniexCache.Cache[username] = tmp
	s.poloniexCache.Unlock()

	tmp.Lock()
	return tmp, nil
}

func (s *State) PoloniexGetBalances(username string) (*poloniex.PoloniexBalance, error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}
	defer acc.Unlock()

	take()
	return s.PoloniexAPI.GetBalances(acc.APIKey, acc.Secret)
}

// PoloniexGetAvailableBalances returns:
//		map[string] :: key = "exchange", "lending", "margin"
//		|-->	map[string] :: key = currency
func (s *State) PoloniexGetAvailableBalances(username string) (map[string]map[string]float64, error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}
	defer acc.Unlock()

	take()
	return s.PoloniexAPI.GetAvilableBalances(acc.APIKey, acc.Secret)
}

// PoloniexCreateLoanOffer creates a loan offer
func (s *State) PoloniexCreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, username string) (int64, error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return 0, err
	}
	defer acc.Unlock()

	take()
	return s.PoloniexAPI.CreateLoanOffer(currency, amount, rate, duration, autoRenew, acc.APIKey, acc.Secret)
}

// PoloniexGetInactiveLoans returns your current loans that are not taken
func (s *State) PoloniexGetInactiveLoans(username string) (map[string][]poloniex.PoloniexLoanOffer, error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}
	defer acc.Unlock()

	take()
	return s.PoloniexAPI.GetOpenLoanOffers(acc.APIKey, acc.Secret)
}

// PoloniexGetActiveLoans returns your current loans that are taken
func (s *State) PoloniexGetActiveLoans(username string) (*poloniex.PoloniexActiveLoans, error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}
	defer acc.Unlock()

	take()
	return s.PoloniexAPI.GetActiveLoans(acc.APIKey, acc.Secret)
}

func (s *State) PoloniexCancelLoanOffer(currency string, orderNumber int64, username string) (bool, error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return false, err
	}
	defer acc.Unlock()

	take()
	return s.PoloniexAPI.CancelLoanOffer(currency, orderNumber, acc.APIKey, acc.Secret)
}

func (s *State) PoloniexGetLoanOrders(currency string) (*poloniex.PoloniexLoanOrders, error) {
	take()
	PoloPublicCalls.Inc()
	return s.PoloniexAPI.GetLoanOrders(currency)
}

func (s *State) PoloniexGetTicker() (map[string]poloniex.PoloniexTicker, error) {
	take()
	PoloPublicCalls.Inc()
	return s.PoloniexAPI.GetTicker()
}

func (s *State) PoloniexSingleAuthenticatedTradeHistory(currency, username, start, end string) (resp poloniex.PoloniexAuthenticatedTradeHistoryResponse, err error) {
	if currency == "all" {
		return resp, fmt.Errorf("Cannot be 'all'")
	}
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}
	defer acc.Unlock()

	take()
	respNonCast, err := s.PoloniexAPI.GetAuthenticatedTradeHistory(currency, start, end, acc.APIKey, acc.Secret)
	resp = respNonCast.(poloniex.PoloniexAuthenticatedTradeHistoryResponse)
	return
}

func (s *State) PoloniexAllAuthenticatedTradeHistory(username, start, end string) (resp poloniex.PoloniexAuthenticatedTradeHistoryAll, err error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}
	defer acc.Unlock()

	take()
	respNonCast, err := s.PoloniexAPI.GetAuthenticatedTradeHistory("", start, end, acc.APIKey, acc.Secret)
	resp = respNonCast.(poloniex.PoloniexAuthenticatedTradeHistoryAll)
	return
}

func (s *State) PoloniexOffloadAuthenticatedLendingHistory(username, start, end, limit string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}
	defer acc.Unlock()

	req, err := s.PoloniexAPI.ConstructAuthenticatedLendingHistoryRequest(start, end, limit, acc.APIKey, acc.Secret)
	if err != nil {
		return resp, err
	}

	sendResp, err := s.Master.SendConstructedCall(req)
	if err != nil {
		return resp, err
	}

	if sendResp.Err != nil {
		return resp, sendResp.Err
	}

	err = poloniex.JSONDecode([]byte(sendResp.Response), &resp.Data)
	if err != nil {
		err = fmt.Errorf("%s :: Resp: %s", err.Error(), sendResp.Response)
	}
	return
}

func (s *State) PoloniexAuthenticatedLendingHistory(username, start, end, limit string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
	acc, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}
	defer acc.Unlock()

	take()
	resp, err = s.PoloniexAPI.GetAuthenticatedLendingHistory(start, end, limit, acc.APIKey, acc.Secret)
	return
}
