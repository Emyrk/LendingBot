package controllers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/Emyrk/LendingBot/src/lender"
	"github.com/Emyrk/LendingBot/src/queuer"

	// For Prometheus
	"github.com/prometheus/client_golang/prometheus"
	"net/http"

	// Init logger
	_ "github.com/Emyrk/LendingBot/src/log"

	"github.com/revel/revel"
)

const (
	DEV_FAKE  = "devFake"
	DEV_EMPTY = "devEmpty"
)

var _ = revel.Equal

var Queuer *queuer.Queuer
var Lender *lender.Lender

func Launch() {
	// Prometheus
	lender.RegisterPrometheus()
	queuer.RegisterPrometheus()

	fmt.Println("MODE IS: ", revel.RunMode)
	if revel.RunMode == DEV_FAKE {
		//devFake mode
		//should be all in memory with user account

		state = core.NewStateWithMap()

		//user: a@a.com pass:a
		//should be commonuser level
		//should have populated data
		state.NewUser("a@a.com", "a")
		Populate("a@a.com", state.GetUserStatsDB())

		//user: b@b.com pass:b
		//should be commonuser level
		//should have empty data
		state.NewUser("b@b.com", "b")

		//user: admin@admin.com pass:admin
		//should be sysadmin level
		//should have populated data
		state.NewUser("admin@admin.com", "admin")
		state.UpdateUserPrivilege("admin@admin.com", "SysAdmin")
		Populate("admin@admin.com", state.GetUserStatsDB())

		//should be mainly used for gui creation
	} else if revel.RunMode == DEV_EMPTY {
		//devEmpty mode
		//should be all in memory with empty data

		state = core.NewStateWithMap()
		state.NewUser("admin@admin.com", "admin")
		state.UpdateUserPrivilege("admin@admin.com", "SysAdmin")

		//to be used for unit testing/regression testing
	} else {
		state = core.NewState()
	}

	err := state.VerifyState()
	if err != nil {
		panic(err)
	}

	lenderBot := lender.NewLender(state)
	queuerBot := queuer.NewQueuer(state, lenderBot)

	Queuer = queuerBot
	Lender = lenderBot

	// Start go lending
	go lenderBot.Start()
	go queuerBot.Start()
	go launchPrometheus(9911)
}

func Shutdown() {
	state.Close()
	Lender.Close()
	Queuer.Close()
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

var curarr = []string{"BTC", "BTS", "CLAM", "DOGE", "DASH", "LTC", "MAID", "STR", "XMR", "XRP", "ETH", "FCT"}

func Populate(username string, db *userdb.UserStatisticsDB) {
	for i := 0; i < 31; i++ {
		stats := RandStats()
		stats.Username = username
		stats.Time = time.Now().Add(100 * time.Duration(i) * time.Second)
		// stats.Currency = "BTC"
		db.CurrentIndex = i
		stats.TotalCurrencyMap["BTC"] = 1
		stats.TotalCurrencyMap["FCT"] = 0.3
		stats.TotalCurrencyMap["CLAM"] = 0.1
		stats.TotalCurrencyMap["ETH"] = 0.8
		stats.TotalCurrencyMap["DOGE"] = 0.05
		db.RecordData(stats)
		stats.Time = time.Now().Add(500 * time.Duration(i) * time.Second)
		db.RecordData(stats)

		db.RecordPoloniexStatisticTime("BTC", stats.Currencies["BTC"].AverageActiveRate, time.Now().Add(time.Duration(-i)*time.Minute))
	}

	n := time.Now()
	for i := 0; i < 31; i++ {
		n = n.Add(-24 * time.Hour)
		data := RandomLendingHistoryData(username)
		data.SetTime(n)
		db.SaveLendingHistory(data)
	}
}

func RandStats() *userdb.AllUserStatistic {
	stats := userdb.NewAllUserStatistic()

	for _, v := range curarr {
		r := .1
		if v == "BTC" {
			r = 1
		}
		s := userdb.NewUserStatistic(v, r)
		left := float64(10)

		p := randomFloat(left*80, left*100) / 100
		left -= p
		s.ActiveLentBalance = p

		p = randomFloat(0, left*100) / 100
		left -= p
		s.AvailableBalance = p

		s.OnOrderBalance = left

		s.AverageActiveRate = randomFloat(0.001, 0.002)
		s.AverageOnOrderRate = randomFloat(0.002, 0.0025)

		stats.Currencies[v] = s
	}

	return stats
}

func randomFloat(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func RandomLendingHistoryData(username string) *userdb.AllLendingHistoryEntry {
	all := userdb.NewAllLendingHistoryEntry()
	for _, v := range curarr {
		d := new(userdb.LendingHistoryEntry)
		all.Data[v] = d
		d.Currency = v
		d.AvgDuration = randomFloat(0, 2)
		interest := randomFloat(0.1, 0.005)
		d.Earned = interest * .85
		d.Fees = interest * .15
		d.LoanCounts = rand.Intn(200)
	}
	all.Username = username
	return all
}
