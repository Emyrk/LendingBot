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

type Lender struct {
	State    *core.State
	JobQueue chan *Job
	quit     chan struct{}

	Currency              string
	CurrentLoanRate       float64
	LastCalculateLoanRate time.Time
	CalculateInterval     float64 // In seconds
}

func NewLender(s *core.State) *Lender {
	l := new(Lender)
	l.State = s
	l.CurrentLoanRate = 2.1
	l.Currency = "BTC"
	l.JobQueue = make(chan *Job, 1000)
	l.CalculateInterval = 1

	return l
}

func (l *Lender) Start() {
	for {
		select {
		case <-l.quit:
			l.quit <- struct{}{}
			return
		case j := <-l.JobQueue:
			if time.Since(l.LastCalculateLoanRate).Seconds() >= l.CalculateInterval {
				err := l.CalculateLoanRate()
				if err != nil {
					log.Println("Error in Lending:", err)
				}
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

func (l *Lender) CalculateLoanRate() error {
	s := l.State
	loans, err := s.PoloniecGetLoanOrders(l.Currency)
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

	l.CurrentLoanRate = lowest
	if l.CurrentLoanRate < 2 {
		CurrentLoanRate.Set(l.CurrentLoanRate) // Prometheus
		s.RecordPoloniexStatistics(l.CurrentLoanRate)
	}

	return nil
}

func abs(v float64) float64 {
	if v < 0 {
		return v * -1
	}
	return v
}

func (l *Lender) recordStatistics(username string, bals map[string]map[string]float64,
	inact map[string][]poloniex.PoloniexLoanOffer, activeLoan *poloniex.PoloniexActiveLoans) error {

	stats := new(userdb.UserStatistic)
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
	for _, l := range activeLoan.Used {
		if l.Currency == "BTC" {
			activeLentBal += l.Amount
			activeLentTotalRate += l.Rate
			activeLentCount++
		}
	}

	stats.ActiveLentBalance = activeLentBal
	stats.AverageActiveRate = activeLentTotalRate / activeLentCount

	// On Order

	inactiveLentBal := float64(0)
	inactiveLentTotalRate := float64(0)
	inactiveLentCount := float64(0)
	for _, l := range inact["BTC"] {
		if l.Currency == "BTC" {
			inactiveLentBal += l.Amount
			inactiveLentTotalRate += l.Rate
			inactiveLentCount++
		}
	}

	stats.OnOrderBalance = inactiveLentBal
	stats.AverageOnOrderRate = inactiveLentTotalRate / inactiveLentCount

	return l.State.RecordStatistics(stats)
}

// ProcessJob will calculate the newest loan rate, then it create a loan for 0.1 btc at that rate
// for the user in the Job
func (l *Lender) ProcessJob(j *Job) error {
	s := l.State

	bals, err := s.PoloniexGetAvailableBalances(j.Username)
	if err != nil {
		return err
	}

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
	// if !ok {
	// 	return fmt.Errorf("could not get available balances. Keys: %s, %s", "lending", l.Currency)
	// }

	if avail < MaxLendAmt {
		need := MaxLendAmt - avail
		if inactiveLoans != nil {
			currencyLoans := inactiveLoans[l.Currency]
			sort.Sort(poloniex.PoloniexLoanOfferArray(currencyLoans))
			for _, loan := range currencyLoans {
				if need < 0 {
					break
				}
				// Too close, no point in canceling
				if abs(loan.Rate-l.CurrentLoanRate) < 0.00000009 {
					continue
				}
				worked, err := s.PoloniexCancelLoanOffer(l.Currency, loan.ID, j.Username)
				if err != nil {
					break
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
	_, err = s.PoloniexCreateLoanOffer(l.Currency, amt, l.CurrentLoanRate, 2, false, j.Username)
	if err != nil {
		return err
	}
	LoansCreated.Inc()
	JobsDone.Inc()

	return nil
}
