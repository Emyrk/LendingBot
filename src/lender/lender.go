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
	MinLendAmt map[string]float64
)

func init() {
	MaxLendAmt = make(map[string]float64)
	MaxLendAmt["BTC"] = .1
	MaxLendAmt["BTS"] = 1
	MaxLendAmt["CLAM"] = 20
	MaxLendAmt["DOGE"] = 1
	MaxLendAmt["DASH"] = 1
	MaxLendAmt["LTC"] = 1
	MaxLendAmt["MAID"] = 1
	MaxLendAmt["STR"] = 1
	MaxLendAmt["XMR"] = 1
	MaxLendAmt["XRP"] = 1
	MaxLendAmt["ETH"] = .2
	MaxLendAmt["FCT"] = 200

	MinLendAmt = make(map[string]float64)
	MinLendAmt["BTC"] = .01
	MinLendAmt["BTS"] = 1
	MinLendAmt["CLAM"] = 10
	MinLendAmt["DOGE"] = 1
	MinLendAmt["DASH"] = 1
	MinLendAmt["LTC"] = 1
	MinLendAmt["MAID"] = 1
	MinLendAmt["STR"] = 1
	MinLendAmt["XMR"] = 1
	MinLendAmt["XRP"] = 1
	MinLendAmt["ETH"] = .2
	MinLendAmt["FCT"] = 100
}

var curarr = []string{"BTC", "BTS", "CLAM", "DOGE", "DASH", "LTC", "MAID", "STR", "XMR", "XRP", "ETH", "FCT"}

type LoanRates struct {
	Simple   float64
	AvgBased float64
}

type Lender struct {
	State    *core.State
	JobQueue chan *Job
	quit     chan struct{}

	CurrentLoanRate       map[string]LoanRates
	LastCalculateLoanRate map[string]time.Time
	CalculateLoanInterval float64 // In seconds
	LastTickerUpdate      time.Time
	GetTickerInterval     float64
	Ticker                map[string]poloniex.PoloniexTicker
	PoloniexStats         map[string]*userdb.PoloniexStats
}

func NewLender(s *core.State) *Lender {
	l := new(Lender)
	l.State = s
	l.JobQueue = make(chan *Job, 100)
	l.CalculateLoanInterval = 1
	l.GetTickerInterval = 30
	l.Ticker = make(map[string]poloniex.PoloniexTicker)
	l.PoloniexStats = make(map[string]*userdb.PoloniexStats)
	l.CurrentLoanRate = make(map[string]LoanRates)
	l.CurrentLoanRate["BTC"] = LoanRates{Simple: 2.1}
	l.LastCalculateLoanRate = make(map[string]time.Time)

	// for i, c := range curarr {
	// 	l.LastCalculateLoanRate[c] = time.Now().Add(time.Second * time.Duration(i))
	// }

	return l
}

func (l *Lender) CalcLoop() {
	// ticker := time.NewTicker(time.Second)
	for {
		i := 0
		max := len(curarr)
		for {
			if i >= max {
				i = 0
			}
			err := l.CalculateLoanRate(curarr[i])
			if err != nil {
				log.Printf("[%s] Error in Lending: %s", curarr[i], err)
			}
			// l.LastCalculateLoanRate[curarr[i]] = time.Now()
			time.Sleep(300 * time.Millisecond)
			i++
		}
	}
}

