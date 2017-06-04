package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
)

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

func getUserStats(email string) (*CurrentUserStatistics, *UserBalanceDetails) {
	userStats, err := state.GetUserStatistics(email, 2)

	balanceDetails := newUserBalanceDetails()
	today := newCurrentUserStatistics()
	if err != nil {
		return today, balanceDetails
	}
	l := len(userStats)
	if l > 0 && len(userStats[0]) > 0 {
		now := userStats[0][0]
		// Set balance ratios
		balanceDetails.CurrencyMap = now.TotalCurrencyMap
		balanceDetails.compute()

		today.LoanRate = now.AverageActiveRate
		today.BTCLent = now.ActiveLentBalance
		today.BTCNotLent = now.AverageOnOrderRate + now.AvailableBalance
		today.LendingPercent = today.BTCLent / (today.BTCLent + today.BTCNotLent)

		yesterday := userdb.GetDayAvg(userStats[1])
		if yesterday != nil {
			today.LoanRateChange = percentChange(today.LoanRate, yesterday.LoanRate)
			today.BTCLentChange = percentChange(today.BTCLent, yesterday.BTCLent)
			today.BTCNotLentChange = percentChange(today.BTCNotLent, yesterday.BTCNotLent)
			today.LendingPercentChange = percentChange(today.LendingPercent, yesterday.LendingPercent)
		}
	}

	return today, balanceDetails
}

func (r AppAuthRequired) CurrentUserStats() revel.Result {
	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Println("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}

	data := make(map[string]interface{})
	stats, balanceBreakdown := getUserStats(email)

	data["CurrentUserStats"] = stats
	data["Balances"] = balanceBreakdown
	return r.RenderJSON(data)
}

func (r AppAuthRequired) LendingHistory() revel.Result {
	email := r.Session[SESSION_EMAIL]
	u, err := state.FetchUser(email)
	if err != nil || u == nil {
		fmt.Println("Error fetching user for dashboard")
		return r.Redirect(App.Index)
	}

	//to cache
	completeLoans, err := state.PoloniexAuthenticatedLendingHistory(u.Username, "", "")
	data := make(map[string]interface{})
	data["CompleteLoans"] = completeLoans.Data
	if len(completeLoans.Data) == 0 && revel.DevMode {
		var cl [20]poloniex.PoloniexAuthentictedLendingHistory
		for i := 0; i < 20; i++ {
			cl[i] = poloniex.PoloniexAuthentictedLendingHistory{
				361915250,
				"BTC",
				"0.00066000",
				"0.00011775",
				"0.05150000",
				"0.00000001",
				"0.00000000",
				"0.00000001",
				"2017-06-03 22:55:30",
				"2017-06-04 00:09:39",
			}
		}
		data["CompleteLoans"] = cl
	}

	return r.RenderJSON(data)
}
