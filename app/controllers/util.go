package controllers

import (
	"fmt"

	"github.com/globalsign/mgo"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel/cache"
	log "github.com/sirupsen/logrus"
)

var utilLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "util",
})

func percentChange(a float64, b float64) float64 {
	if a == 0 || b == 0 {
		return 0
	}
	change := ((a - b) / a) * 100
	if abs(change) < 0.001 {
		return 0
	}
	return change
}

func abs(a float64) float64 {
	if a < 0 {
		return a * -1
	}
	return a
}

func CacheGetLendingHistory(email string) (*poloniex.PoloniexAuthentictedLendingHistoryRespone, bool) {
	var poloniexHistory poloniex.PoloniexAuthentictedLendingHistoryRespone
	if err := cache.Get(email+CACHE_LENDING_ENDING, &poloniexHistory); err != nil {
		fmt.Printf("NOT found cache lending history for user %s", email)
		return nil, false
	}
	fmt.Printf("Found cache lending history for user %s\n", email)
	return &poloniexHistory, true
}

func CacheSetLendingHistory(email string, p poloniex.PoloniexAuthentictedLendingHistoryRespone) {
	fmt.Printf("Setting lending history for user %s", email)
	go cache.Set(email+CACHE_LENDING_ENDING, p, CACHE_LEND_HIST_TIME)
}

func GetUserActiveSessions(email, sessionId string) ([]userdb.Session, error) {
	var (
		err error
		uss []userdb.Session
	)
	cs, err := GetActiveSessions(email)
	if err != nil {
		if err.Error() == mgo.ErrNotFound.Error() {
			return uss, nil
		}
		return uss, fmt.Errorf("CRITICAL! This should never happend. Error retrieving user sessions: %s", err.Error())
	}
	uss, err = state.GetActiveSessions(email, cs.Sessions, sessionId)
	if err != nil {
		return uss, fmt.Errorf("Error retrieving user sessions from db: %s", err.Error())
	}
	return uss, nil
}
