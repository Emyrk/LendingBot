package payment

import (
	"gopkg.in/mgo.v2"
)

type CoinbasePaymentCode struct {
	Code string `json:"code" bson:"_id"`
}

func (p *PaymentDatabase) InsertCoinbasePaymentCode(code string) error {
	// Insert PaymentCode to DB
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_CoinbaseCode)
	if err != nil {
		return results, fmt.Errorf("InsertCoinbasePaymentCode: getcol: %s", err)
	}
	defer s.Close()

	return c.Insert(CoinbasePaymentCode{
		Code: code,
	})
}

func (p *PaymentDatabase) PaymentCoinbaseCodeExists(code string) bool {
	// Does code exists?
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_CoinbaseCode)
	if err != nil {
		return results, fmt.Errorf("PaymentCoinbaseCodeExists: getcol: %s", err)
	}
	defer s.Close()

	err = c.FindId(code)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return true
	}
	return false
}

type HODLZONEPaymentCode struct {
	Code string `json:"code" bson:"_id"`
}

func (p *PaymentDatabase) InsertHODLZONEPaymentCode(code string) {
	// Insert PaymentCode to DB
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_HODLZONECode)
	if err != nil {
		return results, fmt.Errorf("InsertHODLZONEPaymentCode: getcol: %s", err)
	}
	defer s.Close()

	return c.Insert(HODLZONEPaymentCode{
		Code: code,
	})
}

func (p *PaymentDatabase) PaymentHODLZONECodeExists(code string) bool {
	// Does code exists?
	// use CoinbasePaymentCode

	s, c, err := p.db.GetCollection(mongo.C_HODLZONECode)
	if err != nil {
		return results, fmt.Errorf("PaymentHODLZONECodeExists: getcol: %s", err)
	}
	defer s.Close()

	err = c.FindId(code)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return true
	}
	return false
}
