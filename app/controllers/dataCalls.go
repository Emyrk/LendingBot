package controllers

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"

	log "github.com/sirupsen/logrus"
)

var dcLog = log.WithField("package", "DataCalls")

// Struct to UserDash
type UserDashStructure struct {
}

type CurrentUserStatistics struct {
	LoanRate       float64 `json:"loanrate"`
	BTCLent        float64 `json:"btclent"`
	BTCNotLent     float64 `json:"btcnotlent"`
	LendingPercent float64 `json:"lendingpercent"`

	LoanRateChange       float64 `json:"loanratechange"`
	BTCLentChange        float64 `json:"btclentchange"`
	BTCNotLentChange     float64 `json:"btcnotlentchange"`
	LendingPercentChange float64 `json:"lendingpercentchange"`

	// From poloniex call
	BTCEarned float64 `json:"btcearned"`
}

func (c *CurrentUserStatistics) scrub() {
	if math.IsNaN(c.LoanRate) {
		c.LoanRate = 0
	}
	if math.IsNaN(c.BTCLent) {
		c.BTCLent = 0
	}
	if math.IsNaN(c.BTCNotLent) {
		c.BTCNotLent = 0
	}
	if math.IsNaN(c.LendingPercent) {
		c.LendingPercent = 0
	}
	if math.IsNaN(c.LoanRateChange) {
		c.LoanRateChange = 0
	}
	if math.IsNaN(c.BTCLentChange) {
		c.BTCLentChange = 0
	}
	if math.IsNaN(c.BTCNotLentChange) {
		c.BTCNotLentChange = 0
	}
	if math.IsNaN(c.LendingPercentChange) {
		c.LendingPercentChange = 0
	}
	if math.IsNaN(c.BTCEarned) {
		c.BTCEarned = 0
	}
}

func newCurrentUserStatistics() *CurrentUserStatistics {
	r := new(CurrentUserStatistics)
	r.LoanRate = 0
	r.BTCLent = 0
	r.BTCNotLent = 0
	r.LendingPercent = 0
	r.BTCEarned = 0

	r.LoanRateChange = 0
	r.BTCLentChange = 0
	r.BTCNotLentChange = 0
	r.LendingPercentChange = 0

	return r
}

// UserBalanceDetails is their current lending balances
type UserBalanceDetails struct {
	CurrencyMap map[string]float64 `json:"currencymap"`
	Percent     map[string]float64 `json:"percentmap"`
}

func newUserBalanceDetails() *UserBalanceDetails {
	u := new(UserBalanceDetails)
	u.CurrencyMap = make(map[string]float64)
	u.Percent = make(map[string]float64)
	return u
}

// compute computed the percentmap
func (u *UserBalanceDetails) compute() {
	total := float64(0)
	for _, v := range u.CurrencyMap {
		total += v
	}

	for k, v := range u.CurrencyMap {
		u.Percent[k] = v / total
	}
}

func (u *UserBalanceDetails) scrub() {
	for k, v := range u.CurrencyMap {
		if math.IsNaN(v) {
			u.CurrencyMap[k] = 0
		}

		if k == "" {
			delete(u.CurrencyMap, k)
		}
	}

	for k, v := range u.Percent {
		if math.IsNaN(v) {
			u.Percent[k] = 0
		}

		if k == "" {
			delete(u.Percent, k)
		}
	}
}

func getUserStats(email string) (*CurrentUserStatistics, *UserBalanceDetails) {
	userStats, err := state.GetUserStatistics(email, 2)

	balanceDetails := newUserBalanceDetails()
	today := newCurrentUserStatistics()
	if err != nil {
		balanceDetails.compute()
		return today, balanceDetails
	}
	l := len(userStats)
	if l > 0 && len(userStats[0]) > 0 {
		now := userStats[0][0]
		// Set balance ratios
		balanceDetails.CurrencyMap = now.TotalCurrencyMap
		balanceDetails.compute()

		totalAct := float64(0)
		for _, v := range now.Currencies {
			// fmt.Println(v)
			today.LoanRate += v.AverageActiveRate * (v.ActiveLentBalance * v.BTCRate)
			totalAct += v.ActiveLentBalance * v.BTCRate
			today.BTCLent += v.ActiveLentBalance * v.BTCRate
			today.BTCNotLent += (v.OnOrderBalance + v.AvailableBalance) * v.BTCRate
		}
		today.LoanRate = today.LoanRate / totalAct

		today.LendingPercent = today.BTCLent / (today.BTCLent + today.BTCNotLent)

		yesterday := userdb.GetCombinedDayAverage(userStats[1])
		if yesterday != nil {
			today.LoanRateChange = today.LoanRate - yesterday.LoanRate
			today.BTCLentChange = today.BTCLent - yesterday.Lent
			today.BTCNotLentChange = today.BTCNotLent - yesterday.NotLent
			today.LendingPercentChange = today.LendingPercent - yesterday.LendingPercent
		}
	}

	return today, balanceDetails
}

func (r AppAuthRequired) CurrentUserStats() revel.Result {
	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Println("Error: CurrentUserStats: fetching user for dashboard")
		return r.Redirect(App.Index)
	}

	data := make(map[string]interface{})
	stats, balanceBreakdown := getUserStats(email)

	// Scrub for NaNs
	stats.scrub()
	balanceBreakdown.scrub()

	data["CurrentUserStats"] = stats
	data["Balances"] = balanceBreakdown
	return r.RenderJSON(data)
}

