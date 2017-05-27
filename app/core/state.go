package core

import (
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/poloniex"
	"github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

type State struct {
	UserDB      *userdb.UserDatabase
	PoloniexAPI *poloniex.Poloniex
	CipherKey   [32]byte
}

func NewState() *State {
	s := new(State)
	s.UserDB = userdb.NewMapUserDatabase()
	s.PoloniexAPI = poloniex.StartPoloniex()
	ck := make([]byte, 32)
	copy(s.CipherKey[:], ck[:])

	return s
}

func (s *State) SetUserMinimumLoan(username string, minimumAmt float64) error {
	u, err := s.UserDB.FetchUser(username)
	if err != nil {
		return err
	}

	u.MiniumLend = minimumAmt
	return s.UserDB.PutUser(u)
}

func (s *State) NewUser(username string, password string) error {
	_, err := s.UserDB.FetchUser(username)
	if err == nil {
		return fmt.Errorf("username already exists")
	}

	u, err := userdb.NewUser(username, password)
	if err != nil {
		return err
	}
	return s.UserDB.PutUser(u)
}

func (s *State) SetUserKeys(username string, acessKey string, secretKey string) error {
	u, err := s.UserDB.FetchUser(username)
	if err != nil {
		return err
	}

	pk, err := userdb.NewPoloniexKeys(acessKey, secretKey, u.GetCipherKey(s.CipherKey))
	if err != nil {
		return err
	}

	u.PoloniexKeys = pk
	return s.UserDB.PutUser(u)
}

func (s *State) FetchUser(username string) (*userdb.User, error) {
	return s.UserDB.FetchUser(username)
}

func (s *State) AuthenticateUser(username string, password string) (bool, *userdb.User, error) {
	return s.UserDB.AuthenticateUser(username, password)
}
