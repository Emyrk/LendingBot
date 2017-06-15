package userdb_test

import (
	//"crypto"
	// "fmt"
	//"os"
	"fmt"
	"math"
	"testing"
	"time"

	//"github.com/DistributedSolutions/twofactor"
	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Println

// func TestUserStat(t *testing.T) {
// 	stats := NewAll("BTC", 1)
// 	// stats.Username = "steven"
// 	stats.AvailableBalance = 100
// 	stats.ActiveLentBalance = 100
// 	stats.OnOrderBalance = 100
// 	stats.AverageActiveRate = .4
// 	stats.AverageOnOrderRate = .1

// 	// stats.TotalCurrencyMap["BTC"] = 1.2

// 	data, err := stats.MarshalMsg()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	u2 := NewUserStatistic()
// 	data, err = u2.UnmarshalBinaryData(data)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if len(data) > 0 {
// 		t.Error("Should be length 0")
// 	}

// 	if !stats.IsSameAs(u2) {
// 		t.Error("Should be same")
// 	}
// }

func TestS(t *testing.T) {
	db, err := NewUserStatisticsDB()
	if err != nil {
		t.Error(err)
	}

	userStats, err := db.GetStatistics("stevenmasley@gmail.com", 2)

	// balanceDetails := newUserBalanceDetails()
	today := newCurrentUserStatistics()
	// if err != nil {
	// balanceDetails.compute()
	// return today, balanceDetails
	// }
	l := len(userStats)
	if l > 0 && len(userStats[0]) > 0 {
		now := userStats[0][0]
		// Set balance ratios
		// balanceDetails.CurrencyMap = now.TotalCurrencyMap
		// balanceDetails.compute()

		totalAct := float64(0)
		for _, v := range now.Currencies {
			today.LoanRate += v.AverageActiveRate * (v.ActiveLentBalance * v.BTCRate)
			totalAct += v.ActiveLentBalance * v.BTCRate
			today.BTCLent += v.ActiveLentBalance * v.BTCRate
			today.BTCNotLent += (v.OnOrderBalance + v.AvailableBalance) * v.BTCRate
		}
		today.LoanRate = today.LoanRate / totalAct

		today.LendingPercent = today.BTCLent / (today.BTCLent + today.BTCNotLent)

		yesterday := GetCombinedDayAverage(userStats[1])
		if yesterday != nil {
			today.LoanRateChange = today.LoanRate - yesterday.LoanRate
			today.BTCLentChange = today.BTCLent - yesterday.Lent
			today.BTCNotLentChange = today.BTCNotLent - yesterday.NotLent
			today.LendingPercentChange = today.LendingPercent - yesterday.LendingPercent
		}
	}

	fmt.Println(today)
	fmt.Println(today.LendingPercent)
}

func TestGetDay(t *testing.T) {
	return
	ti := time.Now()
	for i := 0; i < 100000; i++ {
		last := GetDay(ti)
		ti = ti.Add(time.Duration(1*24) * time.Hour)
		next := GetDay(ti)
		if next-last != 1 {
			t.Errorf("Next should be 1, found %d :: %v", next-last, ti)
		}
	}
}

/*
type UserStatistic struct {
	Username           string    `json:"username"`
	AvailableBalance   float64   `json:"availbal"`
	ActiveLentBalance  float64   `json:"availlent"`
	OnOrderBalance     float64   `json:"onorder"`
	AverageActiveRate  float64   `json:"activerate"`
	AverageOnOrderRate float64   `json:"onorderrate"`
	Time               time.Time `json:"time"`
	Currency           string    `json:"currency"`

	day int
}
*/

func TestGetDayAvg(t *testing.T) {
	return
	u, _ := NewUserStatisticsMapDB()
	var _ = u

	stats := NewAllUserStatistic()
	b := NewUserStatistic("BTC", 1)
	// stats.Username = "steven"
	b.AvailableBalance = 0
	b.ActiveLentBalance = 100
	b.OnOrderBalance = 0
	b.AverageActiveRate = .4
	b.AverageOnOrderRate = .1
	stats.Currencies["BTC"] = b
	stats.Time = time.Now()
	// stats.Currency["BTC"] = b

	var _ = stats

	u.RecordData(stats)
	stats.Currencies["BTC"].AvailableBalance = 0
	stats.Time = stats.Time.Add(5 * time.Second)
	u.RecordData(stats)
	// u.RecordData(stats)

	ustats, _ := u.GetStatistics("steven", 1)
	da := GetCombinedDayAverage(ustats[0])
	if da.LendingPercent != 1 {
		t.Error("Should be 1")
	}
}

func TestAvgAndStd(t *testing.T) {
	return
	var sample []PoloniexRateSample
	for i := float64(0); i < 13; i++ {
		sample = append(sample, PoloniexRateSample{0, i})
	}

	avg, std := GetAvgAndStd(sample)
	if fmt.Sprintf("%.3f", std) != "3.894" {
		t.Errorf("[Std] Exp: %f, Found %f", 3.894440482, std)
	}
	if avg != 6 {
		t.Errorf("[Avg] Exp: %f, Found %f", 6.0, avg)
	}
}

// func TestThisThing(t *testing.T) {
// 	thingy := func(i int, offset int) int {
// 		i += offset
// 		if i > 30 {
// 			overFlow := i - 30
// 			i = -1 + overFlow
// 		}

// 		if i < 0 {
// 			underFlow := i * -1
// 			i = 31 - underFlow
// 		}
// 		return i
// 	}

// 	for i := 0; i < 100; i++ {
// 		fmt.Println(thingy(1, -1*(i%30)))
// 	}

// }

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
