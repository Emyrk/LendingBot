package poloniex

import (
	"net/url"
)

type IBotExchange interface {
	Setup(exch Exchanges)
	Start()
	SetDefaults()
	GetName() string
	IsEnabled() bool
}

type IPoloniex interface {
	SetDefaults()
	GetName() string
	SetEnabled(enabled bool)
	IsEnabled() bool
	Setup(exch Exchanges)
	Start()
	GetFee() float64
	Run()
	GetTicker() (map[string]PoloniexTicker, error)
	GetVolume() (interface{}, error)
	GetOrderbook(currencyPair string, depth int) (map[string]PoloniexOrderbook, error)
	GetTradeHistory(currencyPair, start, end string) ([]PoloniexTradeHistory, error)
	GetChartData(currencyPair, start, end, period string) ([]PoloniexChartData, error)
	GetCurrencies() (map[string]PoloniexCurrencies, error)
	GetLoanOrders(currency string) (*PoloniexLoanOrders, error)
	GetBalances(accessKey, secretKey string) (*PoloniexBalance, error)
	GetCompleteBalances(accessKey, secretKey string) (PoloniexCompleteBalances, error)
	GetDepositAddresses(accessKey, secretKey string) (PoloniexDepositAddresses, error)
	GenerateNewAddress(currency string, accessKey, secretKey string) (string, error)
	GetDepositsWithdrawals(start, end string, accessKey, secretKey string) (PoloniexDepositsWithdrawals, error)
	GetOpenOrders(currency string, accessKey, secretKey string) (interface{}, error)
	GetAuthenticatedTradeHistory(currency, start, end string, accessKey, secretKey string) (interface{}, error)
	PlaceOrder(currency string, rate, amount float64, immediate, fillOrKill, buy bool, accessKey, secretKey string) (PoloniexOrderResponse, error)
	CancelOrder(orderID int64, accessKey, secretKey string) (bool, error)
	MoveOrder(orderID int64, rate, amount float64, accessKey, secretKey string) (PoloniexMoveOrderResponse, error)
	Withdraw(currency, address string, amount float64, accessKey, secretKey string) (bool, error)
	GetFeeInfo(accessKey, secretKey string) (PoloniexFee, error)
	GetTradableBalances(accessKey, secretKey string) (map[string]map[string]float64, error)
	GetAvilableBalances(accessKey, secretKey string) (map[string]map[string]float64, error)
	TransferBalance(currency, from, to string, amount float64, accessKey, secretKey string) (bool, error)
	GetMarginAccountSummary(accessKey, secretKey string) (PoloniexMargin, error)
	PlaceMarginOrder(currency string, rate, amount, lendingRate float64, buy bool, accessKey, secretKey string) (PoloniexOrderResponse, error)
	GetMarginPosition(currency string, accessKey, secretKey string) (interface{}, error)
	CloseMarginPosition(currency string, accessKey, secretKey string) (bool, error)
	CreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, accessKey, secretKey string) (int64, error)
	CancelLoanOffer(currency string, orderNumber int64, accessKey, secretKey string) (bool, error)
	GetOpenLoanOffers(accessKey, secretKey string) (map[string][]PoloniexLoanOffer, error)
	GetActiveLoans(accessKey, secretKey string) (*PoloniexActiveLoans, error)
	ToggleAutoRenew(orderNumber int64, accessKey, secretKey string) (bool, error)
	SendAuthenticatedHTTPRequest(method, endpoint string, values url.Values, result interface{}, accessKey, secretKey string) error
	GetAuthenticatedLendingHistory(start, end string, accessKey, secretKey string) (PoloniexAuthentictedLendingHistoryRespone, error)
}
