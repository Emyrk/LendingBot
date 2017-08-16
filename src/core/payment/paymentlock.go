package payment

import (
	"sync"
	"time"
)

// Need to lock when deleting or adding

type MapLock struct {
	sync.RWMutex
	locks map[string]*PaymentLock
}

func (l *MapLock) Set(key string, pl *PaymentLock) {
	l.Lock()
	defer l.Unlock()

	pl.LastUpdated = time.Now().UTC()

	l.locks[key] = pl
}

func (l *MapLock) Get(key string) (*PaymentLock, bool) {
	l.RLock()
	defer l.RUnlock()

	pl, ok := l.locks[key]
	if !ok {
		pl = &PaymentLock{}
	}

	pl.LastAccessed = time.Now().UTC()
	return pl, ok
}

type PaymentLock struct {
	sync.RWMutex
	LastAccessed time.Time
	LastUpdated  time.Time
}
