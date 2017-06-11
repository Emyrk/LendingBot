package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	"net/http"
	"time"
)

const (
	CACHE_TIME          = 10 * time.Minute
	CACHE_TIME_POLONIEX = 15 * time.Minute
	SESSION_EMAIL       = "email"
)

func DeleteCacheToken(sessionId string) error {
	fmt.Printf("Deleting SessionID[%s]\n", sessionId)
	go cache.Set(sessionId, "", 1*time.Second)
	go cache.Delete(sessionId)
	return nil
}

func SetCacheEmail(sessionId string, email string) error {
	fmt.Printf("Set SessionID[%s], email[%s]\n", sessionId, email)
	go cache.Set(sessionId, email, CACHE_TIME)
	return nil
}

func ValidCacheEmail(sessionId string, email string) bool {
	var e string
	if err := cache.Get(sessionId, &e); err != nil {
		return false
	}

	fmt.Printf("Comparing strings [%s]s, [%s]\n", e, email)

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
