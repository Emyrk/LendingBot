package coinbase

import (
	"encoding/json"
	"time"
)

type CoinbaseNotification struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Resource struct {
			ID          string `json:"id"`
			Code        string `json:"code"`
			Type        string `json:"type"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Amount      struct {
				Amount   string `json:"amount"`
				Currency string `json:"currency"`
			} `json:"amount"`
			ReceiptURL    string `json:"receipt_url"`
			Resource      string `json:"resource"`
			ResourcePath  string `json:"resource_path"`
			Status        string `json:"status"`
			BitcoinAmount struct {
				Amount   string `json:"amount"`
				Currency string `json:"currency"`
			} `json:"bitcoin_amount"`
			PayoutAmount   interface{} `json:"payout_amount"`
			BitcoinAddress string      `json:"bitcoin_address"`
			RefundAddress  interface{} `json:"refund_address"`
			BitcoinURI     string      `json:"bitcoin_uri"`
			PaidAt         time.Time   `json:"paid_at"`
			MispaidAt      interface{} `json:"mispaid_at"`
			ExpiresAt      time.Time   `json:"expires_at"`
			Metadata       struct {
			} `json:"metadata"`
			CreatedAt    time.Time   `json:"created_at"`
			UpdatedAt    time.Time   `json:"updated_at"`
			CustomerInfo interface{} `json:"customer_info"`
			Transaction  struct {
				ID           string `json:"id"`
				Resource     string `json:"resource"`
				ResourcePath string `json:"resource_path"`
			} `json:"transaction"`
			Mispayments []interface{} `json:"mispayments"`
			Refunds     []interface{} `json:"refunds"`
		} `json:"resource"`
	} `json:"data"`
	User struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"user"`
	Account struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"account"`
	DeliveryAttempts int       `json:"delivery_attempts"`
	CreatedAt        time.Time `json:"created_at"`
	Resource         string    `json:"resource"`
	ResourcePath     string    `json:"resource_path"`
}

func IncomingNotification(data []byte) (*CoinbaseNotification, error) {
	n := new(CoinbaseNotification)
	err := json.Unmarshal(data, n)
	if err != nil {
		return nil, err
	}
	return n, nil
}
