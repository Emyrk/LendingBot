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

var poloLogger = log.WithFields(log.Fields{
	"pacakge": "bee",
	"file":    "PoloniexLender",
})

type PoloniexLender struct {
	Polo *balancer.PoloniexAPIWithRateLimit
	Bee  *Bee
	GS   IGlobalServer

	usersDoneLock sync.RWMutex
	usersDone     map[string]time.Time

	recordMapLock sync.RWMutex
	recordMap     map[int]map[string]time.Time

	// Report
	usercycles int
	cycles     int
}

func NewPoloniexLender(b *Bee, g IGlobalServer, polo *balancer.PoloniexAPIWithRateLimit) *PoloniexLender {
	l := new(PoloniexLender)
	l.Bee = b
	l.Polo = polo
	l.GS = g
	l.recordMap = make(map[int]map[string]time.Time)
	l.recordMap[balancer.PoloniexExchange] = make(map[string]time.Time)
	l.recordMap[balancer.BitfinexExchange] = make(map[string]time.Time)

	l.usersDone = make(map[string]time.Time)

	return l
}

func (l *PoloniexLender) Report() string {
	return fmt.Sprintf("Cycles: %d, UsersProcesses: %d", l.cycles, l.usercycles)
}

func (l *PoloniexLender) Close() {

}

func (l *PoloniexLender) ProcessPoloniexUser(u *LendUser) error {
	historySaved := false
	notes := ""
	defer func(monthtoo bool, n string) {
		if monthtoo {
			l.Bee.updateUser(u.U.Username, u.U.Exchange, n, time.Now(), time.Now())
		} else {
			l.Bee.updateUser(u.U.Username, u.U.Exchange, n, time.Now(), time.Time{})
		}
	}(historySaved, notes)
	dbu, err := l.Bee.FetchUser(u.U.Username)
	if err != nil {
		return err
	}

	// If lending halted, do not lend
	if dbu.LendingHalted.Halt {
		// TODO: Renable not lending if not paid
		// Commenting this out makes Hodl.zone free
		// return nil
	}

	l.usersDoneLock.RLock()
	v, _ := l.usersDone[u.U.Username]
	l.usersDoneLock.RUnlock()

	// Only process once per 15s max
	if time.Since(v) < time.Second*15 {
		return nil
	}

	l.usersDoneLock.Lock()
	l.usersDone[u.U.Username] = time.Now()
	l.usersDoneLock.Unlock()

	logentry := fmt.Sprintf("PoloniexBot analyzed your account and found nothing needed to be done")

	if u.U.AccessKey == "" {
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("PoloniexBot could not find an access key"))
		return fmt.Errorf("No access key for user %s", u.U.Username)
	}

	if u.U.SecretKey == "" {
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("PoloniexBot could not find a secret key"))
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
			flog.WithField("retry", i).Errorf("Error getting balances: %s", shortError(err).Error())
			if strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				if i == 2 {
					logentry = fmt.Sprintf("PoloniexBot encounterd an error getting available balances: %s", err.Error())
					l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)
					return err
				}
				continue
			} else {
				logentry = fmt.Sprintf("PoloniexBot encounterd an error getting available balances: %s", err.Error())
				l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)
				return err
			}
		}
		break
	}

	var inactiveLoans map[string][]poloniex.PoloniexLoanOffer
	// 3 types of balances: Not lent, Inactive, Active
	for i := 0; i < 3; i++ {
		inactiveLoans, err = l.Polo.PoloniexGetInactiveLoans(u.U.AccessKey, u.U.SecretKey)
		if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			flog.WithField("retry", i).Errorf("Error getting inactive loans: %s", shortError(err).Error())
			if strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				if i == 2 {
					logentry = fmt.Sprintf("PoloniexBot encounterd an error getting inactive loans: %s", err.Error())
					l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)
					return err
				}
				continue
			} else {
				logentry = fmt.Sprintf("PoloniexBot encounterd an error getting inactive loans: %s", err.Error())
				l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)
				return err
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
				flog.WithField("retry", i).Errorf("Error in calculating statistic: %s", shortError(err).Error())
			}
		} else if err != nil && !strings.Contains(err.Error(), "Unable to JSON Unmarshal response. Resp: []") {
			flog.WithField("retry", i).Errorf("Error getting active loans: %s", shortError(err).Error())
			if strings.Contains(err.Error(), "Connection timed out. Please try again.") {
				// Let it retry
				if i == 2 {
					logentry = fmt.Sprintf("PoloniexBot encounterd an error getting active loans: %s", err.Error())
					l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)
					return err
				}
				continue
			} else {
				logentry = fmt.Sprintf("PoloniexBot encounterd an error getting active loans: %s", err.Error())
				l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)
				return err
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

	currLogs := ""
	for _, curr := range dbu.PoloniexEnabled.Keys() { //u.U.MinimumLend {
		min := dbu.PoloniexMiniumLend.Get(curr)
		clog := flog.WithFields(log.Fields{"currency": curr, "exchange": balancer.GetExchangeString(u.U.Exchange)})

		// Move min from a % to it's value
		min = min / 100

		// Rate calculation
		lr, _ := l.GS.GetLoanRate(balancer.PoloniexExchange, curr)
		rate := lr.AvgBased
		if rate == 0 { // No rate
			continue
		}

		// Set to min if we are below
		if rate < min {
			rate = min
		}

		if curr == "BTC" {
			if rate < 2 {
				CompromisedBTC.Set(rate)
			} else {
				msg := fmt.Sprintf("Rate is going to high. Trying to %s set at %f", curr, rate)
				clog.Errorf(msg)
				notes += msg + "\n"
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
					msg := fmt.Sprintf("[Cancel] Error in Lending: %s", shortError(err).Error())
					clog.Errorf(msg)
					notes += msg + "\n"
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
			msg := fmt.Sprintf("Error creating loan: %s", shortError(err).Error())
			clog.Errorf(msg)
			notes += msg + "\n"

			currLogs += fmt.Sprintf("   PoloniexBot attempted to create a loan for %f %s at %f, but failed: %s\n", amt, curr, rate, shortError(err).Error())
		} else {
			clog.WithFields(log.Fields{"rate": rate, "amount": amt}).Infof("Created Loan")
			notes += fmt.Sprintf("Created loan for %f %s at %f\n", amt, curr, rate)
			currLogs += fmt.Sprintf("   Loan made for %f %s at %f\n", amt, curr, rate)
		}
		JobPart2.Observe(float64(time.Since(part2).Nanoseconds()))
		LoansCreated.Inc()
	}
	if len(currLogs) > 0 {
		logentry = fmt.Sprintf("PoloniexBot Lending Actions:\n%s", currLogs)
	}
	l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)

	l.usersDoneLock.Lock()
	l.usersDone[u.U.Username] = time.Now().Add(15 * time.Second)
	l.usersDoneLock.Unlock()

	historySaved = l.GS.SavePoloniexMonth(dbu, u.U.AccessKey, u.U.SecretKey)
	if historySaved {
		u.U.LastHistorySaved = time.Now()
	}

	return nil
}

