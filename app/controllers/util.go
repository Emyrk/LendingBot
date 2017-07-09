package controllers

import (
	"fmt"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	log "github.com/sirupsen/logrus"

	"net/http"
	"time"
)

var utilLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "util",
})

const (
	CACHE_TIME           = 10 * time.Minute
	CACHE_TIME_POLONIEX  = 15 * time.Minute
	SESSION_EMAIL        = "email"
	CACHE_LEND_HIST_TIME = 2 * time.Hour
	CACHE_LENDING_ENDING = "_lendHist"
)

func DeleteCacheToken(sessionId string) error {
	llog := utilLog.WithField("method", "CurrentUserStats")
	llog.Infof("Deleting SessionID[%s]\n", sessionId)
	go cache.Set(sessionId, "", 1*time.Second)
	go cache.Delete(sessionId)
	return nil
}

func SetCacheEmail(sessionId string, email string) error {
	llog := utilLog.WithField("method", "SetCacheEmail")
	llog.Infof("Set SessionID[%s], email[%s]\n", sessionId, email)
	go cache.Set(sessionId, email, CACHE_TIME)
	return nil
}

func ValidCacheEmail(sessionId string, email string) bool {
	var e string
	if err := cache.Get(sessionId, &e); err != nil {
		time.Sleep(100 * time.Millisecond)
		if err := cache.Get(sessionId, &e); err != nil {
			return false
		}
	}

	llog := utilLog.WithField("method", "ValidCacheEmail")
	llog.Infof("Comparing strings [%s]s, [%s]\n", e, email)

	return e == email && len(e) > 0 && len(email) > 0
}

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

func GetTimeoutCookie() *http.Cookie {
	t := time.Now().Add(CACHE_TIME)

	timeoutCookie := &http.Cookie{
		Name:    "HODL_TIMEOUT",
		Value:   fmt.Sprintf("%d", t.Unix()),
		Domain:  revel.CookieDomain,
		Path:    "/",
		Expires: t.UTC(),
		Secure:  revel.CookieSecure,
		MaxAge:  int(CACHE_TIME.Seconds()),
	}
	return timeoutCookie
}

func CacheGetLendingHistory(email string) (*poloniex.PoloniexAuthentictedLendingHistoryRespone, bool) {
	llog := utilLog.WithField("method", "CacheGetLendingHistory")
	var poloniexHistory poloniex.PoloniexAuthentictedLendingHistoryRespone
	if err := cache.Get(email+CACHE_LENDING_ENDING, &poloniexHistory); err != nil {
		llog.Infof("NOT found cache lending history for user %s", email)
		return nil, false
	}
	llog.Infof("Found cache lending history for user %s\n", email)
	return &poloniexHistory, true
}

func CacheSetLendingHistory(email string, p poloniex.PoloniexAuthentictedLendingHistoryRespone) {
	llog := utilLog.WithField("method", "CacheGetLendingHistory")
	llog.Infof("Setting lending history for user %s", email)
	go cache.Set(email+CACHE_LENDING_ENDING, p, CACHE_LEND_HIST_TIME)
}
