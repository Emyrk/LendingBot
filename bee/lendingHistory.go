package bee

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	// "github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/bitfinex"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var llog = generalBeeLogger.WithField("subpackage", "LendingHistory")

type LendingHistoryKeeper struct {
	WorkingOnPolo  map[string]time.Time
	WorkOnLockPolo sync.RWMutex

	WorkingOnBit  map[string]time.Time
	WorkOnLockBit sync.RWMutex

	LocalCall bool

	MyBee *Bee
}

func NewLendingHistoryKeeper(b *Bee) *LendingHistoryKeeper {
	l := new(LendingHistoryKeeper)
	l.WorkingOnPolo = make(map[string]time.Time)
	l.WorkingOnBit = make(map[string]time.Time)
	l.MyBee = b

	// Make sure users don't all save back to back
	userStrings := []string{}
	for _, u := range b.Users {
		userStrings = append(userStrings, u.Username)
	}
	l.InitRandomTimes(userStrings)

	return l
}

func (l *LendingHistoryKeeper) InitRandomTimes(users []string) {
	n := time.Now()
	for _, u := range users {
		n = n.Add(time.Minute * 5)
		l.WorkingOnPolo[u] = n
		l.WorkingOnBit[u] = n
	}
}

func (l *LendingHistoryKeeper) FindStart(username string, startTime time.Time, exch int) int {
	start := 0
	prev := 0
	for i := 0; i < 4; i++ {

		v, err := l.MyBee.userStatDB.GetLendHistorySummary(username, startTime.Add(-24*time.Hour*time.Duration(i)))
		if v == nil || err != nil {
			start = prev * 9
			return start
		}

		set := v.PoloSet
		if exch == balancer.BitfinexExchange {
			set = v.BitfinSet
		}
		if !set {
			start = prev * 9
			return start
		}

		prev = i
		start = i * 9
	}
	return start
}

func (l *LendingHistoryKeeper) SavePoloniexMonth(username, accesskey, secretkey string) bool {
	flog := llog.WithField("func", "SavePoloniexMonth").WithField("exch", "Poloniex")
	l.WorkOnLockPolo.RLock()
	v, ok := l.WorkingOnPolo[username]
	l.WorkOnLockPolo.RUnlock()
	if !ok {
		l.WorkingOnPolo[username] = time.Now()
	} else {
		// If done within 5hrs, don't bother
		if time.Since(v).Seconds() < 60*60*10 {
			return false
		}
	}

	l.WorkOnLockPolo.Lock()
	l.WorkingOnPolo[username] = time.Now()
	l.WorkOnLockPolo.Unlock()
	defer func() {
		l.WorkOnLockPolo.Lock()
		l.WorkingOnPolo[username] = time.Now()
		l.WorkOnLockPolo.Unlock()
	}()
	start := time.Now()
	flog.WithField("user", username).Infof("[P] Saving Month starting")

	skipped := 0
	total := float64(0)

	n := time.Now().UTC()
	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
	// Must start 2 days back to ensure all loans covered
	top = top.Add(-24 * time.Hour)
	curr := top.Add(time.Hour * -72).Add(1 * time.Second)
	for i := l.FindStart(username, top, balancer.PoloniexExchange); i < 30; i++ {
		per := time.Now()
		flog.Infof("[P] Username: %s, Top: %s", username, top.String())

		v, err := l.MyBee.userStatDB.GetLendHistorySummary(username, top) //l.St.LoadLendingSummary(username, curr)
		if v == nil || err != nil || !v.PoloSet {
			resp, err := l.getPoloLendhist(accesskey, secretkey, fmt.Sprintf("%d", curr.Unix()-1), fmt.Sprintf("%d", top.Unix()), "")
			if err != nil {
				flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error getting Lending history: %s", err.Error())
				break
			} else {
				compiled, err := compilePoloniexData(resp.Data, top, v)
				if err != nil {
					flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error compiling Lending history: %s", err.Error())
					break
				} else {
					compiled.Username = username
					compiled.SetTime(top)
					compiled.PoloSet = true
					err := l.MyBee.userStatDB.SaveLendingHistory(compiled)
					if err != nil {
						flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error saving Lending history: %s", err.Error())
						break
					} else {
						// l.MyBee.
						for _, loan := range resp.Data {
							err := l.MyBee.AddPoloniexDebt(username, loan)
							if err != nil {
								// This person was not charged
								flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error charging user: %s", err.Error())
							}
						}
					}
				}
			}
		} else {
			skipped++
		}

		top = top.Add(-24 * time.Hour)
		curr = curr.Add(-24 * time.Hour)
		total += time.Since(per).Seconds()
	}

	flog.WithField("user", username).Infof("[P] Saving month completed in %fs. %d Skipped, avg %fs", time.Since(start).Seconds(), skipped, total/28)
	return true
}

