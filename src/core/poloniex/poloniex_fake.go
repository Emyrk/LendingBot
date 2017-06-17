package poloniex

import (
	"errors"
	"fmt"
	"math/rand"
	// "net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/scraper/client"
)

var (
	NotImplementedError error = errors.New("Not implemented")
)

type FakeLoanStruct struct {
	Loan PoloniexLoanOffer

	// Time it takes for the loan to be active
	TakeTime time.Time

	// Time for loan to be returned
	ReturnTime time.Time
}

func (fk *FakeLoanStruct) Active(t time.Time) bool {
	if fk.TakeTime.Before(t) {
		return true
	}
	return false
}

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

	// Fake Specific
	Scraper       *client.ScraperClient
	MyLoans       map[int64]*FakeLoanStruct
	availBalLock  sync.RWMutex
	AvailBalances map[string]map[string]float64
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

	p.Scraper = client.NewScraperClient("Scraper", "localhost:50051")
	p.MyLoans = make(map[int64]*FakeLoanStruct)
	p.AvailBalances = make(map[string]map[string]float64)
	p.AvailBalances["lending"] = make(map[string]float64)
	p.AvailBalances["margin"] = make(map[string]float64)
	p.AvailBalances["exchange"] = make(map[string]float64)

	for k, _ := range p.AvailBalances {
		p.AvailBalances[k]["BTC"] = 0
	}
}

func (p *FakePoloniex) AddFunds(currency string, amt float64) {
	p.availBalLock.Lock()
	p.AvailBalances["lending"][currency] += amt
	p.availBalLock.Unlock()
	return
}

func (p *FakePoloniex) LoadDay(day []byte) error {
	return p.Scraper.LoadDay(day)
}

func (p *FakePoloniex) GetLastDayAndSecond() (day []byte, second []byte, err error) {
	return p.Scraper.GetLastDayAndSecond()
}

func (p *FakePoloniex) String() string {
	n := time.Now()
	header := fmt.Sprintf("-- Fake Poloniex Summary %s --", time.Now().String())
	balance := p.GetBalanceDetails(n)

	return header + balance
}

func (p *FakePoloniex) GetBalanceDetails(t time.Time) string {
	p.CheckLoanReturns()
	p.availBalLock.RLock()
	availAmt := p.AvailBalances["lending"]["BTC"]
	avail := fmt.Sprintf("%20s:%f BTC\n", "Available Balances", availAmt)
	p.availBalLock.RUnlock()

	var takenAmt, waitingAmt float64
	for _, l := range p.MyLoans {
		if l.Active(t) {
			takenAmt += l.Loan.Amount
		} else {
			waitingAmt += l.Loan.Amount
		}
	}

	taken := fmt.Sprintf("%20s:%f BTC\n", "Active Loan", takenAmt)
	waiting := fmt.Sprintf("%20s:%f BTC\n", "InActive Loan", waitingAmt)
	total := fmt.Sprintf("%20s:%f BTC\n", "Total Balance", takenAmt+waitingAmt+availAmt)

	return total + avail + taken + waiting
}

//
// Fake These
//

func (p *FakePoloniex) GetAvilableBalances(accessKey, secretKey string) (map[string]map[string]float64, error) {
	newMap := make(map[string]map[string]float64)
	p.availBalLock.Lock()
	for k, v := range p.AvailBalances {
		newMap[k] = make(map[string]float64)
		for k2, v2 := range v {
			newMap[k][k2] = v2
		}
	}
	p.availBalLock.Unlock()

	return newMap, nil

}

