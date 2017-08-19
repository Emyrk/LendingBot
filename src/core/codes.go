package core

import (
	"crypto/rand"
	"encoding/hex"
)

func (s *State) GenerateHODLZONECode() (string, error) {
	code := make([]byte, 10)
	_, err := rand.Read(code)
	if err != nil {
		return "", err
	}

	str := hex.EncodeToString(code)
	return str, s.paymentDB.InsertHODLZONEPaymentCode(str)
}

func (s *State) HODLZONECodeExists(code string) (bool, error) {
	return s.paymentDB.PaymentHODLZONECodeExists(code)
}

func (s *State) InsertCoinbaseCode(code string) error {
	return s.paymentDB.InsertCoinbasePaymentCode(code)
}

func (s *State) CoinbaseCodeExists(code string) (bool, error) {
	return s.paymentDB.PaymentCoinbaseCodeExists(code)
}
