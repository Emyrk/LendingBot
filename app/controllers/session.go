package controllers

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	log "github.com/sirupsen/logrus"
)

var sessionLog = log.WithFields(log.Fields{
	"package": "controllers",
	"file":    "session",
})

type CacheSession struct {
	Email    string
	Sessions map[string]time.Time
	Expiry   time.Duration
}

const (
	SESSION_EMAIL = "email"

	CACHE_TIME_USER_SESSION_MIN = 5 * time.Minute
	CACHE_TIME_USER_SESSION_MAX = 5 * time.Hour

	CACHE_TIME_POLONIEX = 15 * time.Minute

	CACHE_LEND_HIST_TIME = 2 * time.Hour
	CACHE_LENDING_ENDING = "_lendHist"
)

func getTimeoutCookie(dur time.Duration) *http.Cookie {
	t := time.Now().Add(dur)

	timeoutCookie := &http.Cookie{
		Name:    "HODL_TIMEOUT",
		Value:   fmt.Sprintf("%d", t.Unix()),
		Domain:  revel.CookieDomain,
		Path:    "/",
		Expires: t.UTC(),
		Secure:  revel.CookieSecure,
		MaxAge:  int(dur.Seconds()),
	}
	return timeoutCookie
}

func SetCacheDurEnd(email string, expiryDur time.Duration) error {
	var (
		cacheSes CacheSession
		err      error
	)
	if err = cache.Get(email, &cacheSes); err != nil {
		//did not find user sessions, make session
		return fmt.Errorf("Error finding user[%s] cache session: %s", email, err.Error())
	}
	cacheSes.Expiry = expiryDur
	if err = cache.Set(email, cacheSes, CACHE_TIME_USER_SESSION_MAX); err != nil {
		return fmt.Errorf("Error setting user [%s] session cache: %s", email, err.Error())
	}
	return nil
}

func DeleteCacheToken(sessionId, ip, email string) error {
	var (
		cacheSes CacheSession
		err      error
	)
	if err = cache.Get(email, &cacheSes); err != nil {
		//did not find user sessions
		return fmt.Errorf("Error fetching user[%s] to delete new session: %s", email, err.Error())
	} else {
		//found user sessions
		state.WriteSession(sessionId, email, time.Now(), net.ParseIP(ip), false)
		delete(cacheSes.Sessions, sessionId)
		if len(cacheSes.Sessions) == 0 {
			err := cache.Delete(email)
			if err != nil {
				return fmt.Errorf("Error with deleting user[%s] sessions: %s", email, err.Error())
			}
		}
	}
	return nil
}

func SetCacheEmail(sessionId, ip, email string) (*http.Cookie, error) {
	var (
		cacheSes CacheSession
		err      error
		now      time.Time
	)
	if err = cache.Get(email, &cacheSes); err != nil {
		//did not find user sessions, make session
		u, err := state.FetchUser(email)
		if err != nil {
			return nil, fmt.Errorf("Error fetching user[%s] to create new session: %s", email, err.Error())
		}
		m := make(map[string]time.Time)
		now = time.Now().UTC()
		m[sessionId] = now
		cacheSes = CacheSession{Email: email, Sessions: m, Expiry: u.SessionExpiryTime}
	} else {
		//found session update it
		now = time.Now().UTC()
		cacheSes.Sessions[sessionId] = now
	}
	state.WriteSession(sessionId, email, now, net.ParseIP(ip), true)
	if err = cache.Set(email, cacheSes, CACHE_TIME_USER_SESSION_MAX); err != nil {
		return nil, fmt.Errorf("Error setting user [%s] session cache: %s", email, err.Error())
	}
	return getTimeoutCookie(cacheSes.Expiry), nil
}

func ValidCacheEmail(sessionId, ip, email string) bool {
	llog := sessionLog.WithField("method", "ValidCacheEmail")

	var (
		cacheSes CacheSession
		err      error
	)
	if err = cache.Get(email, &cacheSes); err != nil {
		//did not find user sessions
		llog.Errorf("Error fetching user[%s] ip[%s] session cache to validate: %s", email, ip, err.Error())
		return false
	} else {
		//found session update it
		sesLastUpdateTime := cacheSes.Sessions[sessionId]
		now := time.Now().UTC()
		if sesLastUpdateTime.Add(cacheSes.Expiry).Format(userdb.SESSION_FORMAT) < now.Format(userdb.SESSION_FORMAT) {
			fmt.Println("SESSIONS LEFT ", len(cacheSes.Sessions), cacheSes, len(cacheSes.Sessions) == 0)
			delete(cacheSes.Sessions, sessionId)
			state.WriteSession(sessionId, email, time.Now(), net.ParseIP(ip), false)
			fmt.Println("SESSIONS LEFT ", len(cacheSes.Sessions), cacheSes, len(cacheSes.Sessions) == 0)
			if len(cacheSes.Sessions) == 0 {
				err := cache.Delete(email)
				if err != nil {
					llog.Errorf("Error with deleting user[%s] sessions: %s", email, err.Error())
				}
				return false
			}
			if err = cache.Set(email, cacheSes, CACHE_TIME_USER_SESSION_MAX); err != nil {
				llog.Errorf("Error setting user [%s] session cache: %s", email, err.Error())
				return false
			}
			llog.Infof("Info session user[%s] ip[%s] no longer valid saved time[%s], given time[%s], expire duration[%d], new saved plus[%s]", email, ip, sesLastUpdateTime.Format(userdb.SESSION_FORMAT), now.Format(userdb.SESSION_FORMAT), cacheSes.Expiry, sesLastUpdateTime.Add(cacheSes.Expiry).Format(userdb.SESSION_FORMAT))
			return false
		}
	}

	return true
}
