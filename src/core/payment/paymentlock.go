package payment

import (
	"sync"
	"time"
)

// Need to lock when deleting or adding

type MapLock struct {
	lock  sync.RWMutex
	locks map[string]*PaymentLock
}

func NewMapLock() *MapLock {
	m := new(MapLock)
	m.locks = make(map[string]*PaymentLock)

	return m
}

func (l *MapLock) Set(key string, pl *PaymentLock) {
	l.lock.Lock()
	defer l.lock.Unlock()

	pl.LastUpdated = time.Now().UTC()

	l.locks[key] = pl
}

func (l *MapLock) Get(key string) (*PaymentLock, bool) {
	l.lock.RLock()

	pl, ok := l.locks[key]
	l.lock.RUnlock()
	if !ok {
		pl = NewPaymentLock(key)
		l.lock.Lock()
		l.locks[key] = pl
		l.lock.Unlock()
	}

	pl.LastAccessed = time.Now().UTC()
	return pl, ok
}

func (l *MapLock) GetLocked(key string) (*PaymentLock, bool) {
	pl, ok := l.Get(key)
	pl.Lock()
	return pl, ok
}

func (l *MapLock) UnlockPayment(username string, pl *PaymentLock) {
	pl.LastUpdated = time.Now().UTC()
	pl.Unlock()
}

type PaymentLock struct {
	sync.RWMutex
	Username     string
	LastAccessed time.Time
	LastUpdated  time.Time
}

func NewPaymentLock(username string) *PaymentLock {
	p := new(PaymentLock)
	p.Username = username

	return p
}
