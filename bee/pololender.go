package bee

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var poloLogger = log.WithFields(log.Fields{"package": "PoloLender"})

type Lender struct {
	Polo  *balancer.PoloniexAPIWithRateLimit
	Users []*LendUser
	Bee   *Bee

	recordMapLock sync.RWMutex
	recordMap     map[int]map[string]time.Time

	LendingRatesChannel chan map[int]map[string]balancer.LoanRates
	TickerChannel       chan map[string]poloniex.PoloniexTicker

	loanrateLock       sync.RWMutex
	currentLoanRate    map[int]map[string]balancer.LoanRates
	LastLoanRateUpdate time.Time

	tickerlock       sync.RWMutex
	ticker           map[string]poloniex.PoloniexTicker
	LastTickerUpdate time.Time

	quit chan bool

	BitfinLender *BitfinexLender
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
	l.recordMap = make(map[int]map[string]time.Time)
	l.recordMap[balancer.PoloniexExchange] = make(map[string]time.Time)
	l.recordMap[balancer.BitfinexExchange] = make(map[string]time.Time)
	l.ticker = make(map[string]poloniex.PoloniexTicker)
	l.currentLoanRate = make(map[int]map[string]balancer.LoanRates)
	l.BitfinLender = NewBitfinexLender()

	return l
}

type LendUser struct {
	U balancer.User
}

