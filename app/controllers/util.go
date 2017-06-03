package controllers

import (
	"fmt"
	"github.com/revel/revel/cache"
	"time"
)

const (
	CACHE_TIME      = 10 * time.Minute
	CACHE_TIME_TEST = 1 * time.Second
	SESSION_EMAIL   = "email"
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

	return e == email
}
