// TODO: Public: Orderbook, Trades, Lends, Symbols, Symbols Details
// TODO: Authenticated: New deposit, New order, Multiple new orders, Cancel order, Cancel multiple orders, Cancel all active orders, Replace order, Order status, Active Orders, Active Positions, Claim position, Past trades, Offer status, Active Swaps used in a margin position, Balance history, Close swap, Account informations, Margin informations

package bitfinex

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var LendingCurrencies = []string{"BTC", "ETH", "ETC", "ZEC", "XMR", "LTC", "DSH", "IOT", "EOS", "USD"}

const (
	// APIURL points to Bitfinex API URL, found at https://www.bitfinex.com/pages/API
	APIURL = "https://api.bitfinex.com"
	// LEND ...
	LEND = "lend"
	// BORROW ...
	BORROW = "borrow"
)

// API structure stores Bitfinex API credentials
type API struct {
	APIKey    string
	APISecret string
}

// ErrorMessage ...
type ErrorMessage struct {
	Message string `json:"message"` // Returned only on error
}

type SummaryVolume struct {
	Currency string `json:"curr"`
	Volume   string `json:"vol"`
}
type SummaryProfit struct {
	Currency string `json:"curr"`
	Volume   string `json:"amount"`
}
type Summary struct {
	TradeVolume   SummaryVolume `json:"trade_vol_30d"`
	FundingProfit SummaryProfit `json:"funding_profit_30d"`
	MakerFee      string        `json:"maker_fee"`
	TakerFee      string        `json:"taker_fee"`
}

type V2FundingTicker struct {
	Symbol             string  `json:"symbol"`
	FRR                float64 `json:"frr"`
	Bid                float64 `json:"bid"`
	BidSize            float64 `json:"bidsize"`
	BidPeriod          float64 `json:"bidperiod"`
	Ask                float64 `json:"ask"`
	AskSize            float64 `json:"asksize"`
	AskPeriod          float64 `json:"askperiod"`
	DailyChange        float64 `json:"dailychange"`
	DailyChangePercent float64 `json:"dailchangeperc"`
	LastPrice          float64 `json:"lastprice"`
	Volume             float64 `json:"volume"`
	High               float64 `json:"high"`
	Low                float64 `json:"low"`
}

type V2Ticker struct {
	Symbol             string  `json:"symbol"`
	Bid                float64 `json:"bid"`
	BidSize            float64 `json:"bidsize"`
	Ask                float64 `json:"ask"`
	AskSize            float64 `json:"asksize"`
	DailyChange        float64 `json:"dailychange"`
	DailyChangePercent float64 `json:"dailchangeperc"`
	LastPrice          float64 `json:"lastprice"`
	Volume             float64 `json:"volume"`
	High               float64 `json:"high"`
	Low                float64 `json:"low"`
}

// Ticker ...
type Ticker struct {
	Mid       float64 `json:"mid,string"`        // mid (price): (bid + ask) / 2
	Bid       float64 `json:"bid,string"`        // bid (price): Innermost bid.
	Ask       float64 `json:"ask,string"`        // ask (price): Innermost ask.
	LastPrice float64 `json:"last_price,string"` // last_price (price) The price at which the last order executed.
	Low       float64 `json:"low,string"`        // low (price): Lowest trade price of the last 24 hours
	High      float64 `json:"high,string"`       // high (price): Highest trade price of the last 24 hours
	Volume    float64 `json:"volume,string"`     // volume (price): Trading volume of the last 24 hours
	Timestamp float64 `json:"timestamp,string"`  // timestamp (time) The timestamp at which this information was valid.
}

// Stats ...
type Stats []Stat

// Stat ...
type Stat struct {
	Period int     `json:"period"`        // period (integer), period covered in days
	Volume float64 `json:"volume,string"` // volume (price)
}

// Lendbook ...
type Lendbook struct {
	Bids []LendbookOffer // bids (array of loan demands)
	Asks []LendbookOffer // asks (array of loan offers)
}

// Orderbook ... Public (NEW)
type Orderbook struct {
	Bids []OrderbookOffer // bids (array of bid offers)
	Asks []OrderbookOffer // asks (array of ask offers)
}

