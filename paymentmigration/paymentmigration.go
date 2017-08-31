package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/payment"
)

//all individuals who did the survey
var SURVEY_DISCOUNT = []string{
	"admin@admin.com",
}

func main() {
	state := core.NewStateWithMongo()
	mongoBalPass := os.Getenv("MONGO_BAL_PASS")
	if mongoBalPass == "" {
		panic("Running in prod, but no balancer pass given in env var 'MONGO_BAL_PASS'")
	}
	if os.Getenv("MONGO_REVEL_PASS") == "" {
		panic("Running in prod, but no revel pass given in env var 'MONGO_REVEL_PASS'")
	}
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
		btcAmount := getCourtesyAmount(u.Username)
		err = state.MakePayment(u.Username, payment.Paid{
			BTCPaid:     btcAmount,
			Username:    u.Username,
			PaymentDate: time.Now(),
			Code:        "Alpha User Courtesy",
		})
		if err != nil {
			fmt.Println("ERROR ADDING USER PAYMENT: %s", u.Username)
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
		alphaDiscount := 0.1
		switch {
		case u.StartTime.UTC().Nanosecond() > dateToTime("2017-May-01").UTC().Nanosecond():
			alphaDiscount = 0.05
		}

		_, apiErr := state.AddCustomChargeReduction(u.Username, fmt.Sprintf("%f", alphaDiscount), "Alpha User")
		if apiErr != nil {
			fmt.Println("Error adding user[%s] alpha discount: %s", u.Username, apiErr.LogError.Error())
		}
	}
}

func applySurveyDiscount(state *core.State) {
	users, err := state.FetchAllUsers()
	if err != nil {
		panic(fmt.Sprintf("Unable to fetch all users: %s", err.Error()))
	}

	for _, u := range users {
		//should be in percentage reduction
		surveyDiscount := 0.1
		switch {
		case u.StartTime.UTC().Nanosecond() > dateToTime("2017-May-01").UTC().Nanosecond():
			surveyDiscount = 0.05
		}

		_, apiErr := state.AddCustomChargeReduction(u.Username, fmt.Sprintf("%f", surveyDiscount), "Took Payment Survey")
		if err != nil {
			fmt.Println("Error adding user[%s] survey discount: %s", u.Username, apiErr.LogError.Error())
		}
	}
}

func getCourtesyAmount(email string) int64 {
	//STEVE HERE
	return 0
}

const shortForm = "2006-Jan-02"

func dateToTime(t string) time.Time {
	nt, err := time.Parse(shortForm, t)
	if err != nil {
		fmt.Println("ERROR PARSING TIME: YOU BROKE IT: %s", err.Error())
		return time.Now()
	}
	return nt
}
