package coinbase

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	CheckoutAPIURL = "https://api.coinbase.com/v2/checkouts"
)

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
}

type CheckoutResponse struct {
	Data struct {
		ID          string `json:"id"`
		Code        string `json:"code"`
		Status      string `json:"status"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Amount      struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Metadata *MetaDataField `json:"metadata"`
	} `json:"data"`
}

func CreatePayment(username string) ([]byte, error) {
	// https://api.coinbase.com/v2/checkouts
	client := http.Client{}

	meta := new(MetaDataField)
	meta.Username = username
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respData, nil
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