func (r AppAuthRequired) GetDetailedUserStats() revel.Result {
	data := make(map[string]interface{})

	llog := dcLog.WithField("method", "GetDetailedUserStats")
	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		llog.Error("Fetching user for dashboard")
		return r.Redirect(App.Index)
	}

	stats, err := state.GetUserStatistics(email, 30)
	if err != nil {
		llog.Errorf("Getting user stats: %s", err.Error())
		data[JSON_ERROR] = "Internal Error grabbing stats"
		r.Response.Status = 500
		return r.RenderJSON(data)
	}

	data[JSON_DATA] = stats

	return r.RenderJSON(data)
}

func (r AppAuthRequired) LendingHistorySummary() revel.Result {
	llog := dcLog.WithField("method", "LendingHistorySummary()")
	data := make(map[string]interface{})

	email := r.Session[SESSION_EMAIL]
	var month []*userdb.AllLendingHistoryEntry

	t := time.Now().Add(-48 * time.Hour)
	for i := 0; i < 28; i++ {
		day, err := state.LoadLendingSummary(email, t)
		if err != nil {
			llog.Errorf("Error loading %s: %s", email, err.Error())
		}
		t = t.Add(-24 * time.Hour)
		month = append(month, day)
	}

	data["LoanHistory"] = month
	data["USDRates"] = getRates()
	rateLock.RLock()
	defer rateLock.RUnlock()

	return r.RenderJSON(data)
}

var rateLock sync.RWMutex
var Rates map[string]float64
var lastRates time.Time

func init() {
	Rates = make(map[string]float64)
	for _, c := range curarr {
		Rates[fmt.Sprintf("USD_%s", c)] = 0
	}
}

func getRates() map[string]float64 {
	if Lender.Ticker == nil {
		return Rates
	}

	if Rates != nil && time.Since(lastRates).Seconds() < 100 {
		return Rates
	}
	lastRates = time.Now()

	m := make(map[string]float64)

	Lender.TickerLock.RLock()
	btcusd, _ := Lender.Ticker["USDT_BTC"]
	for _, c := range curarr {
		if c == "BTC" {
			m["USD_BTC"] = btcusd.Last
		} else {
			btccur, ok := Lender.Ticker[fmt.Sprintf("BTC_%s", c)]
			if !ok {
				m[fmt.Sprintf("USD_%s", c)] = 0
				continue
			}
			m[fmt.Sprintf("USD_%s", c)] = btccur.Last * btcusd.Last
		}
	}
	Lender.TickerLock.RUnlock()

	rateLock.Lock()
	Rates = m
	rateLock.Unlock()
	return m
}

func (r AppAuthRequired) LendingHistory() revel.Result {
	llog := dcLog.WithField("method", "LendingHistory()")

	email := r.Session[SESSION_EMAIL]

	data := make(map[string]interface{})

	var completeLoans *poloniex.PoloniexAuthentictedLendingHistoryRespone
	tempCompleteLoans, found := CacheGetLendingHistory(email)
	if !found {
		u, err := state.FetchUser(email)
		if err != nil || u == nil {
			llog.Error("Error: LendingHistory: fetching user for dashboard")
			return r.Redirect(App.Index)
		}

		past := time.Now().Add(-24 * time.Hour * 30)
		start := fmt.Sprintf("%d", past.Unix())
		end := fmt.Sprintf("%d", time.Now().Unix())

		tc, err := state.PoloniexOffloadAuthenticatedLendingHistory(u.Username, start, end, "")
		if err != nil {
			llog.Errorf("Error getting lend history for %s: %s\n", email, err.Error())
		}
		completeLoans = &tc
		if len(completeLoans.Data) == 0 && revel.DevMode {
			var cl [20]poloniex.PoloniexAuthentictedLendingHistory
			for i := 0; i < 20; i++ {
				tempA := (20 - i) * 24
				earned := rand.Float32()
				cl[i] = poloniex.PoloniexAuthentictedLendingHistory{
					361915250,
					"BTC",
					fmt.Sprintf("%f", rand.Float32()),
					fmt.Sprintf("%f", rand.Float32()),
					fmt.Sprintf("%f", rand.Float32()),
					fmt.Sprintf("%f", rand.Float32()),
					fmt.Sprintf("%f", (earned * .015)),
					fmt.Sprintf("%f", earned),
					time.Now().Add(-time.Duration(tempA) * time.Hour).Add(1 * time.Hour).String(),
					time.Now().Add(-time.Duration(tempA) * time.Hour).Add(2 * time.Hour).String(),
				}
			}
			completeLoans = &poloniex.PoloniexAuthentictedLendingHistoryRespone{
				cl[:],
			}
		}
		CacheSetLendingHistory(email, *completeLoans)
	} else {
		completeLoans = tempCompleteLoans
	}

	coin := r.Params.Get("coin")
	if len(coin) > 0 && userdb.CoinExists(coin) {
		llog.Info("Getting coin: " + coin)
		var tempSpecificLoans []poloniex.PoloniexAuthentictedLendingHistory
		for _, clCoin := range completeLoans.Data {
			if clCoin.Currency == coin {
				tempSpecificLoans = append(tempSpecificLoans, clCoin)
			}
		}
		completeLoans = &poloniex.PoloniexAuthentictedLendingHistoryRespone{
			tempSpecificLoans[:],
		}
	}

	data["CompleteLoans"] = completeLoans.Data

	return r.RenderJSON(data)
}

// TODO: Cache this response
func (r App) GetPoloniexStatistics() revel.Result {
	return r.RenderJSON(state.GetPoloniexStatistics("BTC"))
}