func (l *Lender) Runloop() {
	go l.BitfinLender.Run()
	for {
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
				err := l.ProcessPoloniexUser(u)
				if err != nil {
					poloLogger.WithFields(log.Fields{"func": "ProcessPoloniexUser", "user": u.U.Username,
						"exchange": balancer.GetExchangeString(u.U.Exchange)}).Errorf("[PoloLending] Error: %s", err.Error())
				}
			case balancer.BitfinexExchange:
				err := l.ProcessBitfinexUser(u)
				if err != nil {
					poloLogger.WithFields(log.Fields{"func": "ProcessBitfinexUser", "user": u.U.Username,
						"exchange": balancer.GetExchangeString(u.U.Exchange)}).Errorf("[BitfinexLending] Error: %s", err.Error())
				}
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

func (l *Lender) ProcessPoloniexUser(u *LendUser) error {
	dbu, err := l.Bee.FetchUser(u.U.Username)
	if err != nil {
		return err
	}

	if u.U.AccessKey == "" {
		return fmt.Errorf("No access key for user %s", u.U.Username)
	}

	if u.U.SecretKey == "" {
		return fmt.Errorf("No secret key for user %s", u.U.Username)
	}

	flog := poloLogger.WithFields(log.Fields{"func": "ProcessPoloniexUser()", "user": u.U.Username})

	part1 := time.Now()
	var _ = part1
	//JobPart2
	bals := make(map[string]map[string]float64)
	// Try 3 times if timeout
	for i := 0; i < 3; i++ {
		bals, err = l.Polo.PoloniexGetAvailableBalances(u.U.AccessKey, u.U.SecretKey)
		if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			flog.WithField("retry", i).Errorf("Error getting balances: %s", err.Error())
			if !strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				continue
			}
		}
		break
	}

	var inactiveLoans map[string][]poloniex.PoloniexLoanOffer
	// 3 types of balances: Not lent, Inactive, Active
	for i := 0; i < 3; i++ {
		inactiveLoans, err = l.Polo.PoloniexGetInactiveLoans(u.U.AccessKey, u.U.SecretKey)
		if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			flog.WithField("retry", i).Errorf("Error getting inactive loans: %s", err.Error())
			if !strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				continue
			}
		}
		break
	}

	var activeLoans *poloniex.PoloniexActiveLoans
	stats := userdb.NewAllUserStatistic()

	for i := 0; i < 3; i++ {
		activeLoans, err = l.Polo.PoloniexGetActiveLoans(u.U.AccessKey, u.U.SecretKey)
		if err == nil && activeLoans != nil {
			stats, err = l.recordStatistics(u.U.Username, bals, inactiveLoans, activeLoans)
			if err != nil {
				flog.WithField("retry", i).Errorf("Error in calculating statistic: %s", err.Error())
			}
		} else if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			flog.WithField("retry", i).Errorf("Error getting active loans: %s", err.Error())
			if !strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				continue
			}
		}
		break
	}
	var _ = stats

	// u.Active = activeLoans
	// u.Balances = bals
	// u.Inactive = inactiveLoans
	JobPart1.Observe(float64(time.Since(part1).Nanoseconds()))
	part2 := time.Now()

	for _, curr := range dbu.PoloniexEnabled.Keys() { //u.U.MinimumLend {
		min := dbu.PoloniexMiniumLend.Get(curr)
		clog := flog.WithFields(log.Fields{"currency": curr, "exchange": balancer.GetExchangeString(u.U.Exchange)})

		// Move min from a % to it's value
		min = min / 100

		// Make sure we have a rate to use
		l.loanrateLock.RLock()
		if l.currentLoanRate == nil || l.currentLoanRate[balancer.PoloniexExchange] == nil {
			l.loanrateLock.RUnlock()
			continue
		}
		if _, ok := l.currentLoanRate[balancer.PoloniexExchange][curr]; !ok {
			l.loanrateLock.RUnlock()
			continue
		}

		// Rate calculation
		rate := l.currentLoanRate[balancer.PoloniexExchange][curr].AvgBased
		l.loanrateLock.RUnlock()
		// Set to min if we are below
		if rate < min {
			rate = min
		}

		if curr == "BTC" {
			if rate < 2 {
				CompromisedBTC.Set(rate)
			} else {
				clog.Errorf("Rate is going to high. Trying to %s set at %f", curr, rate)
			}
		}

		avail, ok := bals["lending"][curr]
		var _ = ok

		maxLend := balancer.MaxLendAmt[balancer.PoloniexExchange][curr]
		// if ri == 2 {
		// 	maxLend = maxLend * 2
		// }

		if maxLend < avail*0.20 {
			maxLend = avail * 0.20
		}

		// We need to find some more crypto to lkend
		//if avail < MaxLendAmt[j.Currency[i]] {
		need := maxLend - avail
		if inactiveLoans != nil {
			currencyLoans := inactiveLoans[curr]
			sort.Sort(poloniex.PoloniexLoanOfferArray(currencyLoans))
			for _, loan := range currencyLoans {
				// We don't need any more funds, so we can exit this loop. But if the rate is less
				// than our minimum, we want to cancel that
				if need < 0 || loan.Rate < min {
					if loan.Rate < min {
						l.Polo.PoloniexCancelLoanOffer(curr, loan.ID, u.U.AccessKey, u.U.SecretKey)
					}
					continue
				}

				// So if the rate is less than the min, we don't want to cancel anything, unless the condition above
				if rate < min {
					rate = min
				}

				// Too close, no point in canceling
				if abs(loan.Rate-rate) < 0.0000015 {
					continue
				}
				worked, err := l.Polo.PoloniexCancelLoanOffer(curr, loan.ID, u.U.AccessKey, u.U.SecretKey)
				if err != nil {
					clog.Errorf("[Cancel] Error in Lending: %s", err.Error())
					continue
				}
				if worked && err == nil {
					need -= loan.Amount
					avail += loan.Amount
					LoansCanceled.Inc()
				}
			}
		}

		// Ensure we lend at least 10% of all at a time if we are under this value, but we will not cancel loans to get to
		// this value. This ensures we loan out a lot quick if we have a lot waiting, but don't if we have a lot lent
		if cStats, ok := stats.Currencies[curr]; ok {
			total := cStats.OnOrderBalance + cStats.ActiveLentBalance + cStats.AvailableBalance
			if total*0.1 > maxLend {
				maxLend = 0.1 * total
			}
		}

		amt := maxLend
		if avail < maxLend {
			amt = avail
		} else if avail < maxLend+balancer.MinLendAmt[u.U.Exchange][curr] {
			// If we make a loan, and don't have enough to make a following one, make this one to the available balance
			// This prevents the following:
			// 		I have 153FCT
			//		The minimum is 100
			//		We don't want to lend 100, then sit on 53 that can't be lent. It's better to lend all 153
			amt = avail
		}

		// To little for a loan
		if amt < balancer.MinLendAmt[u.U.Exchange][curr] {
			continue
		}

		// Yea.... no
		if rate == 0 || rate > 5 {
			continue
		}

		_, err = l.Polo.PoloniexCreateLoanOffer(curr, amt, rate, 2, false, u.U.AccessKey, u.U.SecretKey)
		if err != nil { //} && strings.Contains(err.Error(), "Too many requests") {
			// Sleep in our inner loop. This can be dangerous, should put these calls in a seperate queue to handle
			// time.Sleep(5 * time.Second)
			// _, err = s.PoloniexCreateLoanOffer(j.Currency[i], amt, rate, 2, false, j.Username)
			if err != nil {
				clog.Errorf("Error creating loan: %s", err.Error())
			}
		} else {
			clog.WithFields(log.Fields{"rate": rate, "amount": amt}).Infof("Created Loan")
		}

		if err != nil {
			clog.Errorf("[Offer] Error in Lending: %s", err.Error())
			continue
		}
		JobPart2.Observe(float64(time.Since(part2).Nanoseconds()))
		LoansCreated.Inc()
	}
	return nil
}

