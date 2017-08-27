package payment

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CoinbasePaymentCode struct {
	Code string `json:"code" bson:"_id"`
}

func (p *PaymentDatabase) InsertCoinbasePaymentCode(code string) error {
	// Insert PaymentCode to DB
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_CoinbaseCode)
	if err != nil {
		return fmt.Errorf("InsertCoinbasePaymentCode: getcol: %s", err)
	}
	defer s.Close()

	return c.Insert(CoinbasePaymentCode{
		Code: code,
	})
}

func (p *PaymentDatabase) PaymentCoinbaseCodeExists(code string) (bool, error) {
	// Does code exists?
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_CoinbaseCode)
	if err != nil {
		return false, fmt.Errorf("PaymentCoinbaseCodeExists: getcol: %s", err)
	}
	defer s.Close()

	var result bson.M
	err = c.FindId(code).Limit(1).One(&result)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return false, nil
	}

	return true, err
}

type HODLZONEPaymentCode struct {
	Code string `json:"code" bson:"_id"`
}

func (p *PaymentDatabase) InsertHODLZONEPaymentCode(code string) error {
	// Insert PaymentCode to DB
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_HODLZONECode)
	if err != nil {
		return fmt.Errorf("InsertHODLZONEPaymentCode: getcol: %s", err)
	}
	defer s.Close()

	return c.Insert(HODLZONEPaymentCode{
		Code: code,
	})
}

func (p *PaymentDatabase) PaymentHODLZONECodeExists(code string) (bool, error) {
	// Does code exists?
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_HODLZONECode)
	if err != nil {
		return false, fmt.Errorf("PaymentHODLZONECodeExists: getcol: %s", err)
	}
	defer s.Close()

	var result bson.M
	err = c.FindId(code).Limit(1).One(&result)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
