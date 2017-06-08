package lender

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Print

var (
	MaxLendAmt map[string]float64
)

func init() {
	MaxLendAmt = make(map[string]float64)
	MaxLendAmt["BTC"] = .1
	MaxLendAmt["BTS"] = 1
	MaxLendAmt["CLAM"] = 1
	MaxLendAmt["DOGE"] = 1
	MaxLendAmt["DASH"] = 1
	MaxLendAmt["LTC"] = 1
	MaxLendAmt["MAID"] = 1
	MaxLendAmt["STR"] = 1
	MaxLendAmt["XMR"] = 1
	MaxLendAmt["XRP"] = 1
	MaxLendAmt["ETH"] = .2
	MaxLendAmt["FCT"] = 20
}

type LoanRates struct {
	Simple   float64
	AvgBased float64
}

type Lender struct {
	State    *core.State
	JobQueue chan *Job
	quit     chan struct{}

	Currency              string
	CurrentLoanRate       map[string]LoanRates
	LastCalculateLoanRate time.Time
	CalculateLoanInterval float64 // In seconds
	LastTickerUpdate      time.Time
	GetTickerInterval     float64
	Ticker                map[string]poloniex.PoloniexTicker
	PoloniexStats         map[string]*userdb.PoloniexStats
}

func NewLender(s *core.State) *Lender {
	l := new(Lender)
	l.State = s
	l.Currency = "BTC"
	l.JobQueue = make(chan *Job, 1000)
	l.CalculateLoanInterval = 1
	l.GetTickerInterval = 30
	l.Ticker = make(map[string]poloniex.PoloniexTicker)
	l.PoloniexStats = make(map[string]*userdb.PoloniexStats)
	l.CurrentLoanRate = make(map[string]LoanRates)
	l.CurrentLoanRate["BTC"] = LoanRates{Simple: 2.1}

	return l
}

func (l *Lender) Start() {
	l.UpdateTicker()
	for {
		select {
		case <-l.quit:
			l.quit <- struct{}{}
			return
		case j := <-l.JobQueue:
			// Update loan rate
			if time.Since(l.LastCalculateLoanRate).Seconds() >= l.CalculateLoanInterval {
				err := l.CalculateLoanRate("BTC")
				if err != nil {
					log.Println("[BTC] Error in Lending:", err)
				}

				err = l.CalculateLoanRate("FCT")
				if err != nil {
					log.Println("[FCT] Error in Lending:", err)
				}
			}

			// Update Ticker
			if time.Since(l.LastTickerUpdate).Seconds() >= l.GetTickerInterval {
				l.UpdateTicker()
			}

			err := l.ProcessJob(j)
			if err != nil {
				log.Println("Error in Lending:", err)
			}
		}
	}
}

func (l *Lender) Close() {
	l.quit <- struct{}{}
}

func (l *Lender) AddJob(j *Job) error {
	l.JobQueue <- j
	return nil
}

func (l *Lender) JobQueueLength() int {
	return len(l.JobQueue)
}

func (l *Lender) UpdateTicker() {
	ticker, err := l.State.PoloniexAPI.GetTicker()
	if err == nil {
		l.Ticker = ticker
	}
	l.LastTickerUpdate = time.Now()
	l.PoloniexStats["BTC"] = l.State.GetPoloniexStatistics("BTC")
	// Prometheus
	if l.PoloniexStats["BTC"] != nil {
		PoloniexStatsHourlyAvg.Set(l.PoloniexStats["BTC"].HrAvg)
		PoloniexStatsDailyAvg.Set(l.PoloniexStats["BTC"].DayAvg)
		PoloniexStatsWeeklyAvg.Set(l.PoloniexStats["BTC"].WeekAvg)
		PoloniexStatsMonthlyAvg.Set(l.PoloniexStats["BTC"].MonthAvg)
		PoloniexStatsHourlyStd.Set(l.PoloniexStats["BTC"].HrStd)
		PoloniexStatsDailyStd.Set(l.PoloniexStats["BTC"].DayStd)
		PoloniexStatsWeeklyStd.Set(l.PoloniexStats["BTC"].WeekStd)
		PoloniexStatsMonthlyStd.Set(l.PoloniexStats["BTC"].MonthStd)
	}

	if v, ok := ticker["BTC_FCT"]; ok {
		TickerFCTValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_BTS"]; ok {
		TickerBTSValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_CLAM"]; ok {
		TickerCLAMValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_DOGE"]; ok {
		TickerDOGEValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_LTC"]; ok {
		TickerLTCValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_MAID"]; ok {
		TickerMAIDValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_STR"]; ok {
		TickerSTRValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_XMR"]; ok {
		TickerXMRValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_XRP"]; ok {
		TickerXRPValue.Set(v.Last)
	}
	if v, ok := ticker["BTC_ETH"]; ok {
		TickerETHValue.Set(v.Last)
	}

	LenderUpdateTicker.Inc()
}

