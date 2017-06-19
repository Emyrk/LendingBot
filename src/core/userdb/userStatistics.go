package userdb

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/database"
	"github.com/tinylib/msgp/msgp"
)

var _ = strings.Compare
var _ = fmt.Println

var curarr = []string{"BTC", "BTS", "CLAM", "DOGE", "DASH", "LTC", "MAID", "STR", "XMR", "XRP", "ETH", "FCT"}

var (
	UserStatisticDBMetaDataBucket = []byte("UserStatisticsDBMeta")
	CurrentDayKey                 = []byte("CurrentDay")
	CurrentIndex                  = []byte("CurrentIndex")
	PoloniexPrefix                = []byte("PoloniexBucket")
	BucketMarkForDelete           = []byte("Mark for delete")
	LendingHistoryPrefix          = []byte("LendingHistory")
)

var (
	Currencies []string = []string{"BTC"}
)

type UserStatisticsDB struct {
	db database.IDatabase

	LastPoloniexRateSave map[string]time.Time
	LastCurrentIndexCalc time.Time
	CurrentDay           int
	CurrentIndex         int // 0 to 30
}

func (u *UserStatisticsDB) Close() error {
	return u.db.Close()
}

func NewUserStatisticsMapDB() (*UserStatisticsDB, error) {
	return newUserStatisticsDB(true)
}

func NewUserStatisticsDB() (*UserStatisticsDB, error) {
	return newUserStatisticsDB(false)
}

func newUserStatisticsDB(mapDB bool) (*UserStatisticsDB, error) {
	u := new(UserStatisticsDB)

	userStatsPath := os.Getenv("USER_STATS_DB")
	if userStatsPath == "" {
		userStatsPath = "UserStats.db"
	}

	if mapDB {
		u.db = database.NewMapDB()
		u.startDB()
	} else {
		var newDB bool
		if _, err := os.Stat(userStatsPath); os.IsNotExist(err) {
			newDB = false
		} else {
			newDB = true
		}

		u.db = database.NewBoltDB(userStatsPath)
		if !newDB {
			u.startDB()
		}
	}

	err := u.loadCurrentIndex()
	if err != nil {
		return u, err
	}

	err = u.CalculateCurrentIndex()
	if err != nil {
		return u, err
	}

	u.LastPoloniexRateSave = make(map[string]time.Time)
	return u, nil
}

type UserStatisticList []AllUserStatistic

func (slice UserStatisticList) Len() int {
	return len(slice)
}

func (slice UserStatisticList) Less(i, j int) bool {
	if !slice[i].Time.Before(slice[j].Time) {
		return true
	}
	return false
}