// OrderbookOffer ... (NEW)
type OrderbookOffer struct {
	Price     float64 `json:"price,string"`     // price
	Amount    float64 `json:"amount,string"`    // amount (decimal)
	Timestamp float64 `json:"timestamp,string"` // time
}

// LendbookOffer ...
type LendbookOffer struct {
	Rate      float64 `json:"rate,string"`      // rate (rate in % per 365 days)
	Amount    float64 `json:"amount,string"`    // amount (decimal)
	Period    int     `json:"period"`           // period (days): minimum period for the loan
	Timestamp float64 `json:"timestamp,string"` // timestamp (time)
	FRRString string  `json:"frr"`              // frr (yes/no): "Yes" if the offer is at Flash Return Rate, "No" if the offer is at fixed rate
	FRR       bool
}

// WalletBalance ...
type WalletBalance struct {
	Type      string  `json:"type"`             // "trading", "deposit" or "exchange".
	Currency  string  `json:"currency"`         // Currency
	Amount    float64 `json:"amount,string"`    // How much balance of this currency in this wallet
	Available float64 `json:"available,string"` // How much X there is in this wallet that is available to trade.
}

// WalletKey ...
type WalletKey struct {
	Type, Currency string
}

// WalletBalances ...
type WalletBalances map[WalletKey]WalletBalance

// MyTrades ... (NEW)
type MyTrades []MyTrade

// MyTrade ... (NEW)
type MyTrade struct {
	Price       float64 `json:"price,string"`      // price
	Amount      float64 `json:"amount,string"`     // amount (decimal)
	Timestamp   float64 `json:"timestamp,string"`  // time
	Until       float64 `json:"until,string"`      // until (time): return only trades before or a the time specified here
	Exchange    string  `json:"exchange"`          // exchange
	Type        string  `json:"type"`              // type - "Sell" or "Buy"
	FeeCurrency string  `json:"fee_currency"`      // fee_currency (string) Currency you paid this trade's fee in
	FeeAmount   float64 `json:"fee_amount,string"` // fee_amount (decimal) Amount of fees you paid for this trade
	TID         int     `json:"tid"`               // tid (integer): unique identification number of the trade
	OrderId     int     `json:"order_id"`          // order_id (integer) unique identification number of the parent order of the trade
}

// Offer ...
type Offer struct {
	ID              int     `json:"id"`
	Currency        string  `json:"currency"`                // The currency name of the offer.
	Rate            float64 `json:"rate,string"`             // The rate the offer was issued at (in % per 365 days).
	Period          int     `json:"period"`                  // The number of days of the offer.
	Direction       string  `json:"direction"`               // Either "lend" or "loan".Either "lend" or "loan".
	Type            string  `json:"type"`                    // Either "market" / "limit" / "stop" / "trailing-stop".
	Timestamp       float64 `json:"timestamp,string"`        // The timestamp the offer was submitted.
	Live            bool    `json:"is_live,bool"`            // Could the offer still be filled?
	Cancelled       bool    `json:"is_cancelled,bool"`       // Has the offer been cancelled?
	ExecutedAmount  float64 `json:"executed_amount,string"`  // How much of the offer has been executed so far in its history?
	RemainingAmount float64 `json:"remaining_amount,string"` // How much is still remaining to be submitted?
	OriginalAmount  float64 `json:"original_amount,string"`  // What was the offer originally submitted for?
}

func (o *Offer) String() string {
	return fmt.Sprintf("ID: %d, Cur: %s, Rate: %f, Pr: %d, Dir: %s, Type: %s, TS: %f, Live: %t, Cancel: %t, ExAmt: %f, RemainAmt: %f, OrigAmt: %f",
		o.ID, o.Currency, o.Rate, o.Period, o.Direction, o.Type, o.Timestamp, o.Live, o.Cancelled, o.ExecutedAmount, o.RemainingAmount, o.OriginalAmount)
}

// Offers ...
type Offers []Offer

// Credit ...
type Credit struct {
	ID        int     `json:"id"`
	Currency  string  `json:"currency"`         // The currency name of the offer.
	Rate      float64 `json:"rate,string"`      // The rate the offer was issued at (in % per 365 days).
	Period    int     `json:"period"`           // The number of days of the offer.
	Amount    float64 `json:"amount,string"`    // How much is the credit for
	Status    string  `json:"status"`           // "Active"
	Timestamp float64 `json:"timestamp,string"` // The timestamp the offer was submitted.

}