func (l *Lender) Start() {
	l.UpdateTicker()
	go l.CalcLoop()
	for {
		select {
		case <-l.quit:
			l.quit <- struct{}{}
			return
		case j := <-l.JobQueue:
			start := time.Now()
			JobQueueCurrent.Set(float64(len(l.JobQueue)))
			if j.Currency == nil {
				fmt.Println("Seems we got a nil currency string:", j.Username)
				break
			}

			// Update Ticker
			if time.Since(l.LastTickerUpdate).Seconds() >= l.GetTickerInterval {
				go l.UpdateTicker()
			}

			err := l.ProcessJob(j)
			if err != nil {
				log.Println("Error in Lending for", j.Username, ":", err)
			}
			JobProcessDuration.Observe(float64(time.Since(start).Nanoseconds()))
			JobsDone.Inc()
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

	l.PoloniexStats["FCT"] = l.State.GetPoloniexStatistics("FCT")
	l.PoloniexStats["BTS"] = l.State.GetPoloniexStatistics("BTS")
	l.PoloniexStats["CLAM"] = l.State.GetPoloniexStatistics("CLAM")
	l.PoloniexStats["DOGE"] = l.State.GetPoloniexStatistics("DOGE")
	l.PoloniexStats["DASH"] = l.State.GetPoloniexStatistics("DASH")
	l.PoloniexStats["LTC"] = l.State.GetPoloniexStatistics("LTC")
	l.PoloniexStats["MAID"] = l.State.GetPoloniexStatistics("MAID")
	l.PoloniexStats["STR"] = l.State.GetPoloniexStatistics("STR")
	l.PoloniexStats["XMR"] = l.State.GetPoloniexStatistics("XMR")
	l.PoloniexStats["XRP"] = l.State.GetPoloniexStatistics("XRP")
	l.PoloniexStats["ETH"] = l.State.GetPoloniexStatistics("ETH")

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
		log.Printf("Error when grabbing loans for CalcRate: %s", err.Error())
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
	l.CurrentLoanRate[currency] = lr
	if l.CurrentLoanRate[currency].Simple < 2 {
		SetSimple(currency, lowest)
		if time.Since(l.LastCalculateLoanRate[currency]).Seconds() > 5 {
			s.RecordPoloniexStatistics(currency, lowest)
			l.LastCalculateLoanRate[currency] = time.Now()
		}
	}
	// lr.AvgBased = lr.Simple

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
	if v, ok := l.PoloniexStats[currency]; !ok || v == nil {
		return 0
	}
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
			btcAmt := l.getBTCAmount(loan.Amount, k)
			inactiveLentBal += btcAmt
			inactiveLentTotalRate += loan.Rate
			inactiveLentCount++
			//}
			stats.TotalCurrencyMap[k] += btcAmt
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
		return l.tierOneProcessJob(j)
	}
}

func (l *Lender) tierOneProcessJob(j *Job) error {
	s := l.State

	bals, err := s.PoloniexGetAvailableBalances(j.Username)
	// if err != nil {
	// 	return fmt.Errorf("[T1-1] %s", err.Error())
	// }

	// 3 types of balances: Not lent, Inactive, Active
	inactiveLoans, _ := s.PoloniexGetInactiveLoans(j.Username)

	activeLoans, err := s.PoloniexGetActiveLoans(j.Username)
	if err == nil && activeLoans != nil {
		err := l.recordStatistics(j.Username, bals, inactiveLoans, activeLoans)
		if err != nil {
			log.Printf("Error in calculating statistic for %s: %s", j.Username, err.Error())
		}
	}

	for i, min := range j.MinimumLend {
		rate := l.CurrentLoanRate[j.Currency[i]].AvgBased
		if l.CurrentLoanRate[j.Currency[i]].Simple == l.CurrentLoanRate[j.Currency[i]].AvgBased {
			if l.rising(j.Currency[i]) == 1 {
				rate += l.PoloniexStats[j.Currency[i]].DayStd * .1
			}
		}

		avail, ok := bals["lending"][j.Currency[i]]
		var _ = ok

		// rate := l.decideRate(rate, avail, total)

		// We need to find some more crypto to lkend
		//if avail < MaxLendAmt[j.Currency[i]] {
		need := MaxLendAmt[j.Currency[i]] - avail
		if inactiveLoans != nil {
			currencyLoans := inactiveLoans[j.Currency[i]]
			sort.Sort(poloniex.PoloniexLoanOfferArray(currencyLoans))
			for _, loan := range currencyLoans {
				// We don't need any more funds, so we can exit this loop. But if the rate is less
				// than our minimum, we want to cancel that
				if need < 0 || loan.Rate < min {
					if loan.Rate < min {
						s.PoloniexCancelLoanOffer(j.Currency[i], loan.ID, j.Username)
					}
					continue
				}

				// So if the rate is less than the min, we don't want to cancel anything, unless the condition above
				if rate < min {
					continue
				}

				// Too close, no point in canceling
				if abs(loan.Rate-rate) < 0.00000009 {
					continue
				}
				worked, err := s.PoloniexCancelLoanOffer(j.Currency[i], loan.ID, j.Username)
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
		//}

		// Don't make the loan
		if rate < min {
			continue
		}

		amt := MaxLendAmt[j.Currency[i]]
		if avail < MaxLendAmt[j.Currency[i]] {
			amt = avail
		} else if avail < MaxLendAmt[j.Currency[i]]+MinLendAmt[j.Currency[i]] {
			// If we make a loan, and don't have enough to make a following one, make this one to the available balance
			amt = avail
		}

		// To little for a loan
		if amt < MinLendAmt[j.Currency[i]] {
			continue
		}

		// Yea.... no
		if rate == 0 {
			continue
		}

		_, err = s.PoloniexCreateLoanOffer(j.Currency[i], amt, rate, 2, false, j.Username)
		if err != nil {
			fmt.Println("Error in lenidng:", err.Error())
			continue
		}
		LoansCreated.Inc()
	}

	return nil
}
