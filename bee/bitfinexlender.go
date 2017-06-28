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
)

type BitfinexLender struct {
	tickerlock    sync.RWMutex
	Ticker        map[string]bitfinex.V2Ticker
	FundingTicker map[string]bitfinex.V2FundingTicker

	usersDoneLock sync.RWMutex
	usersDone     map[string]time.Time

	API         *bitfinex.API
	rateLimiter *rate.RateLimiter

	nextStart time.Time

	quit chan bool
}

func NewBitfinexLender() *BitfinexLender {
	b := new(BitfinexLender)
	b.API = bitfinex.New("Public", "Calls")
	b.Ticker = make(map[string]bitfinex.V2Ticker)
	b.FundingTicker = make(map[string]bitfinex.V2FundingTicker)
	b.GetTickers()
	b.rateLimiter = rate.New(90, time.Minute)

	b.usersDone = make(map[string]time.Time)
	b.quit = make(chan bool)
	return b
}

func (bl *BitfinexLender) Close() {
	bl.quit <- true
}

func (bl *BitfinexLender) take() error {
	ok, remain := bl.rateLimiter.Try()
	if ok {
		return nil
	}
	if remain < time.Second*2 {
		time.Sleep(remain)
		return nil
	}
	bl.nextStart = time.Now().Add(remain)
	return fmt.Errorf("Don't spam Bitfinex. Have to sleep %s before calling again", remain.Seconds())
}

func (l *Lender) ProcessBitfinexUser(u *LendUser) error {
	bl := l.BitfinLender
	// Have to wait before making another call
	if time.Now().Before(bl.nextStart) {
		return nil
	}

	api := bitfinex.New(u.U.AccessKey, u.U.SecretKey)

	// api.Ticker(symbol)
	err := bl.take()
	if err != nil {
		return err
	}
	bals, err := api.WalletBalances()
	if err != nil {
		return err
	}

	// Inactive
	err = bl.take()
	if err != nil {
		return err
	}
	inactMap := make(map[string]bitfinex.Offers)
	inactiveOffers, err := api.ActiveOffers()
	if err != nil {
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
		return err
	}
	activeMap := make(map[string]bitfinex.Credits)
	activeOffers, err := api.ActiveCredits()
	if err != nil {
		return err
	}
	for _, o := range activeOffers {
		activeMap[correctCurencyString(o.Currency)] = append(activeMap[correctCurencyString(o.Currency)], o)
	}

	_, err = l.recordBitfinexStatistics(u.U.Username, bals, inactMap, activeMap)
	if err != nil {
		fmt.Println(err)
	}
	for _, c := range u.U.Currency {
		lower := strings.ToLower(c)
		if lower == "dash" {
			lower = "dsh"
		}

		// You got no money buddy
		if bals[bitfinex.WalletKey{"deposit", lower}].Amount == 0 {
			continue
		}

		avail := bals[bitfinex.WalletKey{"deposit", lower}].Available

		err = bl.take()
		if err != nil {
			return err
		}
		o, err := api.NewOffer(lower, avail, 0, 2, "lend")
		if err != nil {
			return err
		}
		fmt.Println("Created loan: ", o)
		var _ = avail
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
	}

	// Finish Active Averages
	for k := range stats.Currencies {
		stats.Currencies[k].AverageActiveRate = stats.Currencies[k].AverageActiveRate / activeLentCount[k] / 365
	}

	// On Order
	inactiveLentCount := make(map[string]float64)
	for _, v := range balancer.Currencies[balancer.BitfinexExchange] {
		cur := correctCurencyString(v)
		for _, loan := range inact[cur] {
			stats.Currencies[cur].OnOrderBalance += loan.RemainingAmount
			stats.Currencies[cur].AverageOnOrderRate += loan.RemainingAmount
			inactiveLentCount[cur] += 1

			stats.TotalCurrencyMap[cur] += l.getBTCAmount(loan.RemainingAmount, cur)
		}
	}

	for k := range stats.Currencies {
		stats.Currencies[k].AverageOnOrderRate = stats.Currencies[k].AverageOnOrderRate / inactiveLentCount[k] / 365
	}

	// Check if to save
	l.recordMapLock.Lock()
	defer l.recordMapLock.Unlock()
	v, ok := l.recordMap[balancer.BitfinexExchange][username]
	if ok {
		if time.Since(v) < time.Minute*10 {
			return stats, nil
		}
	}
	// Save here
	// TODO: Jesse Save the stats here. This is the userstatistics, we will retrieve these by time
	// db.RecordData(stats)

	l.recordMap[balancer.BitfinexExchange][username] = time.Now()
	return stats, nil
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