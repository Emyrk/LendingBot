package balancer

import (
	"fmt"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var clog = log.WithFields(log.Fields{
	"package": "Lender",
})

var _ = log.Panic

type LoanRates struct {
	Simple   float64
	AvgBased float64

	// Used in parcel
	Currency string
	Exchange int
}

type QueenBee struct {
	PoloniexAPI *PoloniexAPIWithRateLimit

	loanrateLock          sync.RWMutex
	currentLoanRate       map[int]map[string]LoanRates
	lastCalculateLoanRate map[int]map[string]time.Time

	CalculateLoanInterval float64 // In seconds
	LastTickerUpdate      time.Time
	GetTickerInterval     time.Duration

	exchangeStatsLock sync.RWMutex
	exchangeStats     map[int]map[string]*userdb.PoloniexStats

	tickerLock sync.RWMutex
	ticker     map[string]poloniex.PoloniexTicker

	MasterHive *Hive

	quit chan struct{}
}

func NewRateCalculator(h *Hive) *QueenBee {
	q := new(QueenBee)
	q.currentLoanRate = make(map[int]map[string]LoanRates)
	q.lastCalculateLoanRate = make(map[int]map[string]time.Time)
	q.PoloniexAPI = NewPoloniexAPIWithRateLimit()
	q.MasterHive = h
	q.GetTickerInterval = time.Minute
	q.exchangeStats = make(map[int]map[string]*userdb.PoloniexStats)
	q.ticker = make(map[string]poloniex.PoloniexTicker)

	return q
}

func (q *QueenBee) Run() {
	go q.runPolo()
}

func (q *QueenBee) runPolo() {

	for {
		// Get rates of all currencies
		for _, c := range Currencies[PoloniexExchange] {
			err := q.CalculateLoanRate(PoloniexExchange, c)
			if err != nil {
				clog.WithFields(log.Fields{"method": "CalcLoop", "currency": c}).Errorf("Error in Lending: %s", err)
			}
			time.Sleep(250 * time.Millisecond)
		}

		// Update Ticker
		if time.Since(q.LastTickerUpdate) >= q.GetTickerInterval {
			go q.UpdateTicker()
			q.UpdateExchangeStats(PoloniexExchange)
		}

		// Sendout
		q.loanrateLock.RLock()
		q.tickerLock.RLock()
		p := NewLendingRatesP("ALL", q.currentLoanRate, q.ticker)
		q.loanrateLock.RUnlock()
		q.tickerLock.RUnlock()
		q.MasterHive.Slaves.SendParcelTo("ALL", p)

		select {
		case <-q.quit:
			q.quit <- struct{}{}
			return
		default:
		}
	}
}

func (q *QueenBee) CalculateLoanRate(exchange int, currency string) error {
	loans, err := q.PoloniexAPI.GetLoanOrders(currency)
	if err != nil {
		clog.WithFields(log.Fields{"method": "CalcLoan"}).Errorf("Error when grabbing loans for CalcRate: %s", err.Error())
		return err
	}

	if len(loans.Offers) == 0 {
		clog.WithFields(log.Fields{"method": "CalcLoan"}).Errorf("No offers found in loan book.")
	}

	breakoff := q.getAmtForBTCValue(5, currency)

	index := 200
	amt := 0.000

	all := GetDensityOfLoans(loans)
	for i, orderRange := range all {
		amt += orderRange.Amount
		if amt > breakoff {
			index = i
			break
		}
	}

	q.loanrateLock.RLock()
	lr := q.currentLoanRate[exchange][currency]
	q.loanrateLock.RUnlock()

	lowest := float64(1)
	if lr.AvgBased > 0 {
		lowest = lr.AvgBased * 2
	}

	for _, o := range all[index].Orders {
		if o.Rate < lowest {
			lowest = o.Rate
		}
	}

	lr.Simple = lowest
	q.loanrateLock.Lock()
	if q.currentLoanRate[exchange] == nil {
		q.currentLoanRate[exchange] = make(map[string]LoanRates)
	}
	q.currentLoanRate[exchange][currency] = lr
	if q.currentLoanRate[exchange][currency].Simple < 2 {
		SetSimple(currency, lowest)
		if time.Since(q.lastCalculateLoanRate[exchange][currency]).Seconds() > 30 {
			q.RecordExchangeStatistics(exchange, currency, lowest)
			if q.lastCalculateLoanRate[exchange] == nil {
				q.lastCalculateLoanRate[exchange] = make(map[string]time.Time)
			}
			q.lastCalculateLoanRate[exchange][currency] = time.Now()
		}
	}
	q.loanrateLock.Unlock()
	// lr.AvgBased = lr.Simple

	q.calculateAvgBasedLoanRate(exchange, currency)

	return nil
}

func (l *QueenBee) calculateAvgBasedLoanRate(exchange int, currency string) {
	l.loanrateLock.Lock()
	if l.currentLoanRate[exchange] == nil {
		l.currentLoanRate[exchange] = make(map[string]LoanRates)
	}
	rates, ok := l.currentLoanRate[exchange][currency]
	if !ok {
		l.currentLoanRate[exchange][currency] = LoanRates{Simple: 2, AvgBased: 2}
	}
	l.loanrateLock.Unlock()

	rates.AvgBased = rates.Simple

	l.exchangeStatsLock.Lock()
	stats, ok := l.exchangeStats[exchange][currency]
	l.exchangeStatsLock.Unlock()
	if !ok || stats == nil {
		clog.WithFields(log.Fields{"method": "CalcAvg"}).Errorf("[CalcAvg] No poloniex stats for %s", currency)
		l.loanrateLock.Lock()
		l.currentLoanRate[exchange][currency] = rates
		l.loanrateLock.Unlock()
		return
	}

	a := rates.Simple
	// If less than hour average, we need to decide on whether or not to go higher
	if a < stats.HrAvg {
		// Lends are raising, go up
		if l.rising(exchange, currency) >= 1 {
			a = stats.HrAvg + (stats.DayStd * 0.50)
		} else {
			a = stats.HrAvg
		}
	}

	if a < stats.FiveMinAvg && stats.FiveMinAvg > stats.HrAvg {
		a = stats.FiveMinAvg
	}

	rates.AvgBased = a
	l.loanrateLock.Lock()
	l.currentLoanRate[exchange][currency] = rates
	l.loanrateLock.Unlock()
	if a < 2 {
		SetAvg(currency, a)
	}
}

// RecordExchangeStatistics saves the rate for the exchangeinto mongodb.
//		Save the timestamp, currency, and rate
func (q *QueenBee) RecordExchangeStatistics(exchange int, currency string, lowest float64) {
	// TODO: Jesse, save these into Exchange DB
	//	Must be able to l.GetExchangeStatisitics(exch, currency)

}

func (q *QueenBee) getAmtForBTCValue(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	q.tickerLock.RLock()
	t, ok := q.ticker[fmt.Sprintf("BTC_%s", currency)]
	q.tickerLock.RUnlock()
	if !ok {
		return amount
	}

	return amount / t.Last
}

func (q *QueenBee) getBTCAmount(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}
	q.tickerLock.RLock()
	t, ok := q.ticker[fmt.Sprintf("BTC_%s", currency)]
	q.tickerLock.RUnlock()

	if !ok {
		return 0
	}

	return t.Last * amount
}

