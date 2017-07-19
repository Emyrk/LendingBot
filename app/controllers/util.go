package controllers

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	log "github.com/sirupsen/logrus"
	// "gopkg.in/mgo.v2/bson"
)

var utilLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "util",
})

type CacheSession struct {
	Email           string
	LastRenewalTime time.Time
}

const (
	CACHE_TIME           = 10 * time.Minute
	CACHE_TIME_POLONIEX  = 15 * time.Minute
	SESSION_EMAIL        = "email"
	CACHE_LEND_HIST_TIME = 2 * time.Hour
	CACHE_LENDING_ENDING = "_lendHist"

	CACHE_TIME_USER_DUR   = 12 * time.Hour
	CACHE_USER_DUR_ENDING = "_durEnd"
)

func SetCacheDurEnd(sessionId, ip, email string, expiryDur time.Duration) {
	llog := utilLog.WithField("method", "SetCacheDurEnd")

	go func() {
		err := cache.Set(email+CACHE_USER_DUR_ENDING, expiryDur, CACHE_TIME_USER_DUR)
		if err != nil {
			llog.Errorf("Error setting user [%s] expiry duration: %s", email, err.Error())
		}
	}()
}

func GetCacheDurEnd(sessionId, email string, ip net.IP) (*time.Duration, error) {
	expiryDur := userdb.DEFAULT_SESSION_DUR
	if err := cache.Get(email+CACHE_USER_DUR_ENDING, &expiryDur); err != nil {
		u, err := state.FetchUser(email)
		if err != nil {
			return nil, fmt.Errorf("Error fetching user: %s", email)
		}
		if u == nil {
			return nil, fmt.Errorf("Error user is nil: %s", email)
		}
		expiryDur = u.SessionExpiryTime
	}
	return &expiryDur, nil
}

func DeleteCacheToken(sessionId, ip, email string) error {
	go cache.Set(sessionId, "", 1*time.Second)
	go cache.Delete(sessionId)
	state.CloseUserSession(sessionId, email, net.ParseIP(ip))
	return nil
}

func SetCacheEmail(sessionId, ip, email string) error {
	llog := utilLog.WithField("method", "SetCacheEmail")

	expiryDur, err := GetCacheDurEnd(sessionId, email, net.ParseIP(ip))
	if err != nil {
		return fmt.Errorf("Error getting cache duration: ", err.Error())
	}

	newRenewalTime := time.Now().Add(*expiryDur).UTC()

	cs := CacheSession{email, newRenewalTime}
	err = cache.Set(sessionId, cs, CACHE_TIME)
	if err != nil {
		llog.Errorf("Error setting user [%s] session cache: %s", email, err.Error())
	}

	go func() {
		state.UpdateUserSession(sessionId, email, newRenewalTime, net.ParseIP(ip), true)
	}()
	return err
}

func ValidCacheEmail(sessionId, ip, email string) bool {
	llog := utilLog.WithField("method", "ValidCacheEmail")

	var (
		expiryDur *time.Duration
		err       error
	)

	//grab user session expiry time
	//cant get cache go to mongo db session
	expiryDur, err = GetCacheDurEnd(sessionId, email, net.ParseIP(ip))
	if err != nil {
		llog.Errorf("Error getting cache: %s", err.Error())
		return false
	}

	var cacheSes CacheSession
	if err = cache.Get(sessionId, &cacheSes); err != nil {
		//if cant get session
		ses := state.GetUserSession(sessionId, email, net.ParseIP(ip))
		if ses == nil {
			return false
		}

		if ses.LastRenewalTime.Add(*expiryDur).UTC().Format(userdb.SESSION_FORMAT) < time.Now().UTC().Format(userdb.SESSION_FORMAT) {
			if ses.Open == true {
				//close session if it says open and time is incorrect. more for logging than anything else
				go func() {
					state.CloseUserSession(sessionId, email, net.ParseIP(ip))
				}()
			}
			return false
		}

		if ses.Email != email {
			llog.Errorf("Error should not happen, email should not be here [%s] [%s]", ses.Email, email)
			return false
		}

		if ses.Open == false {
			return false
		}

		return true
	}

	if cacheSes.Email != email {
		llog.Errorf("Error should not happen emails are not the same [%s] [%s], sessionId [%s], deleting/closing user sessions", cacheSes.Email, email, sessionId)
		go cache.Delete(sessionId)
		go func() {
			state.CloseUserSession(sessionId, email, net.ParseIP(ip))
		}()
		return false
	}

	if cacheSes.LastRenewalTime.Add(*expiryDur).UTC().Format(userdb.SESSION_FORMAT) < time.Now().UTC().Format(userdb.SESSION_FORMAT) {
		go cache.Delete(sessionId)
		go func() {
			state.CloseUserSession(sessionId, email, net.ParseIP(ip))
		}()
		return false
	}

	return true
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
