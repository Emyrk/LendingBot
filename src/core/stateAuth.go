package core

import (
	"encoding/base64"

	"github.com/Emyrk/LendingBot/src/core/userdb"
)

func (s *State) VerifyEmail(verifyString string) error {
	return s.UserDB.VerifyEmail(verifyString)
}

func (s *State) AuthenticateUser(username string, password string) (bool, *userdb.User, error) {
	return s.UserDB.AuthenticateUser(username, password, "")
}

func (s *State) AuthenticateUser2FA(username string, password string, token string) (bool, *userdb.User, error) {
	return s.UserDB.AuthenticateUser(username, password, token)
}

func (s *State) Add2FA(username string, password string) (qr64 string, err error) {
	qrRaw, err := s.UserDB.Add2FA(username, password)
	if err != nil {
		return "", err
	}

	qr64 = base64.StdEncoding.EncodeToString(qrRaw)
	return
}

func (s *State) Enable2FA(username string, password string, token string, enabled bool) error {
	return s.UserDB.Enable2FA(username, password, token, enabled)
}
