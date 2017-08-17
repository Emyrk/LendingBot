package coinbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/revel/revel"
)

var _ = fmt.Println

const (
	CheckoutAPIURL = "https://api.coinbase.com/v2/checkouts"
)

func InitCoinbaseAPI() {
	time.Sleep(1 * time.Second)
	if revel.DevMode {
		return
	}
	coinbase_access_key = os.Getenv("COINBASE_ACCESS_KEY")
	coinbase_secret_key = os.Getenv("COINBASE_SECRET_KEY")

	if coinbase_access_key == "" || coinbase_secret_key == "" {
		panic("No coinbase API keys given")
	}
}

var coinbase_access_key string
var coinbase_secret_key string

// https://developers.coinbase.com/api/v2?shell#create-checkout
type CheckoutOptions struct {
	Amount                 string      `json:"amount"`
	Currency               string      `json:"currency"`
	Name                   string      `json:"name"`
	Description            string      `json:"description"`
	Type                   string      `json:"type"`
	Style                  string      `json:"style"`
	CustomerDefinedAmount  bool        `json:"customer_defined_amount"`
	AmountPresets          []string    `json:"amount_presets"`
	SuccessURL             string      `json:"success_url"`
	CancelURL              string      `json:"cancel_url"`
	AutoRedirect           bool        `json:"auto_redirect"`
	CollectShippingAddress bool        `json:"collect_shipping_address"`
	CollectEmail           bool        `json:"collect_email"`
	CollectPhoneNumber     bool        `json:"collect_phone_number"`
	CollectCountry         bool        `json:"collect_country"`
	Metadata               interface{} `json:"metadata"`
}

type MetaDataField struct {
	Username string `json:"username"`
	Version  int    `json:"version"`
}

type PaymentButton struct {
	Data struct {
		ID          string `json:"id"`
		EmbedCode   string `json:"embed_code"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Amount      struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Style                 string `json:"style"`
		CustomerDefinedAmount bool   `json:"customer_defined_amount"`
		AmountPresets         []struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount_presets"`
		CallbackURL            interface{} `json:"callback_url"`
		SuccessURL             string      `json:"success_url"`
		CancelURL              string      `json:"cancel_url"`
		AutoRedirect           bool        `json:"auto_redirect"`
		NotificationsURL       interface{} `json:"notifications_url"`
		CollectShippingAddress bool        `json:"collect_shipping_address"`
		CollectEmail           bool        `json:"collect_email"`
		CollectPhoneNumber     bool        `json:"collect_phone_number"`
		CollectCountry         bool        `json:"collect_country"`
		Metadata               struct {
			Username string `json:"username"`
			Version  int    `json:"version"`
		} `json:"metadata"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Resource     string    `json:"resource"`
		ResourcePath string    `json:"resource_path"`
	} `json:"data"`
}

func CreatePayment(username string) (*PaymentButton, error) {
	// https://api.coinbase.com/v2/checkouts
	client := http.Client{}

	meta := new(MetaDataField)
	meta.Username = username
	meta.Version = 1
	o := NewDefaultCheckoutOptions()
	o.Metadata = meta

	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", CheckoutAPIURL, buf)
	if err != nil {
		return nil, err
	}

	// TODO: TEST AUTHENTICATION
	api := apiKeyAuth(coinbase_access_key, coinbase_secret_key)

	err = api.authenticate(req, data)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	button := new(PaymentButton)
	err = json.Unmarshal(respData, button)
	if err != nil {
		return nil, err
	}
	return button, nil
}

/*
amount	string	Required	Order amount (price)
currency	string	Required	Order amount’s currency
name	string	Required	Name of the order
description	string	Optional	More detailed description of the checkout order
type	string	Optional	Checkout’s order type. Available values: order (default), donation
style	string	Optional	Style of a payment button. Currently available values: buy_now_large, buy_now_small, donation_large, donation_small ,custom_large, custom_small
customer_defined_amount	boolean	Optional	Allow customer to define the amount they are paying. This is most commonly used with donations
amount_presets	array	Optional	Allow customer to select one of the predefined amount values. Input value must be an array of number values. Preset values will inherit currency from currency argument
success_url	string	Optional	URL to which the customer is redirected after successful payment
cancel_url	string	Optional	URL to which the customer is redirected after they have canceled a payment
notifications_url	string	Optional	Checkout specific notification URL
auto_redirect	boolean	Optional	Auto-redirect users to success or cancel url after payment
collect_shipping_address	boolean	Optional	Collect shipping address from customer (not for use with inline iframes)
collect_email	boolean	Optional	Collect email address from customer (not for use with inline iframes)
collect_phone_number	boolean	Optional	Collect phone number from customer (not for use with inline iframes)
collect_country	boolean	Optional	Collect country from customer (not for use with inline iframes)
metadata	hash	Optional	Developer defined key value pairs. Read more.
*/
func NewDefaultCheckoutOptions() *CheckoutOptions {
	options := new(CheckoutOptions)

	options.Amount = "0.01"
	options.Currency = "BTC"
	options.Name = "Hodl.zone Bot Credits"
	options.Description = "Purchasing credits will enable the lending bot to start making loans on your behalf. It will do so until the bot runs out of credits. These credits are non-refundable."
	options.Type = "order"
	options.Style = "buy_now_small"
	options.CustomerDefinedAmount = true
	options.AmountPresets = []string{"0.005", "0.01", "0.02"}
	options.SuccessURL = ""
	options.CancelURL = ""
	options.AutoRedirect = false
	options.CollectCountry = false
	options.CollectEmail = true
	options.CollectPhoneNumber = false
	options.CollectShippingAddress = false

	return options
}
