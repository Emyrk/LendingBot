package bee

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/bitfinex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/beefsack/go-rate"
	log "github.com/sirupsen/logrus"
)

type BitfinexLender struct {
	tickerlock    sync.RWMutex
	Ticker        map[string]bitfinex.V2Ticker
	FundingTicker map[string]bitfinex.V2FundingTicker

	usersDoneLock sync.RWMutex
	usersDone     map[string]time.Time

	API         *bitfinex.API
	rateLimiter *rate.RateLimiter
	spamLimiter *rate.RateLimiter

	nextStart time.Time

	iotLastTime time.Time
	iotLast     float64

	eosLastTime time.Time
	eosLast     float64

	quit chan bool
}

func (bl *BitfinexLender) TickerInfo() string {
	bl.tickerlock.RLock()
	str := ""
	str += "-- Tickers --\n"
	for k, t := range bl.Ticker {
		str += fmt.Sprintf(" %s: %f\n", k, t.LastPrice)
	}
	str += "-- Funding --\n"
	for k, t := range bl.FundingTicker {
		str += fmt.Sprintf(" %s: %f\n", k, t.LastPrice)
	}
	bl.tickerlock.RUnlock()
	return str
}

func NewBitfinexLender() *BitfinexLender {
	b := new(BitfinexLender)
	b.API = bitfinex.New("Public", "Calls")
	b.Ticker = make(map[string]bitfinex.V2Ticker)
	b.FundingTicker = make(map[string]bitfinex.V2FundingTicker)
	b.rateLimiter = rate.New(90, time.Minute)
	b.spamLimiter = rate.New(15, time.Second)
	b.usersDone = make(map[string]time.Time)
	b.quit = make(chan bool)

	b.GetTickers()
	return b
}

func (bl *BitfinexLender) Run() {
	go bl.TickerLoop()
}

func (bl *BitfinexLender) Close() {
	bl.quit <- true
}

func (bl *BitfinexLender) take() error {
	ok, remain := bl.rateLimiter.Try()
	if ok {
		return nil
	}
	bl.spamLimiter.Wait()
	if remain < time.Second*1 {
		time.Sleep(remain)
		return nil
	}
	bl.nextStart = time.Now().Add(remain)
	return fmt.Errorf("Don't spam Bitfinex. Have to sleep %s before calling again", remain.Seconds())
}

func (l *Lender) ProcessBitfinexUser(u *LendUser) error {
	flog := poloLogger.WithFields(log.Fields{"func": "ProcessBitfinexUser()", "user": u.U.Username, "exchange": balancer.GetExchangeString(u.U.Exchange)})

	historySaved := false
	bl := l.BitfinLender
	bl.usersDoneLock.RLock()
	v, _ := bl.usersDone[u.U.Username]
	bl.usersDoneLock.RUnlock()

	defer func() {
		bl.usersDoneLock.Lock()
		bl.usersDone[u.U.Username] = time.Now()
		bl.usersDoneLock.Unlock()
	}()

	// Only process once per minute max
	if time.Since(v) < time.Minute { //time.Minute {
		flog.Warningf("Too short %v", time.Since(v))
		return nil
	}

	// Have to wait before making another call
	if time.Now().Before(bl.nextStart) {
		flog.Warningf("Too many calls, must wait %f seconds", time.Since(bl.nextStart).Seconds()*-1)
		return nil
	}
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
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error fetching your account"))
		return err
	}

	api := bitfinex.New(u.U.AccessKey, u.U.SecretKey)

	// api.Ticker(symbol)
	err = bl.take()
	if err != nil {
		flog.Error(err)
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error: %s", err.Error()))
		return err
	}
	bals, err := api.WalletBalances()
	if err != nil {
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error fetching balances: %s", err.Error()))
		return err
	}

	// Inactive
	err = bl.take()
	if err != nil {
		flog.Error(err)
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error: %s", err.Error()))
		return err
	}
	inactMap := make(map[string]bitfinex.Offers)
	inactiveOffers, err := api.ActiveOffers()
	if err != nil {
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error fetching loans: %s", err.Error()))
		return err
	}
	for _, o := range inactiveOffers {
		if strings.ToLower(o.Direction) != "lend" {
			continue
		}
		inactMap[correctCurencyString(o.Currency)] = append(inactMap[correctCurencyString(o.Currency)], o)
	}

	// Active
	err = bl.take()
	if err != nil {
		flog.Error(err)
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error: %s", err.Error()))
		return err
	}
	activeMap := make(map[string]bitfinex.Credits)
	activeOffers, err := api.ActiveCredits()
	if err != nil {
		l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error fetching active loans: %s", err.Error()))
		return err
	}
	for _, o := range activeOffers {
		activeMap[correctCurencyString(o.Currency)] = append(activeMap[correctCurencyString(o.Currency)], o)
	}

	_, err = l.recordBitfinexStatistics(u.U.Username, bals, inactMap, activeMap)
	if err != nil {
		flog.Warningf("Failed to record Bitfinex Statistics: %s", err.Error())
	}

	logmsg := ""
	for _, c := range dbu.BitfinexEnabled.Keys() {
		clog := flog.WithFields(log.Fields{"currency": c})

		lower := strings.ToLower(c)
		if lower == "dash" {
			lower = "dsh"
		}

		// You got no money buddy
		if bals[bitfinex.WalletKey{"deposit", lower}].Amount == 0 {
			continue
		}

		avail := bals[bitfinex.WalletKey{"deposit", lower}].Available

		var last float64 = 0
		bl.tickerlock.RLock()
		t, ok := bl.FundingTicker[fmt.Sprintf("f%s", correctCurencyString(c))]
		if ok {
			last = t.FRR //t.FRR
		}
		bl.tickerlock.RUnlock()

		for _, l := range inactMap[correctCurencyString(c)] {
			dif := abs((l.Rate / 365 / 100) - last)
			if dif > 0.0001 {
				api.CancelOffer(l.ID)
				avail += l.RemainingAmount
			}
		}

		if avail <= 0 {
			continue
		}

		err = bl.take()
		if err != nil {
			flog.Error(err)
			l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error: %s", err.Error()))
			return err
		}
		o, err := api.NewOffer(lower, avail, last, 2, "lend")
		if err != nil {
			//l.Bee.AddBotActivityLogEntry(u.U.Username, fmt.Sprintf("BitfinexBot encountered an error creating loan: %s", err.Error()))
			logmsg += fmt.Sprintf("   Loan made for %f %s at %f\n", avail, c, last)
		}
		var _ = o
		var _ = t

		clog.WithFields(log.Fields{"rate": last, "amount": avail}).Infof("Created Loan")
		var _ = avail
	}

	logentry := fmt.Sprintf("BitfinexBot analyzed your account and found nothing needed to be done")
	if len(logmsg) > 0 {
		logentry = fmt.Sprintf("BitfinexBot Lending Actions:\n%s", logmsg)
	}

	l.Bee.AddBotActivityLogEntry(u.U.Username, logentry)

	historySaved = l.HistoryKeeper.SaveBitfinexMonth(u.U.Username, u.U.AccessKey, u.U.SecretKey)
	if historySaved {
		u.U.LastHistorySaved = time.Now()
	}

	return nil
}

