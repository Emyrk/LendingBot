package coinbase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/payment"
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
		Custom   string `json:"custom"`
		Username string `json:"username"`
		Version  int    `json:"version"`
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

func NotificationPairToPaid(parent *CoinbaseNotification, pay *CoinbasePaymentNotification) (*payment.Paid, error) {
	var err error

	p := new(payment.Paid)
	p.Username = pay.Metadata.Username
	p.ContactUsername = pay.CustomerInfo.Email

	p.PaymentDate = pay.PaidAt
	p.PaymentCreatedAt = pay.CreatedAt
	p.PaymentExpiresAt = pay.ExpiresAt

	if pay.TotalAmountReceived.Currency != "BTC" {
		return nil, fmt.Errorf("payment currency found is not BTC, but %s", pay.TotalAmountReceived.Currency)
	}
	p.BTCPaid = pay.TotalAmountReceived.Amount

	p.CoinbaseNotificationID = parent.ID
	p.CoinbaseUserID = parent.User.ID
	p.CoinbaseAccountID = parent.Account.ID
	p.NotificationCreatedAt = parent.CreatedAt
	p.NotificationDelivedAt = time.Now()
	p.DeliveryAttempts = parent.DeliveryAttempts

	p.CoinbasePaymentID = pay.ID
	p.ReceiptUrl = pay.ReceiptURL
	p.Code = pay.Code
	p.BTCAddress = pay.BitcoinAddress
	p.RefundAddress = pay.RefundAddress

	return p, nil
}

type CoinbaseWatcher struct {
	State *core.State
}

func (h *CoinbaseWatcher) IncomingNotification(data []byte) error {
	n := new(CoinbaseNotification)
	// LOG RAW
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}

	switch n.Type {
	case OrderPaid:
		pay := new(CoinbasePaymentNotification)
		err := json.Unmarshal(n.Data, pay)
		if err != nil {
			return err
		}

		paid, err := NotificationPairToPaid(n, pay)
		if err != nil {
			return err
		}

		paid.RawData = n.Data

		return h.State.MakePayment(paid.Username, *paid)
		// payment.TotalAmountReceived.Currency
	}
	return nil
}