func (l *Lender) CalculateLoanRate(currency string) error {
	s := l.State
	loans, err := s.PoloniexGetLoanOrders(currency)
	if err != nil {
		return err
	}

	breakoff := l.getAmtForBTCValue(5, currency)

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

	lowest := float64(2.1)
	for _, o := range all[index].Orders {
		if o.Rate < lowest {
			lowest = o.Rate
		}
	}

	lr := l.CurrentLoanRate[currency]
	lr.Simple = lowest
	if l.CurrentLoanRate[currency].Simple < 2 {
		SetSimple(currency, lowest)
		s.RecordPoloniexStatistics(currency, lowest)
	}
	// lr.AvgBased = lr.Simple
	l.CurrentLoanRate[currency] = lr

	l.calculateAvgBasedLoanRate(currency)

	return nil
}

func (l *Lender) calculateAvgBasedLoanRate(currency string) {
	rates, ok := l.CurrentLoanRate[currency]
	if !ok {
		l.CurrentLoanRate[currency] = LoanRates{Simple: 2, AvgBased: 2}
	}
	rates.AvgBased = rates.Simple

	stats, ok := l.PoloniexStats[currency]
	if !ok || stats == nil {
		log.Printf("No poloniex stats for %s", currency)
		l.CurrentLoanRate[currency] = rates
		return
	}

	a := rates.Simple
	// If less than hour average, we need to decide on whether or not to go higher
	if a < stats.HrAvg {
		// Lends are raising, go up
		if l.rising(currency) == 1 {
			a = stats.HrAvg + (stats.DayStd * 0.5)
		} else {
			a = stats.HrAvg
		}
	}

	rates.AvgBased = a
	l.CurrentLoanRate[currency] = rates
	if a < 2 {
		SetAvg(currency, a)
	}
}

// rising indicates if the rates are rising
//		0 for not rising
//		1 for rising
//		2 for more rising
func (l *Lender) rising(currency string) int {
	if l.PoloniexStats[currency].HrAvg > l.PoloniexStats[currency].DayAvg {
		return 1
	}
	return 0
}

func abs(v float64) float64 {
	if v < 0 {
		return v * -1
	}
	return v
}

func (l *Lender) recordStatistics(username string, bals map[string]map[string]float64,
	inact map[string][]poloniex.PoloniexLoanOffer, activeLoan *poloniex.PoloniexActiveLoans) error {

	stats := userdb.NewUserStatistic()
	stats.Time = time.Now()
	stats.Username = username
	stats.Currency = "BTC"

	var avail float64 = 0
	// Avail balance
	for k, v := range bals["lending"] {
		if !math.IsNaN(v) {
			avail += l.getBTCAmount(v, k)
		}
	}

	stats.AvailableBalance = avail

	// Active
	activeLentBal := float64(0)
	activeLentTotalRate := float64(0)
	activeLentCount := float64(0)

	for _, loan := range activeLoan.Provided {
		//if l.Currency == "BTC" {
		activeLentBal += l.getBTCAmount(loan.Amount, loan.Currency)
		activeLentTotalRate += loan.Rate
		activeLentCount++
		//}
		stats.TotalCurrencyMap[loan.Currency] += l.getBTCAmount(loan.Amount, loan.Currency)
	}

	stats.ActiveLentBalance = activeLentBal
	stats.AverageActiveRate = activeLentTotalRate / activeLentCount

	// On Order

	inactiveLentBal := float64(0)
	inactiveLentTotalRate := float64(0)
	inactiveLentCount := float64(0)
	for k, _ := range inact {
		for _, loan := range inact[k] {
			//if l.Currency == "BTC" {
			inactiveLentBal += l.getBTCAmount(loan.Amount, k)
			inactiveLentTotalRate += loan.Rate
			inactiveLentCount++
			//}
			stats.TotalCurrencyMap[loan.Currency] += l.getBTCAmount(loan.Amount, k)
		}
	}

	stats.OnOrderBalance = inactiveLentBal
	stats.AverageOnOrderRate = inactiveLentTotalRate / inactiveLentCount

	// Set totals for other coins
	//		Set Available
	availMap, ok := bals["lending"]
	if ok {
		for k, v := range availMap {
			stats.TotalCurrencyMap[k] += l.getBTCAmount(v, k)
		}
	}

	return l.State.RecordStatistics(stats)
}

