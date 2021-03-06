package userdb

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/slack"
	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/database"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	// "github.com/revel/revel"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	db  database.IDatabase
	mdb *mongo.MongoDB

	LastPoloniexRateSave map[string]time.Time
	LastCurrentIndexCalc time.Time
	CurrentDay           int
	CurrentIndex         int // 0 to 30

	// Cache
	cachelock       sync.RWMutex
	cachesPoloStats map[string]*PoloniexStats
	lastHourUpdate  map[string]time.Time
	lastDayUpdate   map[string]time.Time
	lastWeekUpdate  map[string]time.Time
	lastMonthUpdate map[string]time.Time
}

func (u *UserStatisticsDB) Close() error {
	if u.mdb == nil {
		return u.db.Close()
	}
	return nil
}

func NewUserStatisticsMapDB() (*UserStatisticsDB, error) {
	return newUserStatisticsDB("map")
}

func NewUserStatisticsDB() (*UserStatisticsDB, error) {
	return newUserStatisticsDB("bolt")
}

func NewUserStatisticsMongoDB(uri string, dbu string, dbp string) (*UserStatisticsDB, error) {
	db, err := mongo.CreateStatDB(uri, dbu, dbp)
	if err != nil {
		return nil, fmt.Errorf("Error creating user_stat db: %s\n", err.Error())
	}
	return NewUserStatisticsMongoDBGiven(db)
}

func NewUserStatisticsMongoDBGiven(mdb *mongo.MongoDB) (*UserStatisticsDB, error) {
	u, err := newUserStatisticsDB("mongo")
	if err != nil {
		return nil, err
	}
	u.mdb = mdb
	return u, nil
}

func makeTimeMap() map[string]time.Time {
	m := make(map[string]time.Time)
	for _, c := range AvaiableCoins {
		m[c] = time.Time{}
	}
	return m
}

