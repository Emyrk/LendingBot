package balancer

import (
	"fmt"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/slack"
	"github.com/Emyrk/LendingBot/src/core/bitfinex"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var clog = log.WithFields(log.Fields{
	"package": "RateCalculator",
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

	poloTickerLock sync.RWMutex
	poloTicker     map[string]poloniex.PoloniexTicker
	// EOS and IOT
	bitfinexTickerCorrections *ExtraTickers

	cachedTicker *map[string]poloniex.PoloniexTicker
	lastCache    time.Time

	// bitpoloTickerLock sync.RWMutex
	// bitpoloTicker     map[string]poloniex.PoloniexTicker

	MasterHive *Hive

	quit chan struct{}

	usdb *userdb.UserStatisticsDB
}

type ExtraTickers struct {
	iotLast  poloniex.PoloniexTicker
	eosLast  poloniex.PoloniexTicker
	lastDone time.Time
}

func (e *ExtraTickers) Update() {
	if time.Since(e.lastDone) > time.Minute*30 {
		api := bitfinex.New("", "")
		ti, err := api.Ticker("IOTBTC")
		if err == nil {
			e.iotLast.Last = ti.LastPrice
		}

		ti, err = api.Ticker("EOSBTC")
		if err == nil {
			e.eosLast.Last = ti.LastPrice
		}
		e.lastDone = time.Now()
	}
}

func NewRateCalculator(h *Hive, uri, dbu, dbp string) *QueenBee {
	var err error

	q := new(QueenBee)
	q.currentLoanRate = make(map[int]map[string]LoanRates)
	q.lastCalculateLoanRate = make(map[int]map[string]time.Time)
	q.PoloniexAPI = NewPoloniexAPIWithRateLimit()
	q.MasterHive = h
	q.GetTickerInterval = time.Minute
	q.exchangeStats = make(map[int]map[string]*userdb.PoloniexStats)
	q.poloTicker = make(map[string]poloniex.PoloniexTicker)
	q.usdb, err = userdb.NewUserStatisticsMongoDB(uri, dbu, dbp)
	if err != nil {
		if Test {
			slack.SendMessage(":rage:", "hive", "alerts", fmt.Sprintf("@channel ratecalculator %s: Oy!.. failed to connect to the userstat mongodb, I am panicing! Error: %s", err.Error()))
			panic(fmt.Sprintf("Failed to connect to userstat db: %s", err.Error()))
		} else {
			slack.SendMessage(":rage:", "hive", "alerts", fmt.Sprintf("@channel ratecalculator %s: Oy!.. failed to connect to the userstat mongodb, I am panicing! Error: %s", err.Error()))
			panic(fmt.Sprintf("Failed to connect to userstat db: %s", err.Error()))
		}
	}
	tmp := make(map[string]poloniex.PoloniexTicker)
	q.cachedTicker = &tmp

	return q
}

func (q *QueenBee) Run() {
	go q.runPolo()
}

func (q *QueenBee) runPolo() {
	q.UpdateExchangeStats(PoloniexExchange)
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
		q.poloTickerLock.RLock()
		p := NewLendingRatesP("ALL", q.currentLoanRate, q.poloTicker)
		q.loanrateLock.RUnlock()
		q.poloTickerLock.RUnlock()
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

	lr.AvgBased = lr.Simple
	q.currentLoanRate[exchange][currency] = lr
	if q.currentLoanRate[exchange][currency].Simple < 2 {
		SetSimple(currency, lr.Simple)
		if time.Since(q.lastCalculateLoanRate[exchange][currency]).Seconds() > 30 {
			q.RecordExchangeStatistics(exchange, currency, lr.Simple)
			if q.lastCalculateLoanRate[exchange] == nil {
				q.lastCalculateLoanRate[exchange] = make(map[string]time.Time)
			}
			q.lastCalculateLoanRate[exchange][currency] = time.Now()
		}
	}
	q.loanrateLock.Unlock()

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
func (q *QueenBee) RecordExchangeStatistics(exchange int, currency string, lowest float64) error {
	switch exchange {
	case PoloniexExchange:
		return q.usdb.RecordPoloniexStatistic(currency, lowest)
	case BitfinexExchange:
		return nil
	}
	return nil
}

func (q *QueenBee) getAmtForBTCValue(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	q.poloTickerLock.RLock()
	t, ok := q.poloTicker[fmt.Sprintf("BTC_%s", currency)]
	q.poloTickerLock.RUnlock()
	if !ok {
		return amount
	}

	return amount / t.Last
}

func (q *QueenBee) getBTCAmount(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}
	q.poloTickerLock.RLock()
	t, ok := q.poloTicker[fmt.Sprintf("BTC_%s", currency)]
	q.poloTickerLock.RUnlock()

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

	l.poloTickerLock.RLock()
	if v, ok := l.poloTicker["BTC_FCT"]; ok {
		TickerFCTValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_BTS"]; ok {
		TickerBTSValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_CLAM"]; ok {
		TickerCLAMValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_DOGE"]; ok {
		TickerDOGEValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_LTC"]; ok {
		TickerLTCValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_MAID"]; ok {
		TickerMAIDValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_STR"]; ok {
		TickerSTRValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_XMR"]; ok {
		TickerXMRValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_XRP"]; ok {
		TickerXRPValue.Set(v.Last)
	}
	if v, ok := l.poloTicker["BTC_ETH"]; ok {
		TickerETHValue.Set(v.Last)
	}
	l.poloTickerLock.RUnlock()

}

func (l *QueenBee) GetTicker() *map[string]poloniex.PoloniexTicker {
	if time.Since(l.lastCache) < time.Minute*10 {
		return l.cachedTicker
	}
	newTicker := make(map[string]poloniex.PoloniexTicker)
	l.poloTickerLock.RLock()
	for k, v := range l.poloTicker {
		newTicker[k] = v
	}
	l.poloTickerLock.RUnlock()

	l.cachedTicker = &newTicker
	l.lastCache = time.Now()
	return l.cachedTicker
}

func (l *QueenBee) ONLY_USE_FOR_TESTING_GET_TICKER() map[string]poloniex.PoloniexTicker {
	return l.poloTicker
}

func (l *QueenBee) UpdateTicker() {
	l.LastTickerUpdate = time.Now()
	poloTicker, err := l.PoloniexAPI.GetTicker()
	if err == nil {
		l.bitfinexTickerCorrections.Update()
		l.poloTickerLock.Lock()
		l.poloTicker["BTC_EOS"] = l.bitfinexTickerCorrections.eosLast
		l.poloTicker["BTC_IOT"] = l.bitfinexTickerCorrections.eosLast
		l.poloTicker = poloTicker
		l.poloTickerLock.Unlock()
	}
	LenderUpdateTicker.Inc()
}

func (q *QueenBee) GetExchangeStatisitics(exchange int, currency string) *userdb.PoloniexStats {

	var llog = log.WithFields(log.Fields{
		"method": "GetExchangeStatisitics",
	})
	switch exchange {
	case PoloniexExchange:
		u, err := q.usdb.GetPoloniexStatistics(currency)
		if err != nil {
			llog.Error(err.Error())
		}
		return u
	case BitfinexExchange:
		return nil
	}
	return nil
}
