package poloniex

import (
	"errors"
	"math/rand"
	"net/url"
	"time"
)

var (
	NotImplementedError error = errors.New("Not implemented")
)

type FakePoloniex struct {
	Name                    string
	Enabled                 bool
	Verbose                 bool
	Websocket               bool
	RESTPollingDelay        time.Duration
	AuthenticatedAPISupport bool
	// AccessKey, SecretKey    string
	Fee            float64
	BaseCurrencies []string
	AvailablePairs []string
	EnabledPairs   []string
}

func (p *FakePoloniex) SetDefaults() {
	p.Name = "Poloniex"
	p.Enabled = false
	p.Fee = 0
	p.Verbose = false
	p.Websocket = false
	p.RESTPollingDelay = 10
}

func (p *FakePoloniex) GetName() string {
	return p.Name
}

func (p *FakePoloniex) SetEnabled(enabled bool) {
	p.Enabled = enabled
}

func (p *FakePoloniex) IsEnabled() bool {
	return p.Enabled
}

func (p *FakePoloniex) Setup(exch Exchanges) {
	if !exch.Enabled {
		p.SetEnabled(false)
	} else {
		p.Enabled = true
		p.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		// p.SetAPIKeys(exch.APIKey, exch.APISecret)
		p.RESTPollingDelay = exch.RESTPollingDelay
		p.Verbose = exch.Verbose
		p.Websocket = exch.Websocket
		p.BaseCurrencies = SplitStrings(exch.BaseCurrencies, ",")
		p.AvailablePairs = SplitStrings(exch.AvailablePairs, ",")
		p.EnabledPairs = SplitStrings(exch.EnabledPairs, ",")
	}
}

//
// Fake These
//

func (p *FakePoloniex) CreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, accessKey, secretKey string) (int64, error) {
	return rand.Int63(), nil
}

func (p *FakePoloniex) CancelLoanOffer(currency string, orderNumber int64, accessKey, secretKey string) (bool, error) {
	return true, nil
}

func (p *FakePoloniex) GetLoanOrders(currency string) (*PoloniexLoanOrders, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetOpenLoanOffers(accessKey, secretKey string) (map[string][]PoloniexLoanOffer, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetActiveLoans(accessKey, secretKey string) (*PoloniexActiveLoans, error) {
	return nil, NotImplementedError
}

//
//
//

func (p *FakePoloniex) Start() {
	return
}

func (p *FakePoloniex) GetFee() float64 {
	return 0
}

func (p *FakePoloniex) Run() {
	return
}

func (p *FakePoloniex) GetTicker() (map[string]PoloniexTicker, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetVolume() (interface{}, error) {
	return nil, NotImplementedError
}

//TO-DO: add support for individual pair depth fetching
func (p *FakePoloniex) GetOrderbook(currencyPair string, depth int) (map[string]PoloniexOrderbook, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetTradeHistory(currencyPair, start, end string) ([]PoloniexTradeHistory, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetChartData(currencyPair, start, end, period string) ([]PoloniexChartData, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetCurrencies() (map[string]PoloniexCurrencies, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetBalances(accessKey, secretKey string) (*PoloniexBalance, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetCompleteBalances(accessKey, secretKey string) (PoloniexCompleteBalances, error) {
	return PoloniexCompleteBalances{}, NotImplementedError
}

func (p *FakePoloniex) GetDepositAddresses(accessKey, secretKey string) (PoloniexDepositAddresses, error) {
	return PoloniexDepositAddresses{}, NotImplementedError
}

func (p *FakePoloniex) GenerateNewAddress(currency string, accessKey, secretKey string) (string, error) {
	return "", NotImplementedError
}

func (p *FakePoloniex) GetDepositsWithdrawals(start, end string, accessKey, secretKey string) (PoloniexDepositsWithdrawals, error) {
	return PoloniexDepositsWithdrawals{}, NotImplementedError
}

func (p *FakePoloniex) GetOpenOrders(currency string, accessKey, secretKey string) (interface{}, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetAuthenticatedTradeHistory(currency, start, end string, accessKey, secretKey string) (interface{}, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) PlaceOrder(currency string, rate, amount float64, immediate, fillOrKill, buy bool, accessKey, secretKey string) (PoloniexOrderResponse, error) {
	return PoloniexOrderResponse{}, NotImplementedError
}

func (p *FakePoloniex) CancelOrder(orderID int64, accessKey, secretKey string) (bool, error) {
	return true, nil
}

func (p *FakePoloniex) MoveOrder(orderID int64, rate, amount float64, accessKey, secretKey string) (PoloniexMoveOrderResponse, error) {
	return PoloniexMoveOrderResponse{}, NotImplementedError
}

func (p *FakePoloniex) Withdraw(currency, address string, amount float64, accessKey, secretKey string) (bool, error) {
	return true, nil
}

func (p *FakePoloniex) GetFeeInfo(accessKey, secretKey string) (PoloniexFee, error) {
	return PoloniexFee{}, NotImplementedError
}

func (p *FakePoloniex) GetTradableBalances(accessKey, secretKey string) (map[string]map[string]float64, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) GetAvilableBalances(accessKey, secretKey string) (map[string]map[string]float64, error) {
	return nil, NotImplementedError

}

func (p *FakePoloniex) TransferBalance(currency, from, to string, amount float64, accessKey, secretKey string) (bool, error) {
	return true, nil
}

func (p *FakePoloniex) GetMarginAccountSummary(accessKey, secretKey string) (PoloniexMargin, error) {
	return PoloniexMargin{}, NotImplementedError
}

func (p *FakePoloniex) PlaceMarginOrder(currency string, rate, amount, lendingRate float64, buy bool, accessKey, secretKey string) (PoloniexOrderResponse, error) {
	return PoloniexOrderResponse{}, NotImplementedError
}

func (p *FakePoloniex) GetMarginPosition(currency string, accessKey, secretKey string) (interface{}, error) {
	return nil, NotImplementedError
}

func (p *FakePoloniex) CloseMarginPosition(currency string, accessKey, secretKey string) (bool, error) {
	return true, nil
}

func (p *FakePoloniex) ToggleAutoRenew(orderNumber int64, accessKey, secretKey string) (bool, error) {
	return true, nil
}

func (p *FakePoloniex) SendAuthenticatedHTTPRequest(method, endpoint string, values url.Values, result interface{}, accessKey, secretKey string) error {
	return NotImplementedError

}
