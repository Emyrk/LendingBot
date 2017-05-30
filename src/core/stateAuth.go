package core

import (
	"encoding/base64"

	"github.com/Emyrk/LendingBot/src/core/userdb"
)

func (s *State) VerifyEmail(email, verifyString string) error {
	return s.userDB.VerifyEmail(email, verifyString)
}

func (s *State) AuthenticateUser(username string, password string) (bool, *userdb.User, error) {
	return s.userDB.AuthenticateUser(username, password, "")
}

func (s *State) AuthenticateUser2FA(username string, password string, token string) (bool, *userdb.User, error) {
	return s.userDB.AuthenticateUser(username, password, token)
}

func (s *State) Add2FA(username string, password string) (qr64 string, err error) {
	qrRaw, err := s.userDB.Add2FA(username, password)
	if err != nil {
		return "", err
	}

	qr64 = base64.StdEncoding.EncodeToString(qrRaw)
	return
}

func (s *State) Enable2FA(username string, password string, token string, enabled bool) error {
	return s.userDB.Enable2FA(username, password, token, enabled)
}
