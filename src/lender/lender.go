package lender

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Print

var (
	MaxLendAmt float64 = .1
)

type LoanRates struct {
	Simple   float64
	AvgBased float64
}

type Lender struct {
	State    *core.State
	JobQueue chan *Job
	quit     chan struct{}

	Currency              string
	CurrentLoanRate       LoanRates
	LastCalculateLoanRate time.Time
	CalculateLoanInterval float64 // In seconds
	LastTickerUpdate      time.Time
	GetTickerInterval     float64
	Ticker                map[string]poloniex.PoloniexTicker
	PoloniexStats         *userdb.PoloniexStats
}

func NewLender(s *core.State) *Lender {
	l := new(Lender)
	l.State = s
	l.CurrentLoanRate.Simple = 2.1
	l.Currency = "BTC"
	l.JobQueue = make(chan *Job, 1000)
	l.CalculateLoanInterval = 1
	l.GetTickerInterval = 30
	l.Ticker = make(map[string]poloniex.PoloniexTicker)

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
				err := l.CalculateLoanRate()
				if err != nil {
					log.Println("Error in Lending:", err)
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
	l.PoloniexStats = l.State.GetPoloniexStatistics()
	// Prometheus
	if l.PoloniexStats != nil {
		PoloniexStatsHourlyAvg.Set(l.PoloniexStats.HrAvg)
		PoloniexStatsDailyAvg.Set(l.PoloniexStats.DayAvg)
		PoloniexStatsWeeklyAvg.Set(l.PoloniexStats.WeekAvg)
		PoloniexStatsMonthlyAvg.Set(l.PoloniexStats.MonthAvg)
		PoloniexStatsHourlyStd.Set(l.PoloniexStats.HrStd)
		PoloniexStatsDailyStd.Set(l.PoloniexStats.DayStd)
		PoloniexStatsWeeklyStd.Set(l.PoloniexStats.WeekStd)
		PoloniexStatsMonthlyStd.Set(l.PoloniexStats.MonthStd)
	}
	LenderUpdateTicker.Inc()
}

func (l *Lender) CalculateLoanRate() error {
	s := l.State
	loans, err := s.PoloniexGetLoanOrders(l.Currency)
	if err != nil {
		return err
	}

	index := 200
	amt := 0.000

	all := GetDensityOfLoans(loans)
	for i, orderRange := range all {
		amt += orderRange.Amount
		if amt > 5 {
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

	l.CurrentLoanRate.Simple = lowest
	if l.CurrentLoanRate.Simple < 2 {
		CurrentLoanRate.Set(l.CurrentLoanRate.Simple) // Prometheus
		s.RecordPoloniexStatistics(l.CurrentLoanRate.Simple)
	}

	l.calculateAvgBasedLoanRate()

	return nil
}

func (l *Lender) calculateAvgBasedLoanRate() {
	simple := l.CurrentLoanRate.Simple
	a := l.CurrentLoanRate.Simple
	// If less than hour average, we need to decide on whether or not to go higher
	if simple < l.PoloniexStats.HrAvg {
		// Lends are raising, go up
		if l.rising() == 1 {
			a = l.PoloniexStats.HrAvg + (l.PoloniexStats.DayStd * 0.5)
		} else {
			a = l.PoloniexStats.HrAvg
		}
	}

	l.CurrentLoanRate.AvgBased = a
	if a < 2 {
		LenderCurrentAverageBasedRate.Set(a)
	}
}

// rising indicates if the rates are rising
//		0 for not rising
//		1 for rising
//		2 for more rising
func (l *Lender) rising() int {
	if l.PoloniexStats.HrAvg > l.PoloniexStats.DayAvg {
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

	// Avail balance
	avail, ok := bals["lending"]["BTC"]
	var _ = ok
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
			inactiveLentBal += l.getBTCAmount(loan.Amount, loan.Currency)
			inactiveLentTotalRate += loan.Rate
			inactiveLentCount++
			//}
			stats.TotalCurrencyMap[loan.Currency] += l.getBTCAmount(loan.Amount, loan.Currency)
		}
	}

	stats.OnOrderBalance = inactiveLentBal
	stats.AverageOnOrderRate = inactiveLentTotalRate / inactiveLentCount

	// Set totals for other coins
	//		Set Available
	availMap, ok := bals["lending"]
	if ok {
		for k, v := range availMap {
			stats.TotalCurrencyMap[k] += v
		}
	}

	return l.State.RecordStatistics(stats)
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
		r := l.CurrentLoanRate.AvgBased
		if l.CurrentLoanRate.Simple == l.CurrentLoanRate.AvgBased {
			if l.rising() == 1 {
				r += l.PoloniexStats.DayStd * .1
			}
		}
		return l.tierOneProcessJob(j, r)
	}
}

func (l *Lender) decideRate(rate float64, avail float64, total float64) {
	if rate < l.PoloniexStats.DayAvg {

	}
}

func (l *Lender) tierOneProcessJob(j *Job, rate float64) error {
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

	avail, ok := bals["lending"][l.Currency]
	var _ = ok

	// rate := l.decideRate(rate, avail, total)

	// We need to find some more crypto to lkend
	if avail < MaxLendAmt {
		need := MaxLendAmt - avail
		if inactiveLoans != nil {
			currencyLoans := inactiveLoans[l.Currency]
			sort.Sort(poloniex.PoloniexLoanOfferArray(currencyLoans))
			fmt.Println(need, avail)
			for _, loan := range currencyLoans {
				if loan.Currency != "BTC" {
					fmt.Println(loan.Currency)
					fmt.Println("exit 1")
					continue
				}
				if need < 0 {
					fmt.Println("EXIT 2")
					break
				}
				fmt.Println(loan.Rate, rate, need, avail, abs(loan.Rate-rate))

				// Too close, no point in canceling
				if abs(loan.Rate-rate) < 0.00000009 {
					continue
				}
				worked, err := s.PoloniexCancelLoanOffer(l.Currency, loan.ID, j.Username)
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

	amt := MaxLendAmt
	if avail < MaxLendAmt {
		amt = avail
	}

	// To little for a loan
	if amt < 0.01 {
		return nil
	}
	_, err = s.PoloniexCreateLoanOffer(l.Currency, amt, rate, 2, false, j.Username)
	if err != nil {
		return err
	}
	LoansCreated.Inc()
	JobsDone.Inc()

	return nil
}
