package core

import (
	"crypto/rand"
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/poloniex"
	"github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

type State struct {
	UserDB      *userdb.UserDatabase
	PoloniexAPI *poloniex.Poloniex
	CipherKey   [32]byte
	JWTSecret   [32]byte
}

func NewState() *State {
	s := new(State)
	s.UserDB = userdb.NewMapUserDatabase()
	s.PoloniexAPI = poloniex.StartPoloniex()
	ck := make([]byte, 32)
	copy(s.CipherKey[:], ck[:])

	jck := make([]byte, 32)
	_, err := rand.Read(jck)
	if err != nil {
		panic(fmt.Sprintf("Could not generate JWT Siging Key %s", err.Error()))
	}
	copy(s.JWTSecret[:], jck[:])

	return s
}

func (s *State) SetUserMinimumLoan(username string, minimumAmt float64) error {
	u, err := s.UserDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.MiniumLend = minimumAmt
	return s.UserDB.PutUser(u)
}

func (s *State) NewUser(username string, password string) error {
	ou, err := s.UserDB.FetchUserIfFound(username)
	if err != nil {
		if ou != nil {
			return fmt.Errorf("username already exists")
		} else {
			return fmt.Errorf("could not check if username exists: %s", err.Error())
		}
	}

	u, err := userdb.NewUser(username, password)
	if err != nil {
		return err
	}
	return s.UserDB.PutUser(u)
}

func (s *State) SetUserKeys(username string, acessKey string, secretKey string) error {
	u, err := s.UserDB.FetchUserIfFound(username)
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
