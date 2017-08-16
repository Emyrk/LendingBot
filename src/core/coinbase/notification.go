package coinbase

import (
	"encoding/json"
	"fmt"
	"time"
)

var _ = fmt.Println

// Coinbase Notification Types
const (
	OrderPaid = "wallet:orders:paid"
)

type CoinbaseNotification struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
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

type CoinbasePaymentNotification struct {
	ID          string      `json:"id"`
	Code        string      `json:"code"`
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description interface{} `json:"description"`
	Amount      struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	ReceiptURL    string `json:"receipt_url"`
	Resource      string `json:"resource"`
	ResourcePath  string `json:"resource_path"`
	Status        string `json:"status"`
	Overpaid      bool   `json:"overpaid"`
	BitcoinAmount struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"bitcoin_amount"`
	TotalAmountReceived struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"total_amount_received"`
	PayoutAmount     json.RawMessage `json:"payout_amount"`
	BitcoinAddress   string          `json:"bitcoin_address"`
	RefundAddress    string          `json:"refund_address"`
	BitcoinURI       string          `json:"bitcoin_uri"`
	NotificationsURL string          `json:"notifications_url"`
	PaidAt           time.Time       `json:"paid_at"`
	MispaidAt        time.Time       `json:"mispaid_at"`
	ExpiresAt        time.Time       `json:"expires_at"`
	Metadata         struct {
		Custom string `json:"custom"`
	} `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CustomerInfo struct {
		Name        interface{} `json:"name"`
		Email       string      `json:"email"`
		PhoneNumber interface{} `json:"phone_number"`
	} `json:"customer_info"`
	Transaction struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"transaction"`
	Mispayments []interface{} `json:"mispayments"`
	Refunds     []interface{} `json:"refunds"`
}

type CoinbaseWatcher struct {
}

func (h *CoinbaseWatcher) IncomingNotification(data []byte) (*CoinbaseNotification, error) {
	n := new(CoinbaseNotification)
	// LOG RAW
	err := json.Unmarshal(data, n)
	if err != nil {
		return nil, err
	}

	switch n.Type {
	case OrderPaid:
		payment := new(CoinbasePaymentNotification)
		err := json.Unmarshal(n.Data, payment)
		if err != nil {
			return nil, err
		}
		// TODO: Handle Payment

	}
	// LOG MARSHALED
	return n, nil
}

func (h *CoinbaseWatcher) HandlePayment(parent *CoinbaseNotification, payment *CoinbasePaymentNotification) {

}