func (p *FakePoloniex) CreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool, accessKey, secretKey string) (int64, error) {
	p.CheckLoanReturns()
	p.availBalLock.RLock()
	have := p.AvailBalances["lending"]["BTC"]
	p.availBalLock.RUnlock()
	if have < amount {
		return 0, fmt.Errorf("Not enough BTC. Have %f, need %f", amount, have)
	}

	fk := new(FakeLoanStruct)
	fk.Loan.Amount = amount
	fk.Loan.Currency = currency
	fk.Loan.AutoRenew = func() int {
		if autoRenew {
			return 1
		}
		return 0
	}()
	fk.Loan.Date = time.Now().String()
	fk.Loan.Duration = duration
	fk.Loan.ID = rand.Int63()
	fk.Loan.Rate = rate

	takeTime := time.Now().Add(time.Duration(rand.Intn(10)) * time.Second)
	fk.TakeTime = takeTime
	fk.ReturnTime = takeTime.Add(time.Duration(rand.Intn(10)) * time.Second)

	p.MyLoans[fk.Loan.ID] = fk

	p.availBalLock.Lock()
	p.AvailBalances["lending"]["BTC"] -= amount
	p.availBalLock.Unlock()

	return fk.Loan.ID, nil
}

func (p *FakePoloniex) CheckLoanReturns() {
	n := time.Now()
	for id, l := range p.MyLoans {
		if n.Before(l.ReturnTime) {
			continue
		}

		rt := time.Since(l.ReturnTime).Seconds()
		tt := time.Since(l.TakeTime).Seconds()
		total := rt - tt
		totalDays := total / 86400

		p.availBalLock.Lock()
		p.AvailBalances["lending"]["BTC"] += l.Loan.Amount + (l.Loan.Amount * (l.Loan.Rate * totalDays))
		p.availBalLock.Unlock()
		delete(p.MyLoans, id)
	}
}

func (p *FakePoloniex) CancelLoanOffer(currency string, orderNumber int64, accessKey, secretKey string) (bool, error) {
	p.CheckLoanReturns()
	if !p.removeLoan(orderNumber) {
		return false, fmt.Errorf("Loan not found")
	}

	return true, nil
}

func (p *FakePoloniex) removeLoan(loanid int64) bool {
	if _, ok := p.MyLoans[loanid]; ok {
		delete(p.MyLoans, loanid)
		return true
	}
	return false
}

func (p *FakePoloniex) GetLoanOrders(currency string) (*PoloniexLoanOrders, error) {
	data, err := p.Scraper.ReadNext()
	if err != nil {
		return nil, err
	}

	ret := new(PoloniexLoanOrders)
	err = JSONDecode(data, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *FakePoloniex) GetOpenLoanOffers(accessKey, secretKey string) (map[string][]PoloniexLoanOffer, error) {
	p.CheckLoanReturns()

	all := make([]PoloniexLoanOffer, 0)
	open := make(map[string][]PoloniexLoanOffer)

	n := time.Now()
	for _, l := range p.MyLoans {
		if l.TakeTime.Before(n) {
			continue
		}
		all = append(all, l.Loan)
	}
	open["BTC"] = all
	return open, nil
}

func (p *FakePoloniex) GetActiveLoans(accessKey, secretKey string) (*PoloniexActiveLoans, error) {
	p.CheckLoanReturns()

	loans := new(PoloniexActiveLoans)
	loans.Provided = make([]PoloniexLoanOffer, 0)
	loans.Used = make([]PoloniexLoanOffer, 0)

	n := time.Now()
	for _, l := range p.MyLoans {
		if l.TakeTime.Before(n) {
			loans.Used = append(loans.Used, l.Loan)
		} else {
			loans.Provided = append(loans.Provided, l.Loan)
		}
	}

	return loans, nil
}

//
//
//

func (p *FakePoloniex) ConstructAuthenticatedLendingHistoryRequest(start, end, limit string, accessKey, secretKey string) (*RequestHolder, error) {
	return nil, nil
}

func (p *FakePoloniex) GetAuthenticatedLendingHistory(start, end, limit string, accessKey, secretKey string) (resp PoloniexAuthentictedLendingHistoryRespone, err error) {
	return
}

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