// rising indicates if the rates are rising
//		0 for not rising
//		1 for rising
//		2 for more rising
func (l *QueenBee) rising(exchange int, currency string) int {
	l.exchangeStatsLock.RLock()
	defer l.exchangeStatsLock.RUnlock()
	if v, ok := l.exchangeStats[exchange][currency]; !ok || v == nil {
		return 0
	}
	if l.exchangeStats[exchange][currency].HrAvg > l.exchangeStats[exchange][currency].DayAvg+(1*l.exchangeStats[exchange][currency].DayStd) {
		return 2
	} else if l.exchangeStats[exchange][currency].HrAvg > l.exchangeStats[exchange][currency].DayAvg+(.05*l.exchangeStats[exchange][currency].DayStd) {
		return 1
	}

	return 0
}

func (l *QueenBee) UpdateExchangeStats(exchange int) {
	l.exchangeStatsLock.Lock()
	if l.exchangeStats[exchange] == nil {
		l.exchangeStats[exchange] = make(map[string]*userdb.PoloniexStats)
	}
	l.exchangeStats[exchange]["BTC"] = l.GetExchangeStatisitics(exchange, "BTC")
	// Prometheus
	if l.exchangeStats[exchange]["BTC"] != nil {
		PoloniexStatsFiveMinAvg.Set(l.exchangeStats[exchange]["BTC"].FiveMinAvg)
		PoloniexStatsHourlyAvg.Set(l.exchangeStats[exchange]["BTC"].HrAvg)
		PoloniexStatsDailyAvg.Set(l.exchangeStats[exchange]["BTC"].DayAvg)
		PoloniexStatsWeeklyAvg.Set(l.exchangeStats[exchange]["BTC"].WeekAvg)
		PoloniexStatsMonthlyAvg.Set(l.exchangeStats[exchange]["BTC"].MonthAvg)
		PoloniexStatsHourlyStd.Set(l.exchangeStats[exchange]["BTC"].HrStd)
		PoloniexStatsDailyStd.Set(l.exchangeStats[exchange]["BTC"].DayStd)
		PoloniexStatsWeeklyStd.Set(l.exchangeStats[exchange]["BTC"].WeekStd)
		PoloniexStatsMonthlyStd.Set(l.exchangeStats[exchange]["BTC"].MonthStd)
	}

	l.exchangeStats[exchange]["FCT"] = l.GetExchangeStatisitics(exchange, "FCT")
	l.exchangeStats[exchange]["BTS"] = l.GetExchangeStatisitics(exchange, "BTS")
	l.exchangeStats[exchange]["CLAM"] = l.GetExchangeStatisitics(exchange, "CLAM")
	l.exchangeStats[exchange]["DOGE"] = l.GetExchangeStatisitics(exchange, "DOGE")
	l.exchangeStats[exchange]["DASH"] = l.GetExchangeStatisitics(exchange, "DASH")
	l.exchangeStats[exchange]["LTC"] = l.GetExchangeStatisitics(exchange, "LTC")
	l.exchangeStats[exchange]["MAID"] = l.GetExchangeStatisitics(exchange, "MAID")
	l.exchangeStats[exchange]["STR"] = l.GetExchangeStatisitics(exchange, "STR")
	l.exchangeStats[exchange]["XMR"] = l.GetExchangeStatisitics(exchange, "XMR")
	l.exchangeStats[exchange]["XRP"] = l.GetExchangeStatisitics(exchange, "XRP")
	l.exchangeStats[exchange]["ETH"] = l.GetExchangeStatisitics(exchange, "ETH")
	l.exchangeStatsLock.Unlock()

	l.tickerLock.RLock()
	if v, ok := l.ticker["BTC_FCT"]; ok {
		TickerFCTValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_BTS"]; ok {
		TickerBTSValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_CLAM"]; ok {
		TickerCLAMValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_DOGE"]; ok {
		TickerDOGEValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_LTC"]; ok {
		TickerLTCValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_MAID"]; ok {
		TickerMAIDValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_STR"]; ok {
		TickerSTRValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_XMR"]; ok {
		TickerXMRValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_XRP"]; ok {
		TickerXRPValue.Set(v.Last)
	}
	if v, ok := l.ticker["BTC_ETH"]; ok {
		TickerETHValue.Set(v.Last)
	}
	l.tickerLock.RUnlock()

}

func (l *QueenBee) UpdateTicker() {
	l.LastTickerUpdate = time.Now()
	ticker, err := l.PoloniexAPI.GetTicker()
	if err == nil {
		l.tickerLock.Lock()
		l.ticker = ticker
		l.tickerLock.Unlock()
	}
	LenderUpdateTicker.Inc()
}

func (q *QueenBee) GetExchangeStatisitics(exchange int, currency string) *userdb.PoloniexStats {
	// u, err := s.userStatistic.GetPoloniexStatistics(currency)
	// if err != nil {
	// 	fmt.Printf("ERROR: GetPoloniexstatistics: %s\n", err.Error())
	// 	return nil
	// }
	// return u
	return nil
}