func correctCurencyString(cur string) string {
	c := strings.ToUpper(cur)
	if cur == "DSH" {
		return "DASH"
	}
	return c
}

func (l *Lender) recordBitfinexStatistics(username string,
	bals map[bitfinex.WalletKey]bitfinex.WalletBalance,
	inact map[string]bitfinex.Offers,
	activeLoan map[string]bitfinex.Credits) (*userdb.AllUserStatistic, error) {

	// Make stats
	stats := userdb.NewAllUserStatistic()
	stats.Time = time.Now()
	stats.Username = username
	stats.Exchange = userdb.BitfinexExchange

	// Ticker
	for _, v := range balancer.Currencies[balancer.BitfinexExchange] {
		uppered := correctCurencyString(v)
		lowered := strings.ToLower(v)
		if lowered == "dash" {
			lowered = "dsh"
		}
		var last float64 = 1
		if uppered != "BTC" {
			l.tickerlock.RLock()
			if uppered == "USD" {
				lastS, ok := l.ticker["USDT_BTC"]
				if !ok {
					l.tickerlock.RUnlock()
					return nil, fmt.Errorf("No ticker found for %s (used USDT)", uppered)
				}
				last = 1 / lastS.Last
			} else if uppered == "IOT" {
				if time.Since(l.BitfinLender.iotLastTime) > time.Minute*30 || l.BitfinLender.iotLast == 0 {
					api := bitfinex.New("", "")
					ti, err := api.Ticker("IOTBTC")
					if err == nil {
						l.BitfinLender.iotLastTime = time.Now()
						l.BitfinLender.iotLast = ti.LastPrice
					}
				}

				last = l.BitfinLender.iotLast
				if last == 0 {
					l.tickerlock.RUnlock()
					return nil, fmt.Errorf("No ticker found for %s", uppered)
				}
			} else if uppered == "EOS" {
				if time.Since(l.BitfinLender.eosLastTime) > time.Minute*30 || l.BitfinLender.eosLast == 0 {
					api := bitfinex.New("", "")
					ti, err := api.Ticker("EOSBTC")
					if err == nil {
						l.BitfinLender.eosLastTime = time.Now()
						l.BitfinLender.eosLast = ti.LastPrice
					}
				}

				last = l.BitfinLender.eosLast
				if last == 0 {
					l.tickerlock.RUnlock()
					return nil, fmt.Errorf("No ticker found for %s", uppered)
				}
			} else {
				lastS, ok := l.ticker[fmt.Sprintf("BTC_%s", uppered)]
				if !ok {
					l.tickerlock.RUnlock()
					return nil, fmt.Errorf("No ticker found for %s", uppered)
				}
				last = lastS.Last
			}
			l.tickerlock.RUnlock()
		}
		cstat := userdb.NewUserStatistic(v, last)
		stats.Currencies[v] = cstat
	}

	// Available
	for _, v := range balancer.Currencies[balancer.BitfinexExchange] {
		lowered := strings.ToLower(v)
		if lowered == "dash" {
			lowered = "dsh"
		}

		curbal := bals[bitfinex.WalletKey{"deposit", lowered}]
		if !math.IsNaN(curbal.Available) {
			stats.Currencies[correctCurencyString(v)].AvailableBalance = curbal.Available
		}
	}

	// Active
	activeLentCount := make(map[string]float64)

	first := true
	for _, v := range balancer.Currencies[balancer.BitfinexExchange] {
		cur := correctCurencyString(v)
		for _, loan := range activeLoan[cur] {
			loan.Rate = loan.Rate / 100 / 365

			stats.Currencies[cur].ActiveLentBalance += loan.Amount
			stats.Currencies[cur].AverageActiveRate += loan.Rate
			activeLentCount[loan.Currency] += 1
			if first && loan.Rate != 0 {
				stats.Currencies[cur].HighestRate = loan.Rate
				stats.Currencies[cur].LowestRate = loan.Rate
				first = false
			} else {
				if loan.Rate > stats.Currencies[cur].HighestRate && loan.Rate != 0 {
					stats.Currencies[cur].HighestRate = loan.Rate
				}
				if loan.Rate < stats.Currencies[cur].LowestRate && loan.Rate != 0 {
					stats.Currencies[cur].LowestRate = loan.Rate
				}
			}
			stats.TotalCurrencyMap[cur] += l.getBTCAmount(loan.Amount, cur)
		}

		for _, loan := range inact[cur] {
			loan.Rate = loan.Rate / 365 / 100
			stats.Currencies[cur].ActiveLentBalance += loan.ExecutedAmount
			stats.Currencies[cur].AverageActiveRate += loan.Rate
			activeLentCount[cur] += 1

			stats.TotalCurrencyMap[cur] += l.getBTCAmount(loan.ExecutedAmount, cur)
		}
	}

	// Finish Active Averages
	for k := range stats.Currencies {
		stats.Currencies[k].AverageActiveRate = stats.Currencies[k].AverageActiveRate / activeLentCount[k]
	}

	// On Order
	inactiveLentCount := make(map[string]float64)
	for _, v := range balancer.Currencies[balancer.BitfinexExchange] {
		cur := correctCurencyString(v)
		for _, loan := range inact[cur] {
			loan.Rate = loan.Rate / 365 / 100
			stats.Currencies[cur].OnOrderBalance += loan.RemainingAmount
			stats.Currencies[cur].AverageOnOrderRate += loan.Rate
			inactiveLentCount[cur] += 1

			stats.TotalCurrencyMap[cur] += l.getBTCAmount(loan.RemainingAmount, cur)
		}
	}

	for k := range stats.Currencies {
		stats.Currencies[k].AverageOnOrderRate = stats.Currencies[k].AverageOnOrderRate / inactiveLentCount[k]
	}

	// Check if to save
	l.recordMapLock.Lock()
	v, ok := l.recordMap[balancer.BitfinexExchange][username]
	l.recordMapLock.Unlock()
	if ok {
		if time.Since(v) < time.Minute*10 {
			return stats, nil
		}
	}
	// Save here
	// TODO: Jesse Save the stats here. This is the userstatistics, we will retrieve these by time
	// db.RecordData(stats)
	stats.Exchange = userdb.BitfinexExchange
	err := l.Bee.SaveUserStastics(stats)

	l.recordMapLock.Lock()
	l.recordMap[balancer.BitfinexExchange][username] = time.Now()
	l.recordMapLock.Unlock()
	return stats, err
}

func getLendingSymbol(sym string) string {
	return fmt.Sprintf("f%s%s", sym)
}

func getTradeSymbol(sym string, pair string) string {
	return fmt.Sprintf("t%s%s", sym, pair)
}

func (l *BitfinexLender) TickerLoop() {
	ti := time.NewTicker(time.Minute)
	for _ = range ti.C {
		select {
		case <-l.quit:
			l.quit <- true
			return
		default:
		}
		err := l.GetTickers()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (l *BitfinexLender) GetTickers() error {
	err := l.take()
	if err != nil {
		return err
	}
	tt, ft, err := l.API.AllLendingTickers()
	if err != nil {
		return err
	}

	l.tickerlock.Lock()
	for _, t := range tt {
		l.Ticker[t.Symbol] = t
	}
	for _, t := range ft {
		l.FundingTicker[t.Symbol] = t
	}

	l.tickerlock.Unlock()
	return nil
}