func (slice UserStatisticList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func NewAllUserStatistic() *AllUserStatistic {
	us := new(AllUserStatistic)
	us.Currencies = make(map[string]*UserStatistic)
	us.TotalCurrencyMap = make(map[string]float64)
	us.Time = time.Now()

	return us
}

func NewUserStatistic(currency string, last float64) *UserStatistic {
	s := new(UserStatistic)
	s.AvailableBalance = 0
	s.ActiveLentBalance = 0
	s.OnOrderBalance = 0
	s.AverageActiveRate = 0
	s.AverageOnOrderRate = 0
	s.Currency = currency
	s.BTCRate = last
	s.Time = time.Now()
	return s
}

func (us *UserStatistic) Scrub() {
	if math.IsNaN(us.AvailableBalance) {
		us.AvailableBalance = 0
	}

	if math.IsNaN(us.ActiveLentBalance) {
		us.ActiveLentBalance = 0
	}

	if math.IsNaN(us.OnOrderBalance) {
		us.OnOrderBalance = 0
	}

	if math.IsNaN(us.AverageActiveRate) {
		us.AverageActiveRate = 0
	}

	if math.IsNaN(us.AverageOnOrderRate) {
		us.AverageOnOrderRate = 0
	}
}

func (a *AllUserStatistic) Scrub() {
	for _, v := range a.Currencies {
		v.Scrub()
	}
}

func (a *AllUserStatistic) IsSameAs(b *AllUserStatistic) bool {
	if len(a.Currencies) != len(b.Currencies) {
		return false
	}

	for k, v := range a.Currencies {
		if v2, ok := b.Currencies[k]; !ok || !v2.IsSameAs(v) {
			return false
		}
	}
	return true
}

func (a *UserStatistic) IsSameAs(b *UserStatistic) bool {
	if a.AvailableBalance != b.AvailableBalance {
		return false
	}
	if a.ActiveLentBalance != b.ActiveLentBalance {
		return false
	}
	if a.OnOrderBalance != b.OnOrderBalance {
		return false
	}
	if a.AverageActiveRate != b.AverageActiveRate {
		return false
	}
	if a.AverageOnOrderRate != b.AverageOnOrderRate {
		return false
	}
	if a.Currency != b.Currency {
		return false
	}

	return true
}

func (us *UserStatisticsDB) startDB() {
	us.db.Put(UserStatisticDBMetaDataBucket, CurrentIndex, primitives.Uint32ToBytes(0))
	us.db.Put(UserStatisticDBMetaDataBucket, CurrentDayKey, primitives.Uint32ToBytes(0))
}

func (us *UserStatisticsDB) loadCurrentIndex() error {
	data, err := us.db.Get(UserStatisticDBMetaDataBucket, CurrentIndex)
	if err != nil {
		return nil
	}

	u, err := primitives.BytesToUint32(data)
	if err != nil {
		return err
	}

	us.CurrentIndex = int(u)
	return nil
}

func (us *UserStatisticsDB) CalculateCurrentIndex() (err error) {
	data, err := us.db.Get(UserStatisticDBMetaDataBucket, CurrentDayKey)
	if err != nil {
		return err
	}

	u, err := primitives.BytesToUint32(data)
	if err != nil {
		return err
	}

	day := GetDay(time.Now())
	us.CurrentDay = day
	oldDay := int(u)
	if day > oldDay {
		err = us.db.Put(UserStatisticDBMetaDataBucket, CurrentDayKey, primitives.Uint32ToBytes(uint32(day)))
		if err != nil {
			return err
		}

		us.CurrentIndex++
		if us.CurrentIndex > 30 {
			us.CurrentIndex = 0
		}

		return us.db.Put(UserStatisticDBMetaDataBucket, CurrentIndex, primitives.Uint32ToBytes(uint32(us.CurrentIndex)))
	}

	return nil
}

func (us *UserStatisticsDB) RecordData(stats *AllUserStatistic) error {
	if time.Since(us.LastCurrentIndexCalc).Hours() > 1 {
		us.CalculateCurrentIndex()
	}
	seconds := GetSeconds(stats.Time)
	stats.day = GetDay(stats.Time)

	stats.Scrub()

	var buf bytes.Buffer
	err := msgp.Encode(&buf, stats)
	// err := stats.EncodeMsg(&buf)
	// data, err := stats.MarshalMsg(b)
	if err != nil {
		return err
	}

	return us.putStats(stats.Username, seconds, buf.Bytes())
}

type PoloniexRateSample struct {
	SecondsPastMidnight int
	Rate                float64
}

type PoloniexRateSamples []PoloniexRateSample

func (slice PoloniexRateSamples) Len() int {
	return len(slice)
}

func (slice PoloniexRateSamples) Less(i, j int) bool {
	if slice[i].SecondsPastMidnight > slice[j].SecondsPastMidnight {
		return true
	}
	return false
}

func (slice PoloniexRateSamples) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// GetPoloniexDataLastXDays returns a 2D array:
//		[x][y]PoloniexRateSample
//			x  = x days past, E.g 0 = today
func (us *UserStatisticsDB) GetPoloniexDataLastXDays(dayRange int, currency string) [][]PoloniexRateSample {
	historyStats := make([][]PoloniexRateSample, dayRange)
	start := GetDay(time.Now())
	for i := 0; i < dayRange; i++ {
		day := start - i
		bucket := primitives.Uint32ToBytes(uint32(day))
		bucket = append(PoloniexPrefix, bucket...)
		bucket = append(getCurrencyPre(currency), bucket...)
		datas, keys, err := us.db.GetAll(bucket)
		if err != nil {
			continue
		}
		stats := make([]PoloniexRateSample, 0)
		for i, data := range datas {
			rate, err := primitives.BytesToFloat64(data)
			if err != nil {
				continue
			}

			secondsPast, err := primitives.BytesToUint32(keys[i])
			if err != nil {
				continue
			}

			stats = append(stats, PoloniexRateSample{SecondsPastMidnight: int(secondsPast), Rate: rate})
		}

		sort.Sort(PoloniexRateSamples(stats))
		historyStats[i] = stats
	}
	return historyStats
}

type PoloniexStats struct {
	FiveMinAvg float64 `json:"fiveminavg"`
	HrAvg      float64 `json:"hravg"`
	DayAvg     float64 `json:"dayavg"`
	WeekAvg    float64 `json:"weekavg"`
	MonthAvg   float64 `json:"monthavg"`

	HrStd    float64 `json:"hrstd"`
	DayStd   float64 `json:"daystd"`
	WeekStd  float64 `json:"weekstd"`
	MonthStd float64 `json:"monthstd"`
}

func (p *PoloniexStats) String() string {
	return fmt.Sprintf("HrAvg: %f, DayAvg: %f, WeekAvg: %f, MonthAvg: %f, HrStd: %f, DayStd: %f, WeekStd: %f, MonthStd: %f\n",
		p.HrAvg, p.DayAvg, p.WeekAvg, p.MonthAvg, p.HrStd, p.DayStd, p.WeekStd, p.MonthStd)
}

func (us *UserStatisticsDB) GetPoloniexStatistics(currency string) *PoloniexStats {
	poloStats := new(PoloniexStats)

	poloDatStats := us.GetPoloniexDataLastXDays(30, currency)
	// No data
	if len(poloDatStats[0]) == 0 {
		return nil
	}

	sec := GetSeconds(time.Now())
	var lastHr []PoloniexRateSample
	var fivemin []PoloniexRateSample

	in := 0
	not := 0
	for _, v := range poloDatStats[0] {
		if v.SecondsPastMidnight > sec-3600 {
			in++
			lastHr = append(lastHr, v)
			if v.SecondsPastMidnight > sec-300 {
				fivemin = append(fivemin, v)
			}
		} else {
			not++
		}
	}

	var all []PoloniexRateSample
	dayCutoff := 0
	weekCutoff := 0
	count := 0
	for i, v := range poloDatStats {
		all = append(all, v...)
		count += len(v)
		if i == 1 {
			dayCutoff = count
		} else if i == 7 {
			weekCutoff = count
		}
	}

	poloStats.FiveMinAvg, _ = GetAvgAndStd(fivemin)
	poloStats.HrAvg, poloStats.HrStd = GetAvgAndStd(lastHr)
	poloStats.DayAvg, poloStats.DayStd = GetAvgAndStd(all[:dayCutoff])
	poloStats.WeekAvg, poloStats.WeekStd = GetAvgAndStd(all[:weekCutoff])
	poloStats.MonthAvg, poloStats.MonthStd = GetAvgAndStd(all)
	return poloStats
}

func GetAvgAndStd(data []PoloniexRateSample) (avg float64, std float64) {
	total := float64(0)
	count := float64(0)

	for _, v := range data {
		total += v.Rate
		count++
	}
	avg = total / count

	// Standard Deviation
	sum := float64(0)
	for _, v := range data {
		sum += (v.Rate - avg) * (v.Rate - avg)
	}
	std = math.Sqrt(sum / (count - 1))
	return
}

func (us *UserStatisticsDB) RecordPoloniexStatistic(currency string, rate float64) error {
	return us.RecordPoloniexStatisticTime(currency, rate, time.Now())
}

func getCurrencyPre(currency string) []byte {
	var pre []byte
	if currency == "BTC" {
		currency = ""
	} else {
		pre = []byte(currency)
	}
	return pre
}

func (us *UserStatisticsDB) RecordPoloniexStatisticTime(currency string, rate float64, t time.Time) error {
	if t, ok := us.LastPoloniexRateSave[currency]; !ok {
		us.LastPoloniexRateSave[currency] = time.Now()
	} else if time.Since(t).Seconds() < 10 {
		return nil
	}

	day := GetDayBytes(GetDay(t))
	sec := GetSeconds(t)
	buck := append(PoloniexPrefix, day...)
	buck = append(getCurrencyPre(currency), buck...)

	secBytes := primitives.Uint32ToBytes(uint32(sec))
	data, err := primitives.Float64ToBytes(rate)
	if err != nil {
		return err
	}

	us.LastPoloniexRateSave[currency] = time.Now()
	return us.db.Put(buck, secBytes, data)
}

func (us *UserStatisticsDB) GetStatistics(username string, dayRange int) ([][]AllUserStatistic, error) {
	if dayRange > 30 {
		return nil, fmt.Errorf("Day range must be less than 30")
	}

	stats := make([][]AllUserStatistic, dayRange)
	for i := 0; i < dayRange; i++ {
		buc := us.getBucketPlusX(username, i*-1)
		statlist := us.getStatsFromBucket(buc)
		stats[i] = statlist
	}

	return stats, nil
}

type DayAvg struct {
	LoanRate       float64
	Lent           float64
	NotLent        float64
	LendingPercent float64

	AvgBTCValue float64
}

func (da *DayAvg) String() string {
	return fmt.Sprintf("LoanRate: %f, BTCLent: %f, BTCNotLent: %f, LendingPercent: %f",
		da.LoanRate, da.Lent, da.NotLent, da.LendingPercent)
}

func GetCombinedDayAverage(dayStats []AllUserStatistic) *DayAvg {
	da := new(DayAvg)
	da.LoanRate = float64(0)
	da.Lent = float64(0)
	da.NotLent = float64(0)
	da.LendingPercent = float64(0)

	all := make(map[string][]UserStatistic)
	for _, v := range dayStats {
		for _, currency := range curarr {
			tmp := v.Currencies[currency]
			if tmp != nil {
				all[currency] = append(all[currency], *tmp)
			}
		}
	}

	count := float64(0)
	for _, currency := range curarr {
		sD := GetDayAvg(all[currency])
		if sD != nil {
			total := sD.AvgBTCValue * (sD.NotLent + sD.Lent)
			da.LoanRate += sD.LoanRate * (sD.Lent * sD.AvgBTCValue)
			da.Lent += sD.Lent * total
			da.NotLent += sD.NotLent * total
			da.LendingPercent += sD.LendingPercent * total
			count += total
		}
	}

	da.LoanRate = da.LoanRate / count
	da.Lent = da.Lent / count
	da.NotLent = da.NotLent / count
	da.LendingPercent = da.LendingPercent / count
	return da
}

func GetDayAvg(dayStats []UserStatistic) *DayAvg {
	da := new(DayAvg)
	da.LoanRate = float64(0)
	da.Lent = float64(0)
	da.NotLent = float64(0)
	da.LendingPercent = float64(0)

	if len(dayStats) == 0 {
		return nil
	}
	var diff float64
	last := dayStats[0].Time
	totalSeconds := float64(0)
	if len(dayStats) == 1 {
		last = last.Add(-1 * time.Second)
	}

	for _, s := range dayStats {
		diff = timeDiff(last, s.Time)
		// Override any use of time. It'll be off by a bit but ¯\_(ツ)_/¯
		diff = 1
		da.LoanRate += diff * s.AverageActiveRate
		da.Lent += diff * s.ActiveLentBalance
		da.NotLent += diff * (s.AvailableBalance + s.OnOrderBalance)
		v := diff * (s.ActiveLentBalance / (s.AvailableBalance + s.OnOrderBalance + s.ActiveLentBalance))
		if math.IsNaN(v) {
			v = 0
		}
		da.LendingPercent += v
		da.AvgBTCValue += diff * s.BTCRate
		totalSeconds += diff
	}

	da.LoanRate = da.LoanRate / totalSeconds
	da.Lent = da.Lent / totalSeconds
	da.NotLent = da.NotLent / totalSeconds
	da.LendingPercent = da.LendingPercent / totalSeconds
	da.AvgBTCValue = da.AvgBTCValue / totalSeconds

	return da
}

func timeDiff(a time.Time, b time.Time) float64 {
	d := a.Sub(b).Seconds()
	if d < 0 {
		return d * -1
	}
	return d
}

func (us *UserStatisticsDB) getStatsFromBucket(bucket []byte) []AllUserStatistic {
	var resp []AllUserStatistic
	values, _, err := us.db.GetAll(bucket)
	if err != nil {
		return resp
	}

	for _, data := range values {
		var tmp AllUserStatistic
		_, err := tmp.UnmarshalMsg(data)
		// err := tmp.UnmarshalBinary(data)
		if err != nil {
			// Try to unmarshal old
			var old OldUserStatistic
			err := old.UnmarshalBinary(data)
			if err != nil {
				continue
			} else {
				var n UserStatistic
				tmp.Time = old.Time
				tmp.TotalCurrencyMap = old.TotalCurrencyMap
				n.AvailableBalance = old.AvailableBalance
				n.ActiveLentBalance = old.ActiveLentBalance
				n.OnOrderBalance = old.OnOrderBalance
				n.AverageActiveRate = old.AverageActiveRate
				n.AverageOnOrderRate = old.AverageOnOrderRate
				n.Time = old.Time
				n.BTCRate = 1
				tmp.Currencies = make(map[string]*UserStatistic)
				tmp.Currencies["BTC"] = &n
				tmp.Time = n.Time
			}
		}
		resp = append(resp, tmp)
	}

	sort.Sort(UserStatisticList(resp))

	return resp
}

func (us *UserStatisticsDB) putStats(username string, seconds int, data []byte) error {
	buc := us.getBucket(username)
	key := primitives.Uint32ToBytes(uint32(seconds))
	if data, _ := us.db.Get(buc, BucketMarkForDelete); len(data) > 0 && data[0] == 0xFF {
		us.db.Clear(buc)
	}

	// TODO: Make the mark apply less often
	us.db.Put(us.getNextBucket(username), BucketMarkForDelete, []byte{0xFF})

	return us.db.Put(buc, key, data)
}

func (us *UserStatisticsDB) WipeUser(username string) error {
	for i := 0; i < 30; i++ {
		hash := GetUsernameHash(username)
		index := primitives.Uint32ToBytes(uint32(i))
		buc := append(hash[:], index...)
		us.db.Clear(buc)
	}
	return nil
}

func (us *UserStatisticsDB) getBucket(username string) []byte {
	hash := GetUsernameHash(username)
	index := primitives.Uint32ToBytes(uint32(us.CurrentIndex))
	return append(hash[:], index...)
}

func (us *UserStatisticsDB) getNextBucket(username string) []byte {
	i := us.CurrentIndex + 1
	if i > 30 {
		i = 0
	}

	hash := GetUsernameHash(username)
	index := primitives.Uint32ToBytes(uint32(i))
	return append(hash[:], index...)
}

func (us *UserStatisticsDB) getBucketPlusX(username string, offset int) []byte {
	i := us.CurrentIndex + offset

	if i > 30 {
		overFlow := i - 30
		i = -1 + overFlow
	}

	if i < 0 {
		underFlow := i * -1
		i = 31 - underFlow
	}

	hash := GetUsernameHash(username)
	index := primitives.Uint32ToBytes(uint32(i))

	return append(hash[:], index...)
}

func (us *UserStatisticsDB) Fix() {
	bucks, _ := us.db.ListAllBuckets()

	Keys := make([][][]byte, 2)
	Buckets := make([][]byte, 2)
	Data := make([][][]byte, 2)

	i := 0
	for _, b := range bucks {
		if strings.Contains(string(b), "PoloniexBucket") {

			Data[i], Keys[i], _ = us.db.GetAll(b)
			Buckets[i] = b
			i++
		}
	}

	zero, _ := primitives.BytesToUint32(Buckets[0][:len(Buckets[0])-4])
	one, _ := primitives.BytesToUint32(Buckets[1][:len(Buckets[1])-4])
	day := GetDayBytes(GetDay(time.Now()))
	if zero > one {
		for i, k := range Keys[0] {
			us.db.Put(append(PoloniexPrefix, day...), k, Data[0][i])
		}
		day := GetDay(time.Now()) - 1
		for i, k := range Keys[1] {
			us.db.Put(append(PoloniexPrefix, primitives.Uint32ToBytes(uint32(day))...), k, Data[1][i])
		}
	} else {
		for i, k := range Keys[1] {
			us.db.Put(append(PoloniexPrefix, day...), k, Data[1][i])
		}
		day := GetDayBytes(GetDay(time.Now()) - 1)
		for i, k := range Keys[0] {
			us.db.Put(append(PoloniexPrefix, day...), k, Data[0][i])
		}
	}
}

func GetDay(t time.Time) int {
	return int(t.Unix()) / 86400
}

func GetDayBytes(day int) []byte {
	return primitives.Uint32ToBytes(uint32(day))
}

func NewAllLendingHistoryEntry() *AllLendingHistoryEntry {
	l := new(AllLendingHistoryEntry)
	l.Data = make(map[string]*LendingHistoryEntry)
	return l
}

func (a *AllLendingHistoryEntry) Pop() {
	for _, v := range AvaiableCoins {
		if _, ok := a.Data[v]; !ok {
			a.Data[v] = new(LendingHistoryEntry)
		}
	}
}

// LendingHistory
func (us *UserStatisticsDB) SaveLendingHistory(lendHist *AllLendingHistoryEntry) error {
	ld := GetDay(lendHist.Time)
	buc := getLHBucket(lendHist.Username)
	key := append([]byte("Polo"), primitives.Uint32ToBytes(uint32(ld))...)

	var buf bytes.Buffer
	err := msgp.Encode(&buf, lendHist)
	if err != nil {
		return err
	}

	return us.db.Put(buc, key, buf.Bytes())
}

func (us *UserStatisticsDB) GetLendHistorySummary(username string, t time.Time) (*AllLendingHistoryEntry, error) {
	ld := GetDay(t)
	buc := getLHBucket(username)
	key := append([]byte("Polo"), primitives.Uint32ToBytes(uint32(ld))...)
	tmp := NewAllLendingHistoryEntry()
	v, err := us.db.Get(buc, key)
	if v != nil && err == nil {
		_, err := tmp.UnmarshalMsg(v)
		if err != nil {
			return tmp, err
		}
		return tmp, nil
	}
	return tmp, fmt.Errorf("Not found or an error: %v", err)
}

func getLHBucket(username string) []byte {
	h := GetUsernameHash(username)
	return append(LendingHistoryPrefix, h[:]...)
}

func GetSeconds(t time.Time) int {
	// 3 units of time, in seconds from midnight
	sec := t.Second()
	min := t.Minute() * 60
	hour := t.Hour() * 3600

	return sec + min + hour

}

func abs(a float64) float64 {
	if a < 0 {
		return a * -1
	}
	return a
}