func (l *PoloniexLender) recordStatistics(username string, bals map[string]map[string]float64,
	inact map[string][]poloniex.PoloniexLoanOffer, activeLoan *poloniex.PoloniexActiveLoans) (*userdb.AllUserStatistic, error) {

	// Make stats
	stats := userdb.NewAllUserStatistic()
	stats.Time = time.Now()
	stats.Username = username
	stats.Exchange = userdb.PoloniexExchange

	for _, v := range balancer.Currencies[balancer.PoloniexExchange] {
		var last float64 = 1
		if v != "BTC" {
			lastTicker, _ := l.GS.GetPoloniexTicker(fmt.Sprintf("BTC_%s", v))
			last = lastTicker.Last
		}
		cstat := userdb.NewUserStatistic(v, last)
		stats.Currencies[v] = cstat
	}

	// Avail balance
	for k, v := range bals["lending"] {
		if !math.IsNaN(v) {
			stats.Currencies[k].AvailableBalance = v
			stats.TotalCurrencyMap[k] += l.GS.GetBTCAmount(v, k)
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
		stats.TotalCurrencyMap[loan.Currency] += l.GS.GetBTCAmount(loan.Amount, loan.Currency)
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

			stats.TotalCurrencyMap[k] += l.GS.GetBTCAmount(loan.Amount, k)
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
	stats.Exchange = userdb.PoloniexExchange
	err := l.Bee.SaveUserStastics(stats)

	l.recordMapLock.Lock()
	l.recordMap[balancer.PoloniexExchange][username] = time.Now()
	l.recordMapLock.Unlock()
	return stats, err
}
