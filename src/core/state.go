package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core/cryption"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

type State struct {
	userDB        *userdb.UserDatabase
	userStatistic *userdb.UserStatisticsDB
	PoloniexAPI   poloniex.IPoloniex
	CipherKey     [32]byte
	JWTSecret     [32]byte

	// Poloniex Cache
	poloniexCache *PoloniexAccessCache
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
		s.userDB = userdb.NewMapUserDatabase()
	} else {
		v := os.Getenv("USER_DB")
		if len(v) == 0 {
			v = "UserDatabase.db"
		}
		s.userDB = userdb.NewBoltUserDatabase(v)
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

	if withMap {
		s.userStatistic, err = userdb.NewUserStatisticsMapDB()
	} else {
		s.userStatistic, err = userdb.NewUserStatisticsDB()
	}
	if err != nil {
		panic(fmt.Sprintf("Could create user statistic database %s", err.Error()))
	}

	s.poloniexCache = NewPoloniexAccessCache()

	return s
}

func (s *State) Close() error {
	return s.userDB.Close()
}

func (s *State) SetUserMinimumLoan(username string, minimumAmt float64) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.MiniumLend = minimumAmt
	return s.userDB.PutUser(u)
}

func (s *State) NewUser(username string, password string) error {
	ou, err := s.userDB.FetchUser(username)
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

	err = s.userDB.PutVerifystring(userdb.GetUsernameHash(username), u.VerifyString)
	if err != nil {
		return err
	}

	return s.userDB.PutUser(u)
}

func (s *State) SetUserKeys(username string, acessKey string, secretKey string) error {
	u, err := s.userDB.FetchUserIfFound(username)
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
	return s.userDB.PutUser(u)
}

func (s *State) GetStatistics(username string, dayRange int) ([][]*userdb.UserStatistic, error) {
	return s.userStatistic.GetStatistics(username, dayRange)
}

func (s *State) EnableUserLending(username string, enabled bool) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.PoloniexEnabled = enabled
	return s.userDB.PutUser(u)
}

func (s *State) FetchUser(username string) (*userdb.User, error) {
	return s.userDB.FetchUser(username)
}

func (s *State) FetchAllUsers() ([]userdb.User, error) {
	return s.userDB.FetchAllUsers()
}

func (s *State) RecordStatistics(stats *userdb.UserStatistic) error {
	if !s.poloniexCache.shouldRecordStats(stats.Username) {
		return nil
	}
	err := s.userStatistic.RecordData(stats)
	if err != nil {
		return err
	}

	return nil
}

func (s *State) GetNewJWTOTP(username string) (string, error) {
	return s.setupNewJWTOTP(username, cryption.JWT_EXPIRY_TIME_NEW_PASS)
}

func (s *State) setupNewJWTOTP(username string, t time.Duration) (string, error) {
	tokenString, err := cryption.NewJWTString(username, s.JWTSecret, t)
	if err != nil {
		return "", err
	}
	sig, err := cryption.GetJWTSignature(tokenString)
	if err != nil {
		return "", err
	}

	var b [32]byte
	copy(b[:], sig)
	if err = s.userDB.UpdateJWTOTP(username, b); err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *State) CompareClearJWTOTP(tokenString string) bool {
	token, err := cryption.VerifyJWT(tokenString, s.JWTSecret)
	if err != nil {
		fmt.Printf("Error comparing JWT for pass reset: %s\n", err.Error())
		return false
	}

	email, ok := token.Claims().Get("email").(string)
	if !ok {
		fmt.Printf("Error Retrieving email for pass reset: %s\n", err.Error())
		return false
	}

	b, ok := s.userDB.GetJWTOTP(email)
	if !ok {
		fmt.Printf("Error with getting Token for user for pass reset: %s\n", err.Error())
		return false
	}

	tokenSig, err := cryption.GetJWTSignature(tokenString)
	if err != nil {
		fmt.Printf("Error retrieving sig for JWT for pass reset: %s\n", err)
		return false
	}
	userSig := hex.EncodeToString(b[:])

	return userSig == tokenSig
}
