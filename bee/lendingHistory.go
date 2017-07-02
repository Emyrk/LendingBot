package balancer

// import (
// 	"fmt"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"github.com/Emyrk/LendingBot/src/core"
// 	"github.com/Emyrk/LendingBot/src/core/poloniex"
// 	"github.com/Emyrk/LendingBot/src/core/userdb"
// 	log "github.com/sirupsen/logrus"
// )

// var AvaiableCoins = []string{
// 	"BTC",
// 	"BTS",
// 	"CLAM",
// 	"DOGE",
// 	"DASH",
// 	"LTC",
// 	"MAID",
// 	"STR",
// 	"XMR",
// 	"XRP",
// 	"ETH",
// 	"FCT",
// }

// var llog = clog.WithField("subpackage", "LendingHistory")

// type LendingHistoryKeeper struct {
// 	WorkingOn  map[string]time.Time
// 	WorkOnLock sync.RWMutex
// 	LocalCall  bool
// }

// func NewLendingHistoryKeeper(s *core.State) *LendingHistoryKeeper {
// 	l := new(LendingHistoryKeeper)
// 	l.WorkingOn = make(map[string]time.Time)

// 	return l
// }

// func (l *LendingHistoryKeeper) SaveMonth(username string) {
// 	l.WorkOnLock.RLock()
// 	v, ok := l.WorkingOn[username]
// 	l.WorkOnLock.RUnlock()
// 	if !ok {
// 		l.WorkingOn[username] = time.Now()
// 	} else {
// 		// If done within 5hrs, don't bother
// 		if time.Since(v).Seconds() < 60*60*10 {
// 			return
// 		}
// 	}

// 	l.WorkOnLock.Lock()
// 	l.WorkingOn[username] = time.Now()
// 	l.WorkOnLock.Unlock()
// 	defer func() {
// 		l.WorkOnLock.Lock()
// 		l.WorkingOn[username] = time.Now()
// 		l.WorkOnLock.Unlock()
// 	}()

// 	flog := llog.WithField("func", "SaveMonth()")
// 	if len(l.St.Master.Connections) == 0 && !l.LocalCall {
// 		flog.WithField("user", username).Errorf("No slaves to make lending hist calls")
// 		return
// 	}

// 	n := time.Now().UTC()
// 	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
// 	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
// 	// Must start 2 days back to ensure all loans covered
// 	top = top.Add(-24 * time.Hour)
// 	curr := top.Add(time.Hour * -72).Add(1 * time.Second)
// 	for i := 0; i < 28; i++ {
// 		v, err := l.St.LoadLendingSummary(username, curr)
// 		if v == nil || err != nil {
// 			resp, err := l.getLendhist(username, fmt.Sprintf("%d", curr.Unix()-1), fmt.Sprintf("%d", top.Unix()), "")
// 			if err != nil {
// 				flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error getting Lending history: %s", err.Error())
// 				break
// 			} else {
// 				compiled, err := compileData(resp.Data, top)
// 				if err != nil {
// 					flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error compiling Lending history: %s", err.Error())
// 					break
// 				} else {
// 					compiled.Username = username
// 					compiled.SetTime(top)
// 					err := l.St.SaveLendingHistory(compiled)
// 					if err != nil {
// 						flog.WithFields(log.Fields{"time": top.String()}).Errorf("Error saving Lending history: %s", err.Error())
// 						break
// 					}
// 				}
// 			}
// 		}

// 		top = top.Add(-24 * time.Hour)
// 		curr = curr.Add(-24 * time.Hour)
// 	}
// }

// func (l *LendingHistoryKeeper) getLendhist(username, start, end, limit string) (resp poloniex.PoloniexAuthentictedLendingHistoryRespone, err error) {
// 	if l.LocalCall {
// 		return l.St.PoloniexAuthenticatedLendingHistory(username, start, end, limit)
// 	} else {
// 		return l.St.PoloniexOffloadAuthenticatedLendingHistory(username, start, end, limit)
// 	}
// }

// func compileData(data []poloniex.PoloniexAuthentictedLendingHistory, t time.Time) (*userdb.AllLendingHistoryEntry, error) {
// 	ent := userdb.NewAllLendingHistoryEntry()
// 	flog := llog.WithField("func", "compileData()")
// 	// for _, v := range AvaiableCoins {
// 	// 	ent.Data[v] = new(userdb.LendingHistoryEntry)
// 	// }
// 	for _, d := range data {
// 		// fmt.Println(d)
// 		//var dt time.Time
// 		dt, err := time.Parse("2006-01-02 15:04:05", d.Close)
// 		//err := dt.UnmarshalText([]byte(d.Close))
// 		if err != nil {
// 			flog.WithFields(log.Fields{"item": "Close", "currency": d.Currency}).Errorf("Error in parsing time: Raw: %s, err: %s", d.Close, err.Error())
// 			continue
// 		}
// 		if dt.Day() != t.Day() {
// 			continue
// 		}
// 		if _, ok := ent.Data[d.Currency]; !ok {
// 			e := new(userdb.LendingHistoryEntry)
// 			e.Currency = d.Currency
// 			ent.Data[d.Currency] = e
// 		}

// 		dur, err := strconv.ParseFloat(d.Duration, 64)
// 		if err == nil {
// 			ent.Data[d.Currency].AvgDuration += dur
// 		} else {
// 			flog.WithFields(log.Fields{"item": "AvgDur", "currency": d.Currency}).Errorf("Error parsing int: Raw: %s, err: %s", d.Duration, err.Error())
// 		}

// 		f, err := strconv.ParseFloat(d.Fee, 64)
// 		if err == nil {
// 			ent.Data[d.Currency].Fees += f * -1
// 		} else {
// 			flog.WithFields(log.Fields{"item": "Fees", "currency": d.Currency}).Errorf("Error parsing float: Raw: %s, err: %s", d.Fee, err.Error())
// 		}

// 		e, err := strconv.ParseFloat(d.Earned, 64)
// 		if err == nil {
// 			ent.Data[d.Currency].Earned += e
// 		} else {
// 			flog.WithFields(log.Fields{"item": "Earned", "currency": d.Currency}).Errorf("Error parsing float: Raw: %s, err: %s", d.Earned, err.Error())

// 		}

// 		ent.Data[d.Currency].LoanCounts++
// 	}

// 	for k := range ent.Data {
// 		if ent.Data[k].LoanCounts > 0 {
// 			ent.Data[k].AvgDuration = ent.Data[k].AvgDuration / float64(ent.Data[k].LoanCounts)
// 		}
// 	}

// 	return ent, nil
// }
