package core

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	//"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Println

func (s *State) getAccessAndSecret(username string) (string, string, error) {
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

// PoloniecGetInactiveLoans returns your current loans that are not taken
func (s *State) PoloniexGetInactiveLoans(username string) (map[string][]poloniex.PoloniexLoanOffer, error) {
	accessKey, secretKey, err := s.getAccessAndSecret(username)
	if err != nil {
		return nil, err
	}

	return s.PoloniexAPI.GetOpenLoanOffers(accessKey, secretKey)
}

// PoloniecGetActiveLoans returns your current loans that are taken
func (s *State) PoloniecGetActiveLoans(username string) (*poloniex.PoloniexActiveLoans, error) {
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

func (s *State) PoloniecGetLoanOrders(currency string) (*poloniex.PoloniexLoanOrders, error) {
	return s.PoloniexAPI.GetLoanOrders(currency)
}