// Credits ...
type Credits []Credit

// New returns a new Bitfinex API instance
func New(key, secret string) (api *API) {
	api = &API{
		APIKey:    key,
		APISecret: secret,
	}
	return api
}

///////////////////////////////////////
// Main API methods
///////////////////////////////////////

// Ticker returns innermost bid and asks and information on the most recent trade,
//	as well as high, low and volume of the last 24 hours.
func (api *API) Ticker(symbol string) (ticker Ticker, err error) {
	symbol = strings.ToLower(symbol)

	body, err := api.get("/v1/pubticker/" + symbol)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &ticker)
	if err != nil || ticker.LastPrice == 0 { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return ticker, errors.New("API: " + errorMessage.Message)
	}

	return
}

/*

type FundingTicker struct {
	Symbol            string  `json:"symbol"`
	FRR               float64 `json:"frr"`
	Bid               float64 `json:"bid"`
	BID_SIZE          float64 `json:"bidsize"`
	BID_PERIOD        float64 `json:"bidperiod"`
	ASK               float64 `json:"ask"`
	ASK_SIZE          float64 `json:"asksize"`
	ASK_PERIOD        float64 `json:"askperiod"`
	DAILY_CHANGE      float64 `json:"dailychange"`
	DAILY_CHANGE_PERC float64 `json:"dailchangeperc"`
	LAST_PRICE        float64 `json:"lastprice"`
	VOLUME            float64 `json:"volume"`
	HIGH              float64 `json:"high"`
	LOW               float64 `json:"low"`
}

/*
type V2FundingTicker struct {
	Symbol             string  `json:"symbol"`
	FRR                float64 `json:"frr"`
	Bid                float64 `json:"bid"`
	BidSize            float64 `json:"bidsize"`
	BidPeriod          float64 `json:"bidperiod"`
	Ask                float64 `json:"ask"`
	AskSize            float64 `json:"asksize"`
	AskPeriod          float64 `json:"askperiod"`
	DailyChange        float64 `json:"dailychange"`
	DailyChangePercent float64 `json:"dailchangeperc"`
	LastPrice          float64 `json:"lastprice"`
	Volume             float64 `json:"volume"`
	High               float64 `json:"high"`
	Low                float64 `json:"low"`
}

*/

func (api *API) AllLendingTickers() (ticker []V2Ticker, fundings []V2FundingTicker, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered in f", r)
		}
		return
	}()

	var url string
	for _, c := range LendingCurrencies[:len(LendingCurrencies)-1] {
		url += "t" + c + "USD,"
	}

	for i, c := range LendingCurrencies[:len(LendingCurrencies)] {
		url += "f" + c
		if i != len(LendingCurrencies)-1 {
			url += ","
		}
	}

	data, err := api.get("/v2/tickers?symbols=" + url)
	if err != nil {
		return
	}

	var arr [][]json.RawMessage
	err = json.Unmarshal(data, &arr)
	if err != nil {
		return
	}

	ticker = make([]V2Ticker, len(LendingCurrencies)-1)
	c := 0
	for i, _ := range ticker {
		ticker[i], err = jsonArrayToV2Ticker(arr[i])
		if err != nil {
			return
		}
		c++
	}

	fundings = make([]V2FundingTicker, len(LendingCurrencies))
	for i := range fundings {
		fundings[i], err = jsonArrayToV2FundingTicker(arr[c])
		if err != nil {
			return
		}
		c++
	}

	return
}

