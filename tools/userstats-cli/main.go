package main

// Usage
//		userdb-cli -u USERNAME -l admin

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var TotalAmt = float64(10)

var _, _ = fmt.Println, os.Readlink

func main() {
	var (
		username  = flag.String("u", "", "Username to change level of")
		populate  = flag.Bool("p", false, "Populate DB")
		dailyavgs = flag.Bool("d", false, "daily averages")
		polo      = flag.Bool("polo", false, "Polo data")
		fix       = flag.Bool("f", false, "Fix polo data")
	)

	flag.Parse()

	db, err := userdb.NewUserStatisticsDB()
	if err != nil {
		panic(err)
	}

	if *fix {
		PoloFix(db)
		return
	}

	if *polo {
		PoloData(db)
		return
	}

	if *username == "" {
		panic("No user")
	}

	if *populate {
		Populate(*username, db)
		return
	}

	if *dailyavgs {
		DailyAverages(*username, db)
		return
	}

}

func PoloFix(db *userdb.UserStatisticsDB) {
	db.Fix()
}

func PoloData(db *userdb.UserStatisticsDB) {

	stats := db.GetPoloniexStatistics()
	fmt.Println(stats)
}

func Populate(username string, db *userdb.UserStatisticsDB) {
	for i := 0; i < 31; i++ {
		stats := RandStats()
		stats.Username = username
		stats.Time = time.Now().Add(100 * time.Duration(i) * time.Second)
		stats.Currency = "BTC"
		db.CurrentIndex = i
		stats.TotalCurrencyMap["BTC"] = 1
		stats.TotalCurrencyMap["FCT"] = 0.5
		db.RecordData(stats)
		stats.Time = time.Now().Add(500 * time.Duration(i) * time.Second)
		db.RecordData(stats)

		db.RecordPoloniexStatisticTime(stats.AverageActiveRate, time.Now().Add(time.Duration(-i)*time.Minute))
	}
}

func DailyAverages(username string, db *userdb.UserStatisticsDB) {
	db.CurrentIndex = 30

	stats, err := db.GetStatistics(username, 2)
	if err != nil {
		panic(err)
	}

	i := 0
	for _, st := range stats {
		da := userdb.GetDayAvg(st)
		fmt.Println(da)
		i++
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

func RandStats() *userdb.UserStatistic {
	left := TotalAmt

	stats := userdb.NewUserStatistic()
	p := randomFloat(0, left*100) / 100
	left -= p
	stats.AvailableBalance = p

	p = randomFloat(0, left*100) / 100
	left -= p
	stats.ActiveLentBalance = p

	stats.OnOrderBalance = left

	stats.AverageActiveRate = randomFloat(0.001, 0.002)
	stats.AverageOnOrderRate = randomFloat(0.002, 0.0025)

	return stats
}

func randomFloat(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}
