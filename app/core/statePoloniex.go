package core

import (
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/poloniex"
	//"github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

var _ = fmt.Println

func (s *State) PoloniexGetBalances(username string) (*poloniex.PoloniexBalance, error) {
	u, err := s.UserDB.FetchUserIfFound(username)
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

	return s.PoloniexAPI.GetBalances(accessKey, secretKey)
}

// PoloniexGetAvailableBalances returns:
//		map[string] :: key = "exchange", "lending", "margin"
//		|-->	map[string] :: key = currency
func (s *State) PoloniexGetAvailableBalances(username string) (map[string]map[string]float64, error) {
	u, err := s.UserDB.FetchUserIfFound(username)
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

	return s.PoloniexAPI.GetAvilableBalances(accessKey, secretKey)
}

// PoloniexCreateLoanOffer creates a loan offer
func (s *State) PoloniexCreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, username string) (int64, error) {
	u, err := s.UserDB.FetchUserIfFound(username)
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

	return s.PoloniexAPI.CreateLoanOffer(currency, amount, rate, duration, autoRenew, accessKey, secretKey)
}

// PoloniecGetOpenLoanOffers returns your current loans that are not taken
func (s *State) PoloniecGetOpenLoanOffers(username string) (map[string][]poloniex.PoloniexLoanOffer, error) {
	u, err := s.UserDB.FetchUserIfFound(username)
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

	return s.PoloniexAPI.GetOpenLoanOffers(accessKey, secretKey)
}

// PoloniecGetActiveLoans returns your current loans that are taken
func (s *State) PoloniecGetActiveLoans(username string) (poloniex.PoloniexActiveLoans, error) {
	u, err := s.UserDB.FetchUserIfFound(username)
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

	return s.PoloniexAPI.GetActiveLoans(accessKey, secretKey)
}

func (s *State) PoloniecGetLoanOrders(currency string) (poloniex.PoloniexLoanOrders, error) {
	return s.PoloniexAPI.GetLoanOrders(currency)
}