func (l *Lender) getAmtForBTCValue(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	t, ok := l.Ticker[fmt.Sprintf("BTC_%s", currency)]
	if !ok {
		return amount
	}

	return amount / t.Last
}

func (l *Lender) getBTCAmount(amount float64, currency string) float64 {
	if currency == "BTC" {
		return amount
	}

	t, ok := l.Ticker[fmt.Sprintf("BTC_%s", currency)]
	if !ok {
		return 0
	}

	return t.Last * amount
}

// ProcessJob will calculate the newest loan rate, then it create a loan for 0.1 btc at that rate
// for the user in the Job
func (l *Lender) ProcessJob(j *Job) error {
	if j.Username == "" {
		return nil
	}
	switch j.Strategy {
	default:
		r := l.CurrentLoanRate[j.Currency].AvgBased
		if l.CurrentLoanRate[j.Currency].Simple == l.CurrentLoanRate[j.Currency].AvgBased {
			if l.rising(j.Currency) == 1 {
				r += l.PoloniexStats[j.Currency].DayStd * .1
			}
		}
		return l.tierOneProcessJob(j, r, j.Currency)
	}
}

func (l *Lender) tierOneProcessJob(j *Job, rate float64, currency string) error {
	j.MinimumLend = 0.0008
	if rate < j.MinimumLend {
		return nil
	}

	s := l.State
	// total := float64(0)

	bals, err := s.PoloniexGetAvailableBalances(j.Username)
	if err != nil {
		return err
	}

	// 3 types of balances: Not lent, Inactive, Active
	inactiveLoans, _ := s.PoloniexGetInactiveLoans(j.Username)

	activeLoans, err := s.PoloniexGetActiveLoans(j.Username)
	if err == nil && activeLoans != nil {
		err := l.recordStatistics(j.Username, bals, inactiveLoans, activeLoans)
		if err != nil {
			log.Printf("Error in calculating statistic for %s: %s", j.Username, err.Error())
		}
	}

	avail, ok := bals["lending"][currency]
	var _ = ok

	// rate := l.decideRate(rate, avail, total)

	// We need to find some more crypto to lkend
	if avail < MaxLendAmt[currency] {
		need := MaxLendAmt[currency] - avail
		if inactiveLoans != nil {
			currencyLoans := inactiveLoans[currency]
			sort.Sort(poloniex.PoloniexLoanOfferArray(currencyLoans))
			for _, loan := range currencyLoans {
				if need < 0 {
					break
				}

				// Too close, no point in canceling
				if abs(loan.Rate-rate) < 0.00000009 {
					continue
				}
				worked, err := s.PoloniexCancelLoanOffer(currency, loan.ID, j.Username)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if worked && err == nil {
					need -= loan.Amount
					avail += loan.Amount
					LoansCanceled.Inc()
				}
			}
		}
	}

	amt := MaxLendAmt[currency]
	if avail < MaxLendAmt[currency] {
		amt = avail
	}

	// To little for a loan
	if amt < 0.01 {
		return nil
	}
	_, err = s.PoloniexCreateLoanOffer(currency, amt, rate, 2, false, j.Username)
	if err != nil {
		return err
	}
	LoansCreated.Inc()
	JobsDone.Inc()

	return nil
}