func (l *LendingHistoryKeeper) SaveBitfinexMonth(username, accesskey, secretkey string) bool {
	flog := llog.WithField("func", "SaveBitfinexMonth").WithField("exch", "Bitfinex")
	l.WorkOnLockBit.RLock()
	v, ok := l.WorkingOnBit[username]
	l.WorkOnLockBit.RUnlock()
	if !ok {
		l.WorkingOnBit[username] = time.Now()
	} else {
		// If done within 5hrs, don't bother
		if time.Since(v).Seconds() < 60*60*10 {
			return false
		}
	}

	l.WorkOnLockBit.Lock()
	l.WorkingOnBit[username] = time.Now()
	l.WorkOnLockBit.Unlock()
	defer func() {
		l.WorkOnLockBit.Lock()
		l.WorkingOnBit[username] = time.Now()
		l.WorkOnLockBit.Unlock()
	}()
	start := time.Now()
	flog.WithField("user", username).Infof("[B] Saving Month starting")

	skipped := 0
	total := float64(0)

	n := time.Now().UTC()
	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
	// Must start 2 days back to ensure all loans covered
	top = top.Add(-24 * time.Hour)
	curr := top.Add(time.Hour * -24).Add(1 * time.Second)
	for i := l.FindStart(username, top, balancer.BitfinexExchange); i < 30; i++ {
		per := time.Now()
		flog.Infof("[B] Username: %s, Top: %s", username, top.String())

		v, err := l.MyBee.userStatDB.GetLendHistorySummary(username, top) //l.St.LoadLendingSummary(username, curr)
		if v == nil || err != nil || !v.BitfinSet {
			resp, err := l.getBitfinLendhist(accesskey, secretkey, fmt.Sprintf("%d", curr.Unix()-1), fmt.Sprintf("%d", top.Unix()), "")
			if err != nil {
				flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error getting Lending history: %s", err.Error())
				break
			} else {
				compiled, err := compileBitfinexData(resp, top, v)
				if err != nil {
					flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error compiling Lending history: %s", err.Error())
					break
				} else {
					compiled.Username = username
					compiled.SetTime(top)
					compiled.BitfinSet = true
					err := l.MyBee.userStatDB.SaveLendingHistory(compiled)
					if err != nil {
						flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error saving Lending history: %s", err.Error())
						break
					}
				}
			}
		} else {
			skipped++
		}

		top = top.Add(-24 * time.Hour)
		curr = curr.Add(-24 * time.Hour)
		total += time.Since(per).Seconds()
	}

	flog.WithField("user", username).Infof("[B] Saving month completed in %fs. %d Skipped, avg %fs", time.Since(start).Seconds(), skipped, total/28)
	return true
}

func (l *LendingHistoryKeeper) getPoloLendhist(accessKey, secret, start, end, limit string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
	return l.MyBee.LendingBot.Polo.PoloniexAuthenticatedLendingHistory(accessKey, secret, start, end, limit)
}

func (l *LendingHistoryKeeper) getBitfinLendhist(accessKey, secret, start, end, limit string) (resp []bitfinex.FundingEarning, err error) {
	api := bitfinex.New(accessKey, secret)
	return api.GetFundingEarnings(start, end)
}

