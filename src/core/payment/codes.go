package payment

type CoinbasePaymentCode struct {
	Code string `json:"code" bson:"code"`
}

func (p *PaymentDatabase) InsertCoinbasePaymentCode(code string) {
	// Insert PaymentCode to DB
	// use CoinbasePaymentCode
}

func (p *PaymentDatabase) PaymentCoinbaseCodeExists(code string) bool {
	// Does code exists?
	// use CoinbasePaymentCode
}

type HODLZONEPaymentCode struct {
	Code string `json:"code" bson:"code"`
}

func (p *PaymentDatabase) InsertHODLZONEPaymentCode(code string) {
	// Insert PaymentCode to DB
	// use CoinbasePaymentCode
}

func (p *PaymentDatabase) PaymentHODLZONECodeExists(code string) bool {
	// Does code exists?
	// use CoinbasePaymentCode
}
