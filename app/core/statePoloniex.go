package core

import (
	//"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/poloniex"
	//"github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

func (s *State) PoloniexGetBalances(username string) (*poloniex.PoloniexBalance, error) {
	u, err := s.UserDB.FetchUser(username)
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