func compilePoloniexData(data []poloniex.PoloniexAuthentictedLendingHistory, t time.Time, ent *userdb.AllLendingHistoryEntry) (*userdb.AllLendingHistoryEntry, error) {
	if ent == nil {
		ent = userdb.NewAllLendingHistoryEntry()
	}
	flog := llog.WithField("func", "compilePoloniexData()")
	// for _, v := range AvaiableCoins {
	// 	ent.Data[v] = new(userdb.LendingHistoryEntry)
	// }
	for _, d := range data {
		// fmt.Println(d)
		//var dt time.Time
		dt, err := time.Parse("2006-01-02 15:04:05", d.Close)
		//err := dt.UnmarshalText([]byte(d.Close))
		if err != nil {
			flog.WithFields(log.Fields{"item": "Close", "currency": d.Currency}).Errorf("Error in parsing time: Raw: %s, err: %s", d.Close, err.Error())
			continue
		}
		if dt.Day() != t.Day() {
			continue
		}
		if _, ok := ent.PoloniexData[d.Currency]; !ok {
			e := new(userdb.LendingHistoryEntry)
			e.Currency = d.Currency
			ent.PoloniexData[d.Currency] = e
		}

		dur, err := strconv.ParseFloat(d.Duration, 64)
		if err == nil {
			ent.PoloniexData[d.Currency].AvgDuration += dur
		} else {
			flog.WithFields(log.Fields{"item": "AvgDur", "currency": d.Currency}).Errorf("Error parsing int: Raw: %s, err: %s", d.Duration, err.Error())
		}

		f, err := strconv.ParseFloat(d.Fee, 64)
		if err == nil {
			ent.PoloniexData[d.Currency].Fees += f
		} else {
			flog.WithFields(log.Fields{"item": "Fees", "currency": d.Currency}).Errorf("Error parsing float: Raw: %s, err: %s", d.Fee, err.Error())
		}

		e, err := strconv.ParseFloat(d.Earned, 64)
		if err == nil {
			ent.PoloniexData[d.Currency].Earned += e
		} else {
			flog.WithFields(log.Fields{"item": "Earned", "currency": d.Currency}).Errorf("Error parsing float: Raw: %s, err: %s", d.Earned, err.Error())

		}

		ent.PoloniexData[d.Currency].LoanCounts++
	}

	for k := range ent.PoloniexData {
		if ent.PoloniexData[k].LoanCounts > 0 {
			ent.PoloniexData[k].AvgDuration = ent.PoloniexData[k].AvgDuration / float64(ent.PoloniexData[k].LoanCounts)
		}
	}

	return ent, nil
}

func compileBitfinexData(data []bitfinex.FundingEarning, t time.Time, ent *userdb.AllLendingHistoryEntry) (*userdb.AllLendingHistoryEntry, error) {
	if ent == nil {
		ent = userdb.NewAllLendingHistoryEntry()
	}
	flog := llog.WithField("func", "compileBitfinexData()")
	// for _, v := range AvaiableCoins {
	// 	ent.Data[v] = new(userdb.LendingHistoryEntry)
	// }
	for _, d := range data {
		currency := d.Currency
		if strings.ToLower(currency) == "DSH" {
			currency = "DSH"
		}
		currency = strings.ToUpper(currency)

		if _, ok := ent.BitfinexData[currency]; !ok {
			e := new(userdb.LendingHistoryEntry)
			e.Currency = currency
			ent.BitfinexData[currency] = e
		}

		amt, err := strconv.ParseFloat(d.Amount, 64)
		if err == nil {
			fee := (amt * 0.15)
			ent.BitfinexData[currency].Fees += fee * -1
			ent.BitfinexData[d.Currency].Earned += amt - fee

		} else {
			flog.WithFields(log.Fields{"item": "Amount", "currency": d.Currency}).Errorf("Error parsing float: Raw: %s, err: %s", d.Amount, err.Error())
		}

		ent.BitfinexData[d.Currency].LoanCounts++
	}

	return ent, nil
}