func jsonArrayToV2FundingTicker(arr []json.RawMessage) (ticker V2FundingTicker, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered in f", r)
		}
		return
	}()

	ticker.Symbol, err = jsonString(arr[0])
	if err != nil {
		return ticker, err
	}

	ticker.FRR, err = jsonFloat64(arr[1])
	if err != nil {
		return ticker, err
	}

	ticker.Bid, err = jsonFloat64(arr[2])
	if err != nil {
		return ticker, err
	}

	ticker.BidSize, err = jsonFloat64(arr[3])
	if err != nil {
		return ticker, err
	}

	ticker.BidPeriod, err = jsonFloat64(arr[4])
	if err != nil {
		return ticker, err
	}

	ticker.Ask, err = jsonFloat64(arr[5])
	if err != nil {
		return ticker, err
	}

	ticker.AskSize, err = jsonFloat64(arr[6])
	if err != nil {
		return ticker, err
	}

	ticker.AskPeriod, err = jsonFloat64(arr[7])
	if err != nil {
		return ticker, err
	}

	ticker.DailyChange, err = jsonFloat64(arr[8])
	if err != nil {
		return ticker, err
	}

	ticker.DailyChangePercent, err = jsonFloat64(arr[9])
	if err != nil {
		return ticker, err
	}

	ticker.LastPrice, err = jsonFloat64(arr[10])
	if err != nil {
		return ticker, err
	}

	ticker.Volume, err = jsonFloat64(arr[11])
	if err != nil {
		return ticker, err
	}

	ticker.High, err = jsonFloat64(arr[12])
	if err != nil {
		return ticker, err
	}

	ticker.Low, err = jsonFloat64(arr[13])
	if err != nil {
		return ticker, err
	}
	return ticker, nil
}

func jsonArrayToV2Ticker(arr []json.RawMessage) (ticker V2Ticker, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered in f", r)
		}
		return
	}()

	ticker.Symbol, err = jsonString(arr[0])
	if err != nil {
		return ticker, err
	}

	ticker.Bid, err = jsonFloat64(arr[1])
	if err != nil {
		return ticker, err
	}

	ticker.BidSize, err = jsonFloat64(arr[2])
	if err != nil {
		return ticker, err
	}

	ticker.Ask, err = jsonFloat64(arr[3])
	if err != nil {
		return ticker, err
	}

	ticker.AskSize, err = jsonFloat64(arr[4])
	if err != nil {
		return ticker, err
	}

	ticker.DailyChange, err = jsonFloat64(arr[5])
	if err != nil {
		return ticker, err
	}

	ticker.DailyChangePercent, err = jsonFloat64(arr[6])
	if err != nil {
		return ticker, err
	}

	ticker.LastPrice, err = jsonFloat64(arr[7])
	if err != nil {
		return ticker, err
	}

	ticker.Volume, err = jsonFloat64(arr[8])
	if err != nil {
		return ticker, err
	}

	ticker.High, err = jsonFloat64(arr[9])
	if err != nil {
		return ticker, err
	}

	ticker.Low, err = jsonFloat64(arr[10])
	if err != nil {
		return ticker, err
	}
	return ticker, nil
}

func jsonFloat64(data json.RawMessage) (float64, error) {
	var f float64
	err := json.Unmarshal(data, &f)
	return f, err
}

func jsonString(data json.RawMessage) (string, error) {
	var s string
	err := json.Unmarshal(data, &s)
	return s, err
}

// Stats return various statistics about the requested pairs.
func (api *API) Stats(symbol string) (stats Stats, err error) {
	symbol = strings.ToLower(symbol)

	body, err := api.get("/v1/stats/" + symbol)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &stats)
	if err != nil || len(stats) == 0 { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return stats, errors.New("API: " + errorMessage.Message)
	}

	return
}

// Orderbook returns the full order book.
func (api *API) Orderbook(symbol string, limitBids, limitAsks, group int) (orderbook Orderbook, err error) {
	symbol = strings.ToLower(symbol)

	body, err := api.get("/v1/book/" + symbol + "?limit_bids=" + strconv.Itoa(limitBids) + "&limit_asks=" + strconv.Itoa(limitAsks) + "&group=" + strconv.Itoa(group))
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &orderbook)
	if err != nil {
		return
	}

	return
}

// Lendbook returns the full lend book.
func (api *API) Lendbook(currency string, limitBids, limitAsks int) (lendbook Lendbook, err error) {
	currency = strings.ToLower(currency)

	body, err := api.get("/v1/lendbook/" + currency + "?limit_bids=" + strconv.Itoa(limitBids) + "&limit_asks=" + strconv.Itoa(limitAsks))
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &lendbook)
	if err != nil {
		return
	}

	if (limitAsks != 0 && len(lendbook.Asks) == 0) || (limitBids != 0 && len(lendbook.Bids) == 0) {
		return lendbook, errors.New("API: Lendbook empty, likely bad currency specified")
	}

	// Convert FRR strings to boolean values
	for _, p := range [](*[]LendbookOffer){&lendbook.Asks, &lendbook.Bids} {
		for i, e := range *p {
			if strings.ToLower(e.FRRString) == "yes" {
				e.FRR = true
				(*p)[i] = e
			}
		}
	}

	return
}

