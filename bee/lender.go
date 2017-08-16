package bee

import (
	"fmt"
	//"math"
	//"sort"
	//"strings"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var lendLogger = log.WithFields(log.Fields{
	"pacakge": "bee",
	"file":    "Lender",
})

type Lender struct {
	Users []*LendUser
	Bee   *Bee
	Polo  *balancer.PoloniexAPIWithRateLimit

	LendingRatesChannel chan map[int]map[string]balancer.LoanRates
	TickerChannel       chan map[string]poloniex.PoloniexTicker

	loanrateLock       sync.RWMutex
	currentLoanRate    map[int]map[string]balancer.LoanRates
	LastLoanRateUpdate time.Time

	tickerlock       sync.RWMutex
	ticker           map[string]poloniex.PoloniexTicker
	LastTickerUpdate time.Time

	quit chan bool

	cycles         int
	polousercycles int
	bitusercycles  int

	BitfinLender  *BitfinexLender
	PoloLender    *PoloniexLender
	HistoryKeeper *LendingHistoryKeeper
}

func (l *Lender) SavePoloniexMonth(username *userdb.User, accesskey, secretkey string) bool {
	return l.HistoryKeeper.SavePoloniexMonth(username, accesskey, secretkey)
}

func (l *Lender) SaveBitfinexMonth(username, accesskey, secretkey string) bool {
	return l.HistoryKeeper.SaveBitfinexMonth(username, accesskey, secretkey)
}

func (l *Lender) SetPoloniexTicker(currency string, t poloniex.PoloniexTicker) {
	l.tickerlock.Lock()
	l.ticker[currency] = t
	l.tickerlock.Unlock()
}

func (l *Lender) GetPoloniexTicker(currency string) (poloniex.PoloniexTicker, bool) {
	l.tickerlock.RLock()
	defer l.tickerlock.RUnlock()
	v, ok := l.ticker[currency]
	return v, ok
}

func (l *Lender) GetLoanRate(exch int, currency string) (balancer.LoanRates, bool) {
	l.loanrateLock.RLock()
	defer l.loanrateLock.RUnlock()
	v, ok := l.currentLoanRate[exch][currency]
	return v, ok
}

func (l *Lender) SetTicker(t map[string]poloniex.PoloniexTicker) {
	l.ticker = t
}

func NewLender(b *Bee) *Lender {
	l := new(Lender)
	l.Bee = b
	l.Polo = balancer.NewPoloniexAPIWithRateLimit()

	l.LendingRatesChannel = make(chan map[int]map[string]balancer.LoanRates, 100)
	l.TickerChannel = make(chan map[string]poloniex.PoloniexTicker, 100)

	l.ticker = make(map[string]poloniex.PoloniexTicker)
	l.currentLoanRate = make(map[int]map[string]balancer.LoanRates)
	l.currentLoanRate[balancer.PoloniexExchange] = make(map[string]balancer.LoanRates)
	l.currentLoanRate[balancer.BitfinexExchange] = make(map[string]balancer.LoanRates)
	l.HistoryKeeper = NewLendingHistoryKeeper(b)
	l.BitfinLender = NewBitfinexLender(b, l)
	l.PoloLender = NewPoloniexLender(b, l, l.Polo)

	return l
}

func (l *Lender) Report() string {
	return fmt.Sprintf("Cycles: %d, PoloUsersProcesses: %d, BitUserProcesses %d,", l.cycles, l.polousercycles, l.bitusercycles)
}

type LendUser struct {
	U balancer.User
}

func (*LendUser) Prefix() string {
	return fmt.Sprintf("")
}

func (l *Lender) Runloop() {
	go l.BitfinLender.Run()
	for {
		startLoop := time.Now()
		// Process all users
		for _, u := range l.Users {
			// Find the latest update
			var latest map[int]map[string]balancer.LoanRates
			var ticker map[string]poloniex.PoloniexTicker
		LatestRate:
			for {
				select {
				case lr := <-l.LendingRatesChannel:
					latest = lr
				case tc := <-l.TickerChannel:
					ticker = tc
				default:
					break LatestRate
				}
			}

			// Update our rates
			if len(latest) > 0 {
				l.loanrateLock.Lock()
				l.currentLoanRate = latest
				l.LastLoanRateUpdate = time.Now()
				l.loanrateLock.Unlock()
			}

			// Update our ticker
			if len(ticker) > 0 {
				l.tickerlock.Lock()
				l.ticker = ticker
				l.LastTickerUpdate = time.Now()
				l.tickerlock.Unlock()
			}

			// Process User
			duration := time.Now()

			switch u.U.Exchange {
			case balancer.PoloniexExchange:
				err := l.PoloLender.ProcessPoloniexUser(u)
				u.U.LastTouch = time.Now()
				if err != nil {
					lendLogger.WithFields(log.Fields{"func": "ProcessPoloniexUser", "user": u.U.Username,
						"exchange": balancer.GetExchangeString(u.U.Exchange)}).Errorf("[PoloLending] Error: %s", shortError(err).Error())
				}
				l.polousercycles++
			case balancer.BitfinexExchange:
				err := l.BitfinLender.ProcessBitfinexUser(u)
				u.U.LastTouch = time.Now()
				if err != nil {
					lendLogger.WithFields(log.Fields{"func": "ProcessBitfinexUser", "user": u.U.Username,
						"exchange": balancer.GetExchangeString(u.U.Exchange)}).Errorf("[BitfinexLending] Error: %s", shortError(err).Error())
				}
				l.bitusercycles++
			}
			JobProcessDuration.Observe(float64(time.Since(duration).Nanoseconds()))
		}

		// Update User List
		l.CopyBeeList()

		// Quit
		select {
		case <-l.quit:
			l.quit <- true
			return
		default:
		}

		l.cycles++
		took := time.Since(startLoop).Seconds()
		if took < 10 {
			time.Sleep(time.Duration(10-took) * time.Second)
		}
	}
}

func (l *Lender) CopyBeeList() {
	l.Bee.userlock.RLock()
	l.Users = make([]*LendUser, len(l.Bee.Users))
	for i := range l.Bee.Users {
		l.Users[i] = &LendUser{}
		l.Users[i].U = *l.Bee.Users[i]
	}
	l.Bee.userlock.RUnlock()
}

func shortError(err error) error {
	if len(err.Error()) > 100 {
		return fmt.Errorf("%s...", err.Error()[:100])
	}
	return err
}

func (l *Lender) GetAmtForBTCValue(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	if currency == "IOT" {
		return amount / l.BitfinLender.iotLast
	}
	if currency == "EOS" {
		return amount / l.BitfinLender.eosLast
	}

	l.tickerlock.RLock()
	t, ok := l.ticker[fmt.Sprintf("BTC_%s", currency)]
	l.tickerlock.RUnlock()
	if !ok {
		return amount
	}

	return amount / t.Last
}

func (l *Lender) GetBTCAmount(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	if currency == "IOT" {
		return amount * l.BitfinLender.iotLast
	}
	if currency == "EOS" {
		return amount * l.BitfinLender.eosLast
	}

	l.tickerlock.RLock()
	t, ok := l.ticker[fmt.Sprintf("BTC_%s", currency)]
	l.tickerlock.RUnlock()
	if !ok {
		return 0
	}

	return t.Last * amount
}

func abs(v float64) float64 {
	if v < 0 {
		return v * -1
	}
	return v
}