func (l *Lender) recordStatistics(username string, bals map[string]map[string]float64,
	inact map[string][]poloniex.PoloniexLoanOffer, activeLoan *poloniex.PoloniexActiveLoans) (*userdb.AllUserStatistic, error) {

	// Make stats
	stats := userdb.NewAllUserStatistic()
	stats.Time = time.Now()
	stats.Username = username

	for _, v := range balancer.Currencies[balancer.PoloniexExchange] {
		var last float64 = 1
		if v != "BTC" {
			l.tickerlock.RLock()
			last = l.ticker[fmt.Sprintf("BTC_%s", v)].Last
			l.tickerlock.RUnlock()
		}
		cstat := userdb.NewUserStatistic(v, last)
		stats.Currencies[v] = cstat
	}

	// Avail balance
	for k, v := range bals["lending"] {
		if !math.IsNaN(v) {
			stats.Currencies[k].AvailableBalance = v
			stats.TotalCurrencyMap[k] += l.getBTCAmount(v, k)
		}
	}

	// Active
	activeLentCount := make(map[string]float64)

	first := true
	for _, loan := range activeLoan.Provided {
		stats.Currencies[loan.Currency].ActiveLentBalance += loan.Amount
		stats.Currencies[loan.Currency].AverageActiveRate += loan.Rate
		activeLentCount[loan.Currency] += 1
		if first && loan.Rate != 0 {
			stats.Currencies[loan.Currency].HighestRate = loan.Rate
			stats.Currencies[loan.Currency].LowestRate = loan.Rate
			first = false
		} else {
			if loan.Rate > stats.Currencies[loan.Currency].HighestRate && loan.Rate != 0 {
				stats.Currencies[loan.Currency].HighestRate = loan.Rate
			}
			if loan.Rate < stats.Currencies[loan.Currency].LowestRate && loan.Rate != 0 {
				stats.Currencies[loan.Currency].LowestRate = loan.Rate
			}
		}
		stats.TotalCurrencyMap[loan.Currency] += l.getBTCAmount(loan.Amount, loan.Currency)
	}

	for k := range stats.Currencies {
		stats.Currencies[k].AverageActiveRate = stats.Currencies[k].AverageActiveRate / activeLentCount[k]
	}

	// On Order
	inactiveLentCount := make(map[string]float64)
	for k, _ := range inact {
		for _, loan := range inact[k] {
			stats.Currencies[k].OnOrderBalance += loan.Amount
			stats.Currencies[k].AverageOnOrderRate += loan.Rate
			inactiveLentCount[k] += 1

			stats.TotalCurrencyMap[k] += l.getBTCAmount(loan.Amount, k)
		}
	}

	for k := range stats.Currencies {
		stats.Currencies[k].AverageOnOrderRate = stats.Currencies[k].AverageOnOrderRate / inactiveLentCount[k]
	}

	// Check if to save
	l.recordMapLock.RLock()
	v, ok := l.recordMap[balancer.PoloniexExchange][username]
	l.recordMapLock.RUnlock()
	if ok {
		if time.Since(v) < time.Minute*10 {
			return stats, nil
		}
	}
	// Save here
	// TODO: Jesse Save the stats here. This is the userstatistics, we will retrieve these by time
	// db.RecordData(stats)
	err := l.Bee.SaveUserStastics(stats, balancer.PoloniexExchange)

	l.recordMapLock.Lock()
	l.recordMap[balancer.PoloniexExchange][username] = time.Now()
	l.recordMapLock.Unlock()
	return stats, err
}

func (l *Lender) getAmtForBTCValue(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	l.tickerlock.RLock()
	t, ok := l.ticker[fmt.Sprintf("BTC_%s", currency)]
	l.tickerlock.RUnlock()
	if !ok {
		return amount
	}

	return amount / t.Last
}

func (l *Lender) getBTCAmount(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
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