// WalletBalances return your balances.
func (api *API) WalletBalances() (wallet WalletBalances, err error) {
	request := struct {
		URL   string `json:"request"`
		Nonce string `json:"nonce"`
	}{
		"/v1/balances",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}

	tmpBalances := []WalletBalance{}
	err = json.Unmarshal(body, &tmpBalances)
	if err != nil { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return nil, errors.New("API: " + errorMessage.Message)
	}

	wallet = make(WalletBalances)
	for _, w := range tmpBalances {
		wallet[WalletKey{w.Type, w.Currency}] = w
	}

	return
}

func (api *API) GetSummary() (sum Summary, err error) {
	request := struct {
		URL   string `json:"request"`
		Nonce string `json:"nonce"`
	}{
		"/v1/summary",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}
	fmt.Println(string(body))

	err = json.Unmarshal(body, &sum)
	if err != nil { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return sum, errors.New("API: " + errorMessage.Message)
	}
	return
}

// MyTrades returns an array of your past trades for the given symbol.
func (api *API) MyTrades(symbol string, timestamp string, limitTrades int) (mytrades MyTrades, err error) {
	symbol = strings.ToLower(symbol)

	request := struct {
		URL         string `json:"request"`
		Nonce       string `json:"nonce"`
		Symbol      string `json:"symbol"`
		Timestamp   string `json:"timestamp"`
		LimitTrades int    `json:"limit_trades"`
	}{
		URL:         "/v1/mytrades",
		Nonce:       strconv.FormatInt(time.Now().UnixNano(), 10),
		Symbol:      symbol,
		Timestamp:   timestamp,
		LimitTrades: limitTrades,
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &mytrades)
	if err != nil { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return nil, errors.New("API: " + errorMessage.Message)
	}
	return
}

// CancelOffer cancel an offer give its id.
func (api *API) CancelOffer(id int) (err error) {
	request := struct {
		URL     string `json:"request"`
		Nonce   string `json:"nonce"`
		OfferID int    `json:"offer_id"`
	}{
		"/v1/offer/cancel",
		strconv.FormatInt(time.Now().UnixNano(), 10),
		id,
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}

	tmpOffer := struct {
		ID        int  `json:"id"`
		Cancelled bool `json:"is_cancelled,bool"`
	}{}

	err = json.Unmarshal(body, &tmpOffer)
	if err != nil || tmpOffer.ID != id { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return errors.New("API: " + errorMessage.Message)
	}

	if tmpOffer.Cancelled == true {
		return errors.New("API: Offer already cancelled")
	}

	return
}

// ActiveCredits return a list of currently lent funds (active credits).
func (api *API) ActiveCredits() (credits Credits, err error) {
	request := struct {
		URL   string `json:"request"`
		Nonce string `json:"nonce"`
	}{
		"/v1/credits",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &credits)
	if err != nil { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return credits, errors.New("API: " + errorMessage.Message)
	}

	return
}

// ActiveOffers return an array of all your live offers (lending or borrowing).
func (api *API) ActiveOffers() (offers Offers, err error) {
	request := struct {
		URL   string `json:"request"`
		Nonce string `json:"nonce"`
	}{
		"/v1/offers",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &offers)
	if err != nil { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return offers, errors.New("API: " + errorMessage.Message)
	}

	return
}

