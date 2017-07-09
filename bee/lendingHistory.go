package bee

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	// "github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var AvaiableCoins = []string{
	"BTC",
	"BTS",
	"CLAM",
	"DOGE",
	"DASH",
	"LTC",
	"MAID",
	"STR",
	"XMR",
	"XRP",
	"ETH",
	"FCT",
}

var llog = generalBeeLogger.WithField("subpackage", "LendingHistory")

type LendingHistoryKeeper struct {
	WorkingOn  map[string]time.Time
	WorkOnLock sync.RWMutex
	LocalCall  bool

	MyBee *Bee
}

func NewLendingHistoryKeeper(b *Bee) *LendingHistoryKeeper {
	l := new(LendingHistoryKeeper)
	l.WorkingOn = make(map[string]time.Time)
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
		l.WorkingOn[u] = n
	}
}

func (l *LendingHistoryKeeper) SaveMonth(username, accesskey, secretkey string) bool {
	flog := llog.WithField("func", "SaveMonth()")
	l.WorkOnLock.RLock()
	v, ok := l.WorkingOn[username]
	l.WorkOnLock.RUnlock()
	if !ok {
		l.WorkingOn[username] = time.Now()
	} else {
		// If done within 5hrs, don't bother
		if time.Since(v).Seconds() < 60*60*10 {
			return false
		}
	}

	l.WorkOnLock.Lock()
	l.WorkingOn[username] = time.Now()
	l.WorkOnLock.Unlock()
	defer func() {
		l.WorkOnLock.Lock()
		l.WorkingOn[username] = time.Now()
		l.WorkOnLock.Unlock()
	}()
	start := time.Now()
	flog.WithField("user", username).Infof("Saving Month starting")

	skipped := 0
	total := float64(0)

	n := time.Now().UTC()
	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
	// Must start 2 days back to ensure all loans covered
	top = top.Add(-24 * time.Hour)
	curr := top.Add(time.Hour * -72).Add(1 * time.Second)
	for i := 0; i < 28; i++ {
		per := time.Now()
		flog.Infof("Username: %s, Top: %s", username, top.String())

		v, err := l.MyBee.userStatDB.GetLendHistorySummary(username, top) //l.St.LoadLendingSummary(username, curr)
		if v == nil || err != nil {
			resp, err := l.getLendhist(accesskey, secretkey, fmt.Sprintf("%d", curr.Unix()-1), fmt.Sprintf("%d", top.Unix()), "")
			if err != nil {
				flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error getting Lending history: %s", err.Error())
				break
			} else {
				compiled, err := compileData(resp.Data, top)
				if err != nil {
					flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error compiling Lending history: %s", err.Error())
					break
				} else {
					compiled.Username = username
					compiled.SetTime(top)
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

	flog.WithField("user", username).Infof("Saving month completed in %fs. %d Skipped, avg %fs", time.Since(start).Seconds(), skipped, total/28)
	return true
}

func (l *LendingHistoryKeeper) getLendhist(accessKey, secret, start, end, limit string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
	return l.MyBee.LendingBot.Polo.PoloniexAuthenticatedLendingHistory(accessKey, secret, start, end, limit)
}

func compileData(data []poloniex.PoloniexAuthentictedLendingHistory, t time.Time) (*userdb.AllLendingHistoryEntry, error) {
	ent := userdb.NewAllLendingHistoryEntry()
	flog := llog.WithField("func", "compileData()")
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
			ent.PoloniexData[d.Currency].Fees += f * -1
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
