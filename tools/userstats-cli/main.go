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

var _, _ = fmt.Println, os.Readlink

func main() {
	var (
		username  = flag.String("u", "", "Username to change level of")
		populate  = flag.Bool("p", false, "Populate DB")
		dailyavgs = flag.Bool("d", false, "daily averages")
	)

	flag.Parse()

	db, err := userdb.NewUserStatisticsDB()
	if err != nil {
		panic(err)
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

func Populate(username string, db *userdb.UserStatisticsDB) {
	for i := 0; i < 31; i++ {
		stats := RandStats()
		stats.Username = username
		stats.Time = time.Now().Add(100 * time.Duration(i) * time.Second)
		stats.Currency = "BTC"
		db.CurrentIndex = i
		db.RecordData(stats)
		stats.Time = time.Now().Add(500 * time.Duration(i) * time.Second)
		db.RecordData(stats)
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
	stats := userdb.NewUserStatistic()
	stats.AvailableBalance = rand.Float64()
	stats.ActiveLentBalance = rand.Float64()
	stats.OnOrderBalance = rand.Float64()
	stats.AverageActiveRate = rand.Float64()
	stats.AverageOnOrderRate = rand.Float64()

	return stats
}