// NewOffer submits a new offer.
// currency (string): The name of the currency.
// amount (decimal): Offer size: how much to lend or borrow.
// rate (decimal): Rate to lend or borrow at. In percentage per 365 days.
// period (integer): Number of days of the loan (in days)
// direction (string): Either "lend" or "loan".
func (api *API) NewOffer(currency string, amount, rate float64, period int, direction string) (offer Offer, err error) {
	currency = strings.ToUpper(currency)
	direction = strings.ToLower(direction)

	request := struct {
		URL       string  `json:"request"`
		Nonce     string  `json:"nonce"`
		Currency  string  `json:"currency"`
		Amount    float64 `json:"amount,string"`
		Rate      float64 `json:"rate,string"`
		Period    int     `json:"period"`
		Direction string  `json:"direction"`
	}{
		"/v1/offer/new",
		strconv.FormatInt(time.Now().UnixNano(), 10),
		currency,
		amount,
		rate,
		period,
		direction,
	}

	body, err := api.post(request.URL, request)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &offer)
	if err != nil || offer.ID == 0 { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(body, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			return
		}

		return offer, errors.New("API: " + errorMessage.Message)
	}

	return
}

///////////////////////////////////////
// API helper methods
///////////////////////////////////////

// CancelActiveOffers ...
func (api *API) CancelActiveOffers() (err error) {
	offers, err := api.ActiveOffers()
	if err != nil {
		return
	}

	for _, o := range offers {
		err = api.CancelOffer(o.ID)

		if err != nil {
			return
		}
	}

	return
}

// CancelActiveOffersByCurrency ...
func (api *API) CancelActiveOffersByCurrency(currency string) (err error) {
	currency = strings.ToLower(currency)

	offers, err := api.ActiveOffers()
	if err != nil {
		return
	}

	for _, o := range offers {
		if strings.ToLower(o.Currency) == currency {
			err = api.CancelOffer(o.ID)
			if err != nil {
				return
			}
		}
	}

	return
}

type FundingEarning struct {
	Currency    string `json:"currency"`
	Amount      string `json:"amount"`
	Balance     string `json:"balance"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

func (api *API) GetFundingEarningsFromTime(start, end time.Time) ([]FundingEarning, error) {
	return api.GetFundingEarnings(fmt.Sprintf("%d", start.Unix()), fmt.Sprintf("%d", end.Unix()))
}

func (api *API) GetFundingEarnings(start, end string) ([]FundingEarning, error) {
	request := struct {
		URL      string `json:"request"`
		Nonce    string `json:"nonce"`
		Currency string `json:"currency"`
		Wallet   string `json:"wallet"`
		Since    string `json:"since"`
		Until    string `json:"until"`
	}{
		"/v1/history",
		strconv.FormatInt(time.Now().UnixNano(), 10),
		"ETH",
		"deposit",
		start,
		end,
	}

	var all []json.RawMessage

	resp, err := api.post(request.URL, request)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &all)
	if err != nil { // Failed to unmarshal expected message
		// Attempt to unmarshal the error message
		errorMessage := ErrorMessage{}
		err = json.Unmarshal(resp, &errorMessage)
		if err != nil { // Not expected message and not expected error, bailing...
			if len(resp) > 100 {
				resp = resp[:100]
			}
			return nil, fmt.Errorf("Unknown api error: %s", string(resp))
		}

		return nil, errors.New("API: " + errorMessage.Message)
	}

	var earnings []FundingEarning

	for _, i := range all {
		var n FundingEarning
		err := json.Unmarshal(i, &n)
		if err != nil {
			continue
		}
		if n.Description != "Margin Funding Payment on wallet Deposit" {
			continue
		}
		earnings = append(earnings, n)
	}

	return earnings, nil
}

///////////////////////////////////////
// API query methods
///////////////////////////////////////

func (api *API) get(url string) (body []byte, err error) {
	client := http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	resp, err := client.Get(APIURL + url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	return
}

func (api *API) post(url string, payload interface{}) (body []byte, err error) {
	// X-BFX-PAYLOAD
	// parameters-dictionary -> JSON encode -> base64
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return
	}
	payloadBase64 := base64.StdEncoding.EncodeToString(payloadJSON)

	// X-BFX-SIGNATURE
	// HMAC-SHA384(payload, api-secret) as hexadecimal
	h := hmac.New(sha512.New384, []byte(api.APISecret))
	h.Write([]byte(payloadBase64))
	signature := hex.EncodeToString(h.Sum(nil))

	// POST
	req, err := http.NewRequest("POST", APIURL+url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return
	}

	req.Header.Add("X-BFX-APIKEY", api.APIKey)
	req.Header.Add("X-BFX-PAYLOAD", payloadBase64)
	req.Header.Add("X-BFX-SIGNATURE", signature)

	client := http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	return
}
