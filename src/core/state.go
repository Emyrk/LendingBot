package core

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

type State struct {
	UserDB      *userdb.UserDatabase
	PoloniexAPI poloniex.IPoloniex
	CipherKey   [32]byte
	JWTSecret   [32]byte
}

func NewFakePoloniexState() *State {
	return newState(true, true)
}

func NewState() *State {
	return newState(false, false)
}

func NewStateWithMap() *State {
	return newState(true, false)
}

func newState(withMap bool, fakePolo bool) *State {
	s := new(State)
	if withMap {
		s.UserDB = userdb.NewMapUserDatabase()
	} else {
		v := os.Getenv("USER_DB")
		if len(v) == 0 {
			v = "UserDatabase.db"
		}
		s.UserDB = userdb.NewBoltUserDatabase(v)
	}

	if fakePolo {
		s.PoloniexAPI = poloniex.StartFakePoloniex()
	} else {
		s.PoloniexAPI = poloniex.StartPoloniex()
	}

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

func (s *State) Close() error {
	return s.UserDB.Close()
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
	ou, err := s.UserDB.FetchUser(username)
	if err != nil {
		return fmt.Errorf("could not check if username exists: %s", err.Error())
	}

	if ou != nil {
		return fmt.Errorf("username already exists")
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
	// TODO: Make this a separate action
	u.PoloniexEnabled = true
	return s.UserDB.PutUser(u)
}

func (s *State) EnableUserLending(username string, enabled bool) error {
	u, err := s.UserDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.PoloniexEnabled = enabled
	return s.UserDB.PutUser(u)
}

func (s *State) FetchUser(username string) (*userdb.User, error) {
	return s.UserDB.FetchUser(username)
}

func (s *State) FetchAllUsers() ([]userdb.User, error) {
	return s.UserDB.FetchAllUsers()
}
