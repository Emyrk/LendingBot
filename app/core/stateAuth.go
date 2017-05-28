package core

import (
	"encoding/base64"
)

func (s *State) Add2FA(username string, password string) (qr64 string, err error) {
	qrRaw, err := s.UserDB.Add2FA(username, password)
	if err != nil {
		return nil, err
	}

	qr64 = base64.StdEncoding.EncodeToString(qrRaw)
	return
}

func (s *State) Enable2FA(username string, password string, token string, enabled bool) error {
	return s.UserDB.Enable2FA(username, password, token, enabled)
}
