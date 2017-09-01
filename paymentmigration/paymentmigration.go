package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/payment"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = os.DevNull

//all individuals who did the survey
var SURVEY_DISCOUNT = []string{
	"Ra2o2@gmx.net",
	"fb89551@gmail.com",
	"jzhan@atb-intl.com",
	"kinfreng@gmail.com",
	"leonscot.p@gmail.com",
	"rvmouche@gmail.com",
	"s980705s@gmail.com",
	"sfw.sakana@gmail.com",
	"sm4r7m0u53@yahoo.it",
	"snipeboom@gmail.com",
	"special1311@hotmail.com",
}

func main() {
	if os.Getenv("MONGO_BAL_PASS") == "" {
		panic("Running in prod, but no balancer pass given in env var 'MONGO_BAL_PASS'")
	}
	if os.Getenv("MONGO_REVEL_PASS") == "" {
		panic("Running in prod, but no balancer pass given in env var 'MONGO_REVEL_PASS'")
	}
	state := core.NewStateWithMongo()
	fmt.Println("===STARTING COURTESY AMOUNT===")
	applyCourtesyAmount(state)
	fmt.Println("===FINISHED COURTESY AMOUNT===")
	fmt.Println("===STARTING ALPHA AMOUNT===")
	applyAlphaDiscount(state)
	fmt.Println("===FINISHED ALPHA AMOUNT===")
	fmt.Println("===STARTING SURVEY AMOUNT===")
	applySurveyDiscount(state)
	fmt.Println("===FINISHED SURVEY AMOUNT===")
}

func applyCourtesyAmount(state *core.State) {
	users, err := state.FetchAllUsers()
	if err != nil {
		panic(fmt.Sprintf("Unable to fetch all users: %s", err.Error()))
	}
	for _, u := range users {
		btcAmount := getCourtesyAmount(u.Username, state)
		err = state.MakePayment(u.Username, payment.Paid{
			BTCPaid:     btcAmount,
			Username:    u.Username,
			PaymentDate: time.Now(),
			Code:        "Alpha User Courtesy",
		})
		if err != nil {
			fmt.Printf("ERROR ADDING USER PAYMENT: %s\n", u.Username)
		}
	}
}

func applyAlphaDiscount(state *core.State) {
	users, err := state.FetchAllUsers()
	if err != nil {
		panic(fmt.Sprintf("Unable to fetch all users: %s", err.Error()))
	}

	for _, u := range users {
		//should be in percentage reduction
		alphaDiscount := 0.02
		_, apiErr := state.AddCustomChargeReduction(u.Username, fmt.Sprintf("%f", alphaDiscount), "Alpha User")
		if apiErr != nil {
			fmt.Println("Error adding user[%s] alpha discount: %s", u.Username, apiErr.LogError.Error())
		}
	}
}

func applySurveyDiscount(state *core.State) {
	for _, email := range SURVEY_DISCOUNT {
		//should be in percentage reduction
		surveyDiscount := 0.01
		_, apiErr := state.AddCustomChargeReduction(email, fmt.Sprintf("%f", surveyDiscount), "Took Payment Survey")
		if apiErr != nil {
			fmt.Println("Error adding user[%s] survey discount: %s", email, apiErr.LogError.Error())
		}
	}
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

func getCourtesyAmount(email string, state *core.State) int64 {
	//STEVE HERE
	poloUserStats, err := state.GetUserStatistics(email, 2, "polo")
	if err != nil {
		return 220000
	}
	stats := Collapse(poloUserStats)

	total := (stats.BTCNotLent + stats.BTCLent)

	// .15% or 0.0022 BTC ($10)
	totalRewardSats := int64((total * 1e8) * 0.0015)

	if totalRewardSats < 220000 {
		totalRewardSats = 220000
	}

	return totalRewardSats
}

func Collapse(data [][]userdb.AllUserStatistic) *CurrentUserStatistics {
	today := newCurrentUserStatistics()

	l := len(data)
	if l > 0 && len(data[0]) > 0 {
		now := data[0][0]

		totalAct := float64(0)
		for _, v := range now.Currencies {
			today.LoanRate += v.AverageActiveRate * (v.ActiveLentBalance * v.BTCRate)
			totalAct += v.ActiveLentBalance * v.BTCRate
			today.BTCLent += v.ActiveLentBalance * v.BTCRate
			today.BTCNotLent += (v.OnOrderBalance + v.AvailableBalance) * v.BTCRate
		}
		today.LoanRate = today.LoanRate / totalAct

		today.LendingPercent = today.BTCLent / (today.BTCLent + today.BTCNotLent)

		yesterday := userdb.GetCombinedDayAverage(data[1])
		if yesterday != nil {
			today.LoanRateChange = today.LoanRate - yesterday.LoanRate
			today.BTCLentChange = today.BTCLent - yesterday.Lent
			today.BTCNotLentChange = today.BTCNotLent - yesterday.NotLent
			today.LendingPercentChange = today.LendingPercent - yesterday.LendingPercent
		}
	}
	return today
}

const shortForm = "2006-Jan-02"

func dateToTime(t string) time.Time {
	nt, err := time.Parse(shortForm, t)
	if err != nil {
		fmt.Printf("ERROR PARSING TIME: YOU BROKE IT: %s\n", err.Error())
		return time.Now()
	}
	return nt
}
