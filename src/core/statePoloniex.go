package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	//"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Println

type PoloniexAccessCache struct {
	Cache map[string]PoloniexAccessStruct
	sync.RWMutex

	Remove chan string
}

func NewPoloniexAccessCache() *PoloniexAccessCache {
	p := new(PoloniexAccessCache)
	p.Cache = make(map[string]PoloniexAccessStruct)

	return p
}

type PoloniexAccessStruct struct {
	Username string
	APIKey   string
	Secret   string

	LastStatsUpdate time.Time
}

func (p *PoloniexAccessCache) shouldRecordStats(username string) bool {
	p.Lock()
	defer p.Unlock()
	if v, ok := p.Cache[username]; ok {
		if time.Since(v.LastStatsUpdate).Seconds() > 60 {
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

func (s *State) getAccessAndSecret(username string) (string, string, error) {
	s.updateCache()
	s.poloniexCache.RLock()
	c, ok := s.poloniexCache.Cache[username]
	s.poloniexCache.RUnlock()
	if ok {
		return c.APIKey, c.Secret, nil
	}

	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return "", "", err
	}

	ck := u.GetCipherKey(s.CipherKey)
	accessKey, err := u.PoloniexKeys.DecryptAPIKeyString(ck)
	if err != nil {
		return "", "", err
	}

	secretKey, err := u.PoloniexKeys.DecryptAPISecretString(ck)
	if err != nil {
		return "", "", err
	}
	s.poloniexCache.Lock()
	var tmp PoloniexAccessStruct
	tmp.Username = username
	tmp.APIKey = accessKey
	tmp.Secret = secretKey
	s.poloniexCache.Cache[username] = tmp
	s.poloniexCache.Unlock()

	return accessKey, secretKey, nil
}

func (s *State) PoloniexGetBalances(username string) (*poloniex.PoloniexBalance, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}
	return s.PoloniexAPI.GetBalances(accessKey, secretKey)
}

// PoloniexGetAvailableBalances returns:
//		map[string] :: key = "exchange", "lending", "margin"
//		|-->	map[string] :: key = currency
func (s *State) PoloniexGetAvailableBalances(username string) (map[string]map[string]float64, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}

	return s.PoloniexAPI.GetAvilableBalances(accessKey, secretKey)
}

// PoloniexCreateLoanOffer creates a loan offer
func (s *State) PoloniexCreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, username string) (int64, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return 0, err
	}

	return s.PoloniexAPI.CreateLoanOffer(currency, amount, rate, duration, autoRenew, accessKey, secretKey)
}

// PoloniexGetInactiveLoans returns your current loans that are not taken
func (s *State) PoloniexGetInactiveLoans(username string) (map[string][]poloniex.PoloniexLoanOffer, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}

	return s.PoloniexAPI.GetOpenLoanOffers(accessKey, secretKey)
}

// PoloniexGetActiveLoans returns your current loans that are taken
func (s *State) PoloniexGetActiveLoans(username string) (*poloniex.PoloniexActiveLoans, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}

	return s.PoloniexAPI.GetActiveLoans(accessKey, secretKey)
}

func (s *State) PoloniexCancelLoanOffer(currency string, orderNumber int64, username string) (bool, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return false, err
	}

	return s.PoloniexAPI.CancelLoanOffer(currency, orderNumber, accessKey, secretKey)
}

func (s *State) PoloniexGetLoanOrders(currency string) (*poloniex.PoloniexLoanOrders, error) {
	return s.PoloniexAPI.GetLoanOrders(currency)
}

func (s *State) PoloniexSingleAuthenticatedTradeHistory(currency, username, start, end string) (resp poloniex.PoloniexAuthenticatedTradeHistoryResponse, err error) {
	if currency == "all" {
		return resp, fmt.Errorf("Cannot be 'all'")
	}
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}

	respNonCast, err := s.PoloniexAPI.GetAuthenticatedTradeHistory(currency, start, end, accessKey, secretKey)
	resp = respNonCast.(poloniex.PoloniexAuthenticatedTradeHistoryResponse)
	return
}

func (s *State) PoloniexAllAuthenticatedTradeHistory(username, start, end string) (resp poloniex.PoloniexAuthenticatedTradeHistoryAll, err error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}

	respNonCast, err := s.PoloniexAPI.GetAuthenticatedTradeHistory("", start, end, accessKey, secretKey)
	resp = respNonCast.(poloniex.PoloniexAuthenticatedTradeHistoryAll)
	return
}

func (s *State) PoloniexAuthenticatedLendingHistory(username, start, end string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return resp, err
	}

	resp, err = s.PoloniexAPI.GetAuthenticatedLendingHistory(start, end, accessKey, secretKey)
	return
}
