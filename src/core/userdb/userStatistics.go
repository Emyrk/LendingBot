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
)

var _ = strings.Compare
var _ = fmt.Println

var (
	UserStatisticDBMetaDataBucket = []byte("UserStatisticsDBMeta")
	CurrentDayKey                 = []byte("CurrentDay")
	CurrentIndex                  = []byte("CurrentIndex")
	PoloniexPrefix                = []byte("PoloniexBucket")
	BucketMarkForDelete           = []byte("Mark for delete")
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

type UserStatisticList []UserStatistic

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

type UserStatistic struct {
	Username           string    `json:"username"`
	AvailableBalance   float64   `json:"availbal"`
	ActiveLentBalance  float64   `json:"availlent"`
	OnOrderBalance     float64   `json:"onorder"`
	AverageActiveRate  float64   `json:"activerate"`
	AverageOnOrderRate float64   `json:"onorderrate"`
	Time               time.Time `json:"time"`
	Currency           string    `json:"currency"`

	TotalCurrencyMap map[string]float64

	day int
}

func NewUserStatistic() *UserStatistic {
	us := new(UserStatistic)
	us.Username = ""
	us.AvailableBalance = 0
	us.ActiveLentBalance = 0
	us.OnOrderBalance = 0
	us.AverageActiveRate = 0
	us.AverageOnOrderRate = 0
	us.Currency = ""

	us.TotalCurrencyMap = make(map[string]float64)

	return us
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

func (a *UserStatistic) IsSameAs(b *UserStatistic) bool {
	if a.Username != b.Username {
		return false
	}
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

	if len(a.TotalCurrencyMap) != len(b.TotalCurrencyMap) {
		return false
	}

	for k, v := range a.TotalCurrencyMap {
		if b.TotalCurrencyMap[k] != v {
			return false
		}
	}

	return true
}

func (s *UserStatistic) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	b, err := primitives.MarshalStringToBytes(s.Username, UsernameMaxLength)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.AvailableBalance)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.ActiveLentBalance)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.OnOrderBalance)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.AverageActiveRate)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.AverageOnOrderRate)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b = primitives.Uint32ToBytes(uint32(s.day))
	buf.Write(b)

	b, err = s.Time.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.MarshalStringToBytes(s.Currency, 5)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	l := len(s.TotalCurrencyMap)
	buf.Write(primitives.Uint32ToBytes(uint32(l)))

	for k, v := range s.TotalCurrencyMap {
		data, err := primitives.MarshalStringToBytes(k, 5)
		if err != nil {
			return nil, err
		}
		buf.Write(data)

		data, err = primitives.Float64ToBytes(v)
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.Next(buf.Len()), nil
}

func (s *UserStatistic) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
	return err
}

func (s *UserStatistic) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[UserStatistic] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data

	s.Username, newData, err = primitives.UnmarshalStringFromBytesData(newData, UsernameMaxLength)
	if err != nil {
		return nil, err
	}

	s.AvailableBalance, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.ActiveLentBalance, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.OnOrderBalance, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.AverageActiveRate, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.AverageOnOrderRate, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	var u uint32
	u, err = primitives.BytesToUint32(newData)
	if err != nil {
		return nil, err
	}
	s.day = int(u)
	newData = newData[4:]

	td := newData[:15]
	newData = newData[15:]
	err = s.Time.UnmarshalBinary(td)
	if err != nil {
		return nil, err
	}

	s.Currency, newData, err = primitives.UnmarshalStringFromBytesData(newData, 5)
	if err != nil {
		return nil, err
	}

	l, err := primitives.BytesToUint32(newData[:4])
	if err != nil {
		return nil, err
	}
	newData = newData[4:]

	s.TotalCurrencyMap = make(map[string]float64)
	for i := 0; i < int(l); i++ {
		var key string
		key, newData, err = primitives.UnmarshalStringFromBytesData(newData, 5)
		if err != nil {
			return nil, err
		}

		var v float64
		v, newData, err = primitives.BytesToFloat64Data(newData)
		if err != nil {
			return nil, err
		}
		s.TotalCurrencyMap[key] = v
	}
	s.Scrub()

	return
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

func (us *UserStatisticsDB) RecordData(stats *UserStatistic) error {
	if time.Since(us.LastCurrentIndexCalc).Hours() > 1 {
		us.CalculateCurrentIndex()
	}
	seconds := GetSeconds(stats.Time)
	stats.day = GetDay(stats.Time)

	stats.Scrub()

	data, err := stats.MarshalBinary()
	if err != nil {
		return err
	}

	return us.putStats(stats.Username, seconds, data)
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
	HrAvg    float64 `json:"hravg"`
	DayAvg   float64 `json:"dayavg"`
	WeekAvg  float64 `json:"weekavg"`
	MonthAvg float64 `json:"monthavg"`

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

	in := 0
	not := 0
	for _, v := range poloDatStats[0] {
		if v.SecondsPastMidnight > sec-3600 {
			in++
			lastHr = append(lastHr, v)
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

	fmt.Println("RECORD", currency)
	us.LastPoloniexRateSave[currency] = time.Now()
	return us.db.Put(buck, secBytes, data)
}

func (us *UserStatisticsDB) GetStatistics(username string, dayRange int) ([][]UserStatistic, error) {
	if dayRange > 30 {
		return nil, fmt.Errorf("Day range must be less than 30")
	}

	stats := make([][]UserStatistic, dayRange)
	for i := 0; i < dayRange; i++ {
		buc := us.getBucketPlusX(username, i*-1)
		statlist := us.getStatsFromBucket(buc)
		stats[i] = statlist
	}

	return stats, nil
}

type DayAvg struct {
	LoanRate       float64
	BTCLent        float64
	BTCNotLent     float64
	LendingPercent float64
}

func (da *DayAvg) String() string {
	return fmt.Sprintf("LoanRate: %f, BTCLent: %f, BTCNotLent: %f, LendingPercent: %f",
		da.LoanRate, da.BTCLent, da.BTCNotLent, da.LendingPercent)
}

func GetDayAvg(dayStats []UserStatistic) *DayAvg {
	da := new(DayAvg)
	da.LoanRate = float64(0)
	da.BTCLent = float64(0)
	da.BTCNotLent = float64(0)
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
		da.LoanRate += diff * s.AverageActiveRate
		da.BTCLent += diff * s.ActiveLentBalance
		da.BTCNotLent += diff * (s.AvailableBalance + s.OnOrderBalance)
		da.LendingPercent += diff * (s.ActiveLentBalance / (s.AvailableBalance + s.OnOrderBalance + s.ActiveLentBalance))
		totalSeconds += diff
	}

	da.LoanRate = da.LoanRate / totalSeconds
	da.BTCLent = da.BTCLent / totalSeconds
	da.BTCNotLent = da.BTCNotLent / totalSeconds
	da.LendingPercent = da.LendingPercent / totalSeconds

	return da
}

func timeDiff(a time.Time, b time.Time) float64 {
	d := a.Sub(b).Seconds()
	if d < 0 {
		return d * -1
	}
	return d
}

func (us *UserStatisticsDB) getStatsFromBucket(bucket []byte) []UserStatistic {
	var resp []UserStatistic
	values, _, err := us.db.GetAll(bucket)
	if err != nil {
		return resp
	}

	for _, data := range values {
		var tmp UserStatistic
		err := tmp.UnmarshalBinary(data)
		if err != nil {
			continue
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