func newUserStatisticsDB(dbType string) (*UserStatisticsDB, error) {
	u := new(UserStatisticsDB)
	u.cachesPoloStats = make(map[string]*PoloniexStats)
	u.lastHourUpdate = makeTimeMap()
	u.lastDayUpdate = makeTimeMap()
	u.lastWeekUpdate = makeTimeMap()
	u.lastMonthUpdate = makeTimeMap()
	for _, c := range AvaiableCoins {
		u.cachesPoloStats[c] = new(PoloniexStats)
	}

	userStatsPath := os.Getenv("USER_STATS_DB")
	if userStatsPath == "" {
		userStatsPath = "UserStats.db"
	}

	if dbType != "mongo" {
		if dbType != "bolt" {
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
	}

	u.GetPoloniexStatistics("BTC")
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
	if us.mdb == nil {
		us.db.Put(UserStatisticDBMetaDataBucket, CurrentIndex, primitives.Uint32ToBytes(0))
		us.db.Put(UserStatisticDBMetaDataBucket, CurrentDayKey, primitives.Uint32ToBytes(0))
	}
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
	seconds := GetSeconds(stats.Time)
	stats.day = GetDay(stats.Time)

	stats.Scrub()

	if us.mdb != nil {
		s, c, err := us.mdb.GetCollection(mongo.C_UserStat)
		if err != nil {
			return fmt.Errorf("Mongo: RecordData: createSession: %s", err)
		}
		defer s.Close()

		// var eA []AllUserStatistic
		// mus := MongoAllUserStatistics{
		// 	stats.Username,
		// 	eA,
		// }
		// change := mgo.Change{
		// 	Update:    bson.M{"$setOnInsert": mus},
		// 	ReturnNew: false,
		// 	Upsert:    true,
		// }
		// //CAN OPTIMIZE LATER
		// _, err = c.Find(bson.M{"_id": stats.Username}).Apply(change, nil)
		// if err != nil {
		// 	return fmt.Errorf("Mongo: RecordData: create: %s", err)
		// }

		// // db.collection.findAndModify({
		// //   query: { _id: "some potentially existing id" },
		// //   update: {
		// //     $setOnInsert: { foo: "bar" }
		// //   },
		// //   new: true,   // return new doc if one is upserted
		// //   upsert: true // insert the document if it does not exist
		// // })

		// eA = append(eA, *stats)
		// updateAction := bson.M{
		// 	"$push": bson.M{
		// 		"userstats": bson.M{
		// 			"$each": eA,
		// 		},
		// 	},
		// }
		// err = c.UpdateId(stats.Username, updateAction)
		// if err != nil {
		// 	return fmt.Errorf("Mongo: RecordData: insert: %s", err)
		// }
		// return nil

		// key := fmt.Sprintf("%s.ISO(%s)", stats.Username, stats.Time.Format(time.RFC3339))
		// keyString := fmt.Sprintf("ISO(%s)", stats.Time.Format(time.RFC3339))

		upsertKey := bson.M{
			"$and": []bson.M{
				bson.M{"email": stats.Username},
				bson.M{"time": stats.Time},
			},
		}
		upsertAction := bson.M{"$set": stats}

		// fmt.Printf("KEY: %s\n", keyString)

		_, err = c.Upsert(upsertKey, upsertAction)
		if err != nil {
			return fmt.Errorf("Mongo: RecordData: upsert: %s", err)
		}
		return nil
	}

	if time.Since(us.LastCurrentIndexCalc).Hours() > 1 {
		us.CalculateCurrentIndex()
	}
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
		fmt.Printf("Found %d data entries for day %d\n", len(datas), i)
		for i, data := range datas {
			rate, err := primitives.BytesToFloat64(data)
			if err != nil {
				fmt.Println("Could not parse rate")
				continue
			}

			secondsPast, err := primitives.BytesToUint32(keys[i])
			if err != nil {
				fmt.Println("Could not parse seconds")
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

//Used for migration from embedded to mongo
func (us *UserStatisticsDB) GetAllPoloniexStatistics(currency string) (*[]PoloniexStat, error) {
	var poloniexStatsArr []PoloniexStat

	poloDatStats := us.GetPoloniexDataLastXDays(30, currency)

	// No data
	if len(poloDatStats[0]) == 0 {
		return &poloniexStatsArr, nil
	}

	n := time.Now()
	n = time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)

	for _, v := range poloDatStats {
		for _, s := range v {
			poloniexStatsArr = append(poloniexStatsArr, PoloniexStat{Currency: currency, Rate: s.Rate, Time: n.Add(time.Duration(s.SecondsPastMidnight) * time.Second)})
		}
		n = n.Add(-24 * time.Hour)
	}

	return &poloniexStatsArr, nil
}

// GetQuickPoloniexStatistics only uses cache
func (us *UserStatisticsDB) GetQuickPoloniexStatistics(currency string) *PoloniexStats {
	us.cachelock.RLock()
	v := us.cachesPoloStats[currency]
	us.cachelock.RUnlock()
	return v
}

func (us *UserStatisticsDB) GetPoloniexStatistics(currency string) (*PoloniexStats, error) {
	var poloniexStatsArr []PoloniexStat

	if us.mdb != nil {
		s, c, err := us.mdb.GetCollection(mongo.C_Exchange_POL)
		if err != nil {
			return nil, fmt.Errorf("Mongo: GetPoloniexStatistics: getcol: %s", err)
		}
		defer s.Close()

		t := time.Now().Add(-5 * time.Minute)
		update := 5
		if time.Since(us.lastMonthUpdate[currency]) > time.Hour*48 {
			// Need to update month -- Grab all
			t = time.Now().Add(-24 * time.Hour * 30)
			update = 1
		} else if time.Since(us.lastWeekUpdate[currency]) > time.Hour*24 {
			// Need to update week -- Grab week
			t = time.Now().Add(-24 * time.Hour * 7)
			update = 2
		} else if time.Since(us.lastDayUpdate[currency]) > time.Hour {
			// Need to update week -- Grab Day
			t = time.Now().Add(-24 * time.Hour)
			update = 3
		} else if time.Since(us.lastHourUpdate[currency]) > time.Minute*5 {
			// Need to update week -- Grab Hour
			t = time.Now().Add(-1 * time.Hour)
			update = 4
		}
		find := bson.M{
			"$and": []bson.M{
				bson.M{"currency": currency},
				bson.M{"_id": bson.M{"$gt": t}},
			},
		}

		err = c.Find(find).Sort("-_id").All(&poloniexStatsArr)
		if err != nil {
			if err.Error() == "no reachable servers" {
				slack.SendMessage(":rage:", "mongo", "alerts", fmt.Sprintf("@channel mongo problems, here is the error, but check the logs Error: %s", err.Error()))
			}
			return nil, fmt.Errorf("Mongo: getPoloniexStats: findAll: %s", err.Error())
		}

		us.cachelock.Lock()
		switch update {
		case 1:
			us.lastMonthUpdate[currency] = time.Now()
			us.cachesPoloStats[currency].MonthAvg, us.cachesPoloStats[currency].MonthStd = GetAvgAndStd(poloniexStatsArr, time.Now().Add(-24*time.Hour*30))
			fallthrough
		case 2:
			us.lastWeekUpdate[currency] = time.Now()
			us.cachesPoloStats[currency].WeekAvg, us.cachesPoloStats[currency].WeekStd = GetAvgAndStd(poloniexStatsArr, time.Now().Add(-24*time.Hour*7))
			fallthrough
		case 3:
			us.lastDayUpdate[currency] = time.Now()
			us.cachesPoloStats[currency].DayAvg, us.cachesPoloStats[currency].DayStd = GetAvgAndStd(poloniexStatsArr, time.Now().Add(-24*time.Hour))
			fallthrough
		case 4:
			us.lastHourUpdate[currency] = time.Now()
			us.cachesPoloStats[currency].HrAvg, us.cachesPoloStats[currency].HrStd = GetAvgAndStd(poloniexStatsArr, time.Now().Add(-1*time.Hour))
		}
		us.cachesPoloStats[currency].FiveMinAvg, _ = GetAvgAndStd(poloniexStatsArr, time.Now().Add(-5*time.Minute))

		v := us.cachesPoloStats[currency]
		us.cachelock.Unlock()

		return v, nil

	}
	return nil, nil

	// poloDatStats := us.GetPoloniexDataLastXDays(30, currency)

	// // No data
	// if len(poloDatStats[0]) == 0 {
	// 	return nil, nil
	// }

	// sec := GetSeconds(time.Now())
	// var lastHr []PoloniexRateSample
	// var fivemin []PoloniexRateSample

	// in := 0
	// not := 0
	// for _, v := range poloDatStats[0] {
	// 	if v.SecondsPastMidnight > sec-3600 {
	// 		in++
	// 		lastHr = append(lastHr, v)
	// 		if v.SecondsPastMidnight > sec-300 {
	// 			fivemin = append(fivemin, v)
	// 		}
	// 	} else {
	// 		not++
	// 	}
	// }

	// var all []PoloniexRateSample
	// dayCutoff := 0
	// weekCutoff := 0
	// count := 0
	// for i, v := range poloDatStats {
	// 	all = append(all, v...)
	// 	count += len(v)
	// 	if i == 1 {
	// 		dayCutoff = count
	// 	} else if i == 7 {
	// 		weekCutoff = count
	// 	}
	// }

	// poloStats.FiveMinAvg, _ = GetAvgAndStd(fivemin)
	// poloStats.HrAvg, poloStats.HrStd = GetAvgAndStd(lastHr)
	// poloStats.DayAvg, poloStats.DayStd = GetAvgAndStd(all[:dayCutoff])
	// poloStats.WeekAvg, poloStats.WeekStd = GetAvgAndStd(all[:weekCutoff])
	// poloStats.MonthAvg, poloStats.MonthStd = GetAvgAndStd(all)
	// return poloStats, nil
}

func GetAvgAndStd(data []PoloniexStat, cutoff time.Time) (avg float64, std float64) {
	total := float64(0)
	count := float64(0)

	for _, v := range data {
		if v.Time.Before(cutoff) {
			break
		}
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
	return us.RecordPoloniexStatisticTime(currency, rate, time.Now().UTC())
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
	if us.mdb != nil {
		s, c, err := us.mdb.GetCollection(mongo.C_Exchange_POL)
		if err != nil {
			return fmt.Errorf("Mongo: RecordPoloniexStatisticTime: getcol: %s", err)
		}
		defer s.Close()

		p := PoloniexStat{
			t,
			rate,
			currency,
		}

		upsertAction := bson.M{"$set": p}
		_, err = c.UpsertId(t, upsertAction)
		if err != nil {
			return fmt.Errorf("Mongo: RecordPoloniexStatisticTime: upsert: %s", err)
		}
		return nil
	}

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

//TODO REMOVE AFTER MIGRATE
//Deprecated
func (us *UserStatisticsDB) GetStatisticsOneDay(username string, day int) []AllUserStatistic {
	buc := us.getBucketPlusX(username, day*-1)
	statlist := us.getStatsFromBucket(buc)

	return statlist
}

func (us *UserStatisticsDB) GetStatistics(username string, dayRange int, exchange *UserExchange) ([][]AllUserStatistic, error) {
	if dayRange > 30 {
		return nil, fmt.Errorf("Day range must be less than 30")
	}
	stats := make([][]AllUserStatistic, dayRange)

	if us.mdb != nil {
		s, c, err := us.mdb.GetCollection(mongo.C_UserStat)
		if err != nil {
			return nil, fmt.Errorf("Mongo: GetStatistics: createSession: %s", err)
		}
		defer s.Close()

		//CAN OPTIMIZE LATER
		for i := 0; i < dayRange; i++ {
			// var temp AllUserStatistic
			// mongoRetStruct := struct {
			// 	UserStats AllUserStatistic
			// }{
			// 	temp,
			// }
			year, month, day := time.Now().UTC().Add(-time.Duration((i-1)*24) * time.Hour).Date()
			timeDayRangeStart := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

			year, month, day = time.Now().UTC().Add(-time.Duration((i)*24) * time.Hour).Date()
			timeDayRangeEnd := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

			// o1 := bson.D{{"$match", bson.M{"_id": username}}}
			// o2 := bson.D{{"$unwind", "$userstats"}}
			// o3 := bson.D{{"$match", bson.M{"$and": []bson.M{
			// 	bson.M{"userstats.time": bson.M{"$lt": timeDayRangeStart}},
			// 	bson.M{"userstats.time": bson.M{"$gt": timeDayRangeEnd}},
			// }}}}
			// o4 := bson.D{{"$project", bson.M{"_id": 0}}}
			// o5 := bson.D{{"$sort", bson.M{"userstats.time": -1}}}
			// ops := []bson.D{o1, o2, o3, o4, o5}
			// iter := c.Pipe(ops).Iter()
			// tempS := make([]AllUserStatistic, 0)

			tempS := make([]AllUserStatistic, 0)
			retStructAllStat := NewAllUserStatistic()
			var find bson.M
			if exchange != nil {
				find = bson.M{
					"$and": []bson.M{
						bson.M{"time": bson.M{"$lt": timeDayRangeStart}},
						bson.M{"time": bson.M{"$gt": timeDayRangeEnd}},
						bson.M{"email": username},
						bson.M{"exchange": *exchange},
					},
				}
			} else {
				find = bson.M{
					"$and": []bson.M{
						bson.M{"time": bson.M{"$lt": timeDayRangeStart}},
						bson.M{"time": bson.M{"$gt": timeDayRangeEnd}},
						bson.M{"email": username},
					},
				}
			}
			iter := c.Find(find).Sort("-time").Iter()
			for iter.Next(retStructAllStat) {
				tempS = append(tempS, *retStructAllStat)
			}
			stats[i] = tempS
		}

		return stats, nil
	}

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
		for _, currency := range AvaiableCoins {
			tmp := v.Currencies[currency]
			if tmp != nil {
				all[currency] = append(all[currency], *tmp)
			}
		}
	}

	count := float64(0)
	for _, currency := range AvaiableCoins {
		sD := GetDayAvg(all[currency])
		if sD != nil {
			total := sD.AvgBTCValue * (sD.NotLent + sD.Lent)
			da.LoanRate += sD.LoanRate * (sD.Lent * sD.AvgBTCValue)
			da.Lent += sD.Lent * sD.AvgBTCValue
			da.NotLent += sD.NotLent * sD.AvgBTCValue
			da.LendingPercent += sD.LendingPercent * total
			count += total
		}
	}

	da.LoanRate = da.LoanRate / count
	//da.Lent = da.Lent / count
	//da.NotLent = da.NotLent / count
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

				for _, v := range AvaiableCoins {
					var o UserStatistic
					o.Currency = v
					o.Time = old.Time
					tmp.Currencies[v] = &o
				}

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

func (us *UserStatisticsDB) Purge(username string) error {
	return us.PurgeMin(username, 0)
}

func (us *UserStatisticsDB) PurgeMin(username string, minAmount int) error {

	if us.mdb != nil {
		s, c, err := us.mdb.GetCollection(mongo.C_UserStat)
		if err != nil {
			return fmt.Errorf("Mongo: Purge: createSession: %s", err.Error())
		}
		defer s.Close()

		//CAN OPTIMIZE LATER
		for i := 0; i < 30; i++ {
			year, month, day := time.Now().Add(-time.Duration(i*24) * time.Hour).Date()
			timeDayRangeStart := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

			year, month, day = time.Now().Add(-time.Duration((i+1)*24) * time.Hour).Date()
			timeDayRangeEnd := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

			// o1 := bson.D{{"$match", bson.M{"_id": username}}}
			// o2 := bson.D{{"$unwind", "$userstats"}}
			// o3 := bson.D{{"$match", bson.M{"$and": []bson.M{
			// 	bson.M{"userstats.time": bson.M{"$lt": timeDayRangeStart}},
			// 	bson.M{"userstats.time": bson.M{"$gt": timeDayRangeEnd}},
			// }}}}
			// o4 := bson.D{{"$project", bson.M{"_id": 0, "userstats.time": 1}}}
			// o5 := bson.D{{"$sort", bson.M{"userstats.time": -1}}}
			// ops := []bson.D{o1, o2, o3, o4, o5}
			// iter := c.Pipe(ops).Iter()

			q := bson.M{
				"$and": []bson.M{
					bson.M{"time": bson.M{"$lt": timeDayRangeStart}},
					bson.M{"time": bson.M{"$gt": timeDayRangeEnd}},
					bson.M{"email": username},
				},
			}
			sel := bson.M{
				"_id":  0,
				"time": 1,
			}
			//get all record count
			counter, err := c.Find(q).Count()
			if err != nil {
				return fmt.Errorf("Mongo: Purge: count: %s", err.Error())
			}

			//remove only if counter > min amount
			if counter > minAmount {
				iter := c.Find(q).Select(sel).Sort("-time").Iter()

				removeArr := make([]time.Time, 0)
				var temp bson.M
				count := 1
				for iter.Next(&temp) {
					tempTime := temp["time"].(time.Time).UTC()
					if count%4 == 0 {
						removeArr = append(removeArr, tempTime)
					}
					count++
					if counter-count < minAmount {
						break
					}
				}
				// update := bson.M{
				// 	"$pull": bson.M{
				// 		"userstats": bson.M{
				// 			"time": bson.M{
				// 				"$in": removeArr,
				// 			},
				// 		},
				// 	},
				// }
				// _, err := c.UpdateAll(bson.M{"_id": username}, update)
				// if err != nil {
				// 	return fmt.Errorf("Mongo: Purge: updateall: %s", err.Error())
				// }

				//CAN OPTIMIZE LATER
				for _, o := range removeArr {
					sel = bson.M{
						"$and": []bson.M{
							bson.M{"email": username},
							bson.M{"time": o},
						},
					}
					err := c.Remove(sel)
					if err != nil {
						return fmt.Errorf("Mongo: Purge: remove: %s", err.Error())
					}
				}
			}
		}
		return nil
	}

	for i := 0; i < 30; i++ {
		hash := GetUsernameHash(username)
		index := primitives.Uint32ToBytes(uint32(i))
		buc := append(hash[:], index...)
		_, keys, err := us.db.GetAll(buc)
		fmt.Printf("Found %d elements for %s.\n", len(keys), username)
		del := 0
		if err != nil {
			continue
		}
		if len(keys) < 100 {
			continue
		}
		for i := 0; i < len(keys); i++ {
			if i%4 != 0 {
				err := us.db.Delete(buc, keys[i])
				if err != nil {
					fmt.Println(err)
				}
				del++
			}
		}
		fmt.Printf("Deleted %d elements\n", del)
	}
	return nil
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
	l.PoloniexData = make(map[string]*LendingHistoryEntry)
	l.BitfinexData = make(map[string]*LendingHistoryEntry)
	return l
}

func (a *AllLendingHistoryEntry) Pop() {
	for _, v := range AvaiableCoins {
		if _, ok := a.PoloniexData[v]; !ok {
			a.PoloniexData[v] = new(LendingHistoryEntry)
		}
		if _, ok := a.BitfinexData[v]; !ok {
			a.BitfinexData[v] = new(LendingHistoryEntry)
		}
	}
}

// LendingHistory
func (us *UserStatisticsDB) SaveLendingHistory(lendHist *AllLendingHistoryEntry) error {
	if us.mdb != nil {
		s, c, err := us.mdb.GetCollection(mongo.C_LendHist)
		if err != nil {
			return fmt.Errorf("Mongo: SaveLendingHistory: createSession: %s", err.Error())
		}
		defer s.Close()

		//CAN OPTIMIZE LATER
		upsertAction := bson.M{"$set": lendHist}
		_, err = c.UpsertId(lendHist.Username+lendHist.Time.String(), upsertAction)
		if err != nil {
			return fmt.Errorf("Mongo: SaveLendingHistory: upsert: %s", err)
		}
		return nil
	}

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
	if us.mdb != nil {
		result := NewAllLendingHistoryEntry()
		s, c, err := us.mdb.GetCollection(mongo.C_LendHist)
		if err != nil {
			return result, fmt.Errorf("Mongo: GetLendHistorySummary: createSession: %s", err.Error())
		}
		defer s.Close()

		//CAN OPTIMIZE LATER
		q := bson.M{
			"_id": username + t.String(),
		}
		err = c.Find(q).One(result)
		if err != nil {
			return result, fmt.Errorf("Mongo: GetLendHistorySummary: find: %s", err)
		}
		return result, nil
	}

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

func (us *UserStatisticsDB) AddBotActivityLogEntry(username string, botAct *[]BotActivityLogEntry) error {
	s, c, err := us.mdb.GetCollection(mongo.C_BotActivity)
	if err != nil {
		return fmt.Errorf("SetBotActivity: createSession: %s", err.Error())
	}
	defer s.Close()

	//CAN OPTIMIZE LATER
	upsertKey := bson.M{
		"_id": username,
	}
	upsertAction := bson.M{
		"$push": bson.M{
			"activitylog": bson.M{
				"$each":  botAct,
				"$sort":  bson.M{"time": -1},
				"$slice": 100,
			},
		},
	}

	_, err = c.Upsert(upsertKey, upsertAction)
	if err != nil {
		return fmt.Errorf("SetBotActivity: upsert: %s", err)
	}
	return nil
}

func (us *UserStatisticsDB) GetBotActivity(username string) (*BotActivity, error) {
	s, c, err := us.mdb.GetCollection(mongo.C_BotActivity)
	if err != nil {
		return nil, fmt.Errorf("GetBotActivity: createSession: %s", err.Error())
	}
	defer s.Close()

	var result BotActivity

	//CAN OPTIMIZE LATER
	q := bson.M{
		"_id": username,
	}
	err = c.Find(q).One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetBotActivity: find: %s", err)
	}
	return &result, nil
}

func (us *UserStatisticsDB) GetBotActivityTimeGreater(username string, t time.Time) (*[]BotActivityLogEntry, error) {
	s, c, err := us.mdb.GetCollection(mongo.C_BotActivity)
	if err != nil {
		return nil, fmt.Errorf("GetBotActivityTimeGreater: createSession: %s", err.Error())
	}
	defer s.Close()

	botActLogArr := make([]BotActivityLogEntry, 0)

	// //CAN OPTIMIZE LATER
	o1 := bson.D{{"$match", bson.M{"_id": username}}}
	o2 := bson.D{{"$unwind", "$activitylog"}}
	o3 := bson.D{{"$match", bson.M{"activitylog.time": bson.M{"$gt": t}}}}
	o4 := bson.D{{"$project", bson.M{"_id": 0}}}
	ops := []bson.D{o1, o2, o3, o4}
	var result bson.M
	iter := c.Pipe(ops).Iter()
	for iter.Next(&result) {
		botActLogArr = append(botActLogArr, BotActivityLogEntry{
			Time: result["activitylog"].(bson.M)["time"].(time.Time),
			Log:  result["activitylog"].(bson.M)["log"].(string),
		})
	}

	return &botActLogArr, nil
}
