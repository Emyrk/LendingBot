package bee

import (
	"strings"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var llog = log.WithFields(log.Fields{"package": "bee-lender"})

type Lender struct {
	Polo  *balancer.PoloniexAPIWithRateLimit
	Users []*LendUser
}

func NewLender() *Lender {
	l := new(Lender)
	l.Polo = balancer.NewPoloniexAPIWithRateLimit()

	return l
}

type LendUser struct {
	U balancer.User
}

func (l *Lender) ProcessPoloniexUser(u *LendUser) {
	var err error
	flog := llog.WithFields(log.Fields{"func": "ProcessPoloniexUser()", "user": u.U.Username})

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
	// if err != nil {
	// 	return fmt.Errorf("[T1-1] %s", err.Error())
	// }

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

	stats := userdb.NewAllUserStatistic()
	var activeLoans *poloniex.PoloniexActiveLoans
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
}

func (l *Lender) recordStatistics(username string, bals map[string]map[string]float64,
	inact map[string][]poloniex.PoloniexLoanOffer, activeLoan *poloniex.PoloniexActiveLoans) (*userdb.AllUserStatistic, error) {
	return nil, nil
}

/*
func (l *Lender) tierOneProcessJob(j *Job) error {
	var err error
	llog := clog.WithFields(log.Fields{
		"method": "T1ProcJob",
		"user":   j.Username,
	})

	s := l.State
	part1 := time.Now()
	//JobPart2
	bals := make(map[string]map[string]float64)
	// Try 3 times if timeout
	for i := 0; i < 3; i++ {
		bals, err = s.PoloniexGetAvailableBalances(j.Username)
		if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			llog.WithField("retry", i).Errorf("Error getting balances: %s", err.Error())
			if !strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				continue
			}
		}
		break
	}
	// if err != nil {
	// 	return fmt.Errorf("[T1-1] %s", err.Error())
	// }

	var inactiveLoans map[string][]poloniex.PoloniexLoanOffer
	// 3 types of balances: Not lent, Inactive, Active
	for i := 0; i < 3; i++ {
		inactiveLoans, err = s.PoloniexGetInactiveLoans(j.Username)
		if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			llog.WithField("retry", i).Errorf("Error getting inactive loans: %s", err.Error())
			if !strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				continue
			}
		}
		break
	}

	stats := userdb.NewAllUserStatistic()
	var activeLoans *poloniex.PoloniexActiveLoans
	for i := 0; i < 3; i++ {
		activeLoans, err = s.PoloniexGetActiveLoans(j.Username)
		if err == nil && activeLoans != nil {
			stats, err = l.recordStatistics(j.Username, bals, inactiveLoans, activeLoans)
			if err != nil {
				llog.WithField("retry", i).Errorf("Error in calculating statistic: %s", err.Error())
			}
		} else if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			llog.WithField("retry", i).Errorf("Error getting active loans: %s", err.Error())
			if !strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				continue
			}
		}
		break
	}

	JobPart1.Observe(float64(time.Since(part1).Nanoseconds()))
	part2 := time.Now()

	for i, min := range j.MinimumLend {
		// Move min from a % to it's value
		min = min / 100

		// Rate calculation
		rate := l.CurrentLoanRate[j.Currency[i]].AvgBased
		ri := l.rising(j.Currency[i])

		if l.CurrentLoanRate[j.Currency[i]].Simple == l.CurrentLoanRate[j.Currency[i]].AvgBased {
			if ri >= 1 {
				l.PoloniexStatLock.RLock()
				rate += l.PoloniexStats[j.Currency[i]].DayStd * .05
				l.PoloniexStatLock.RUnlock()
			}
		}

		if time.Since(l.LastPoloBotTime).Seconds() < 15 {
			l.LastPoloBotLock.RLock()
			v, ok := l.LastPoloBot[j.Currency[i]]
			l.LastPoloBotLock.RUnlock()
			if ok && v.GetBestReturnRate() > 0 {
				poloRate := v.GetBestReturnRate()
				if rate < poloRate {
					rate = (rate + poloRate) / 2
				}
			}
		}

		if j.Currency[i] == "BTC" {
			if rate < 2 {
				CompromisedBTC.Set(rate)
			} else {
				llog.Errorf("Rate is going to high. Trying to set at %f", rate)
			}
		}

		// End Rate Calculation

		avail, ok := bals["lending"][j.Currency[i]]
		var _ = ok

		maxLend := MaxLendAmt[j.Currency[i]]
		if ri == 2 {
			maxLend = maxLend * 2
		}

		if maxLend < avail*0.20 {
			maxLend = avail * 0.20
		}

		// rate := l.decideRate(rate, avail, total)

		// We need to find some more crypto to lkend
		//if avail < MaxLendAmt[j.Currency[i]] {
		need := maxLend - avail
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
					rate = min
				}

				// Too close, no point in canceling
				if abs(loan.Rate-rate) < 0.00000009 {
					continue
				}
				worked, err := s.PoloniexCancelLoanOffer(j.Currency[i], loan.ID, j.Username)
				if err != nil {
					llog.WithField("currency", j.Currency[i]).Errorf("[Cancel] Error in Lending: %s", err.Error())
					// log.Printf("[Cancel Loan] Error for %s (%s) : %s", j.Username, j.Currency[i], err.Error())
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

		if cStats, ok := stats.Currencies[j.Currency[i]]; ok {
			total := cStats.OnOrderBalance + cStats.ActiveLentBalance + cStats.AvailableBalance
			if total*0.1 > maxLend {
				maxLend = 0.1 * total
			}
		}

		amt := maxLend
		if avail < maxLend {
			amt = avail
		} else if avail < maxLend+MinLendAmt[j.Currency[i]] {
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
		if err != nil { //} && strings.Contains(err.Error(), "Too many requests") {
			// Sleep in our inner loop. This can be dangerous, should put these calls in a seperate queue to handle
			time.Sleep(5 * time.Second)
			_, err = s.PoloniexCreateLoanOffer(j.Currency[i], amt, rate, 2, false, j.Username)
			if err != nil {
				llog.WithFields(log.Fields{"currency": j.Currency[i], "user": j.Username}).Errorf("Error creating loan: %s", err.Error())
			}
		}

		llog.WithField("currency", j.Currency[i]).Infof("Created Loan: %f loaned at %f", amt, rate)
		if err != nil {
			llog.WithField("currency", j.Currency[i]).Errorf("[Offer] Error in Lending: %s", err.Error())
			continue
		}
		JobPart2.Observe(float64(time.Since(part2).Nanoseconds()))
		LoansCreated.Inc()
	}

	return nil
}

*/
