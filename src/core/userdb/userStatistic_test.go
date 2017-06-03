package userdb_test

import (
	//"crypto"
	// "fmt"
	//"os"
	"fmt"
	"testing"
	"time"

	//"github.com/DistributedSolutions/twofactor"
	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Println

func TestUserStat(t *testing.T) {
	stats := NewUserStatistic()
	stats.Username = "steven"
	stats.AvailableBalance = 100
	stats.ActiveLentBalance = 100
	stats.OnOrderBalance = 100
	stats.AverageActiveRate = .4
	stats.AverageOnOrderRate = .1

	stats.TotalCurrencyMap["BTC"] = 1.2

	data, err := stats.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	u2 := NewUserStatistic()
	data, err = u2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	if len(data) > 0 {
		t.Error("Should be length 0")
	}

	if !stats.IsSameAs(u2) {
		t.Error("Should be same")
	}
}

/*
type UserStatistic struct {
	Username           string    `json:"username"`
	AvailableBalance   float64   `json:"availbal"`
	ActiveLentBalance  float64   `json:"availlent"`
	OnOrderBalance     float64   `json:"onorder"`
	AverageActiveRate  float64   `json:"activerate"`
	AverageOnOrderRate float64   `json:"onorderrate"`
	Time               time.Time `json:"time"`
	Currency           string    `json:"currency"`

	day int
}
*/

func TestGetDayAvg(t *testing.T) {
	u, _ := NewUserStatisticsMapDB()
	var _ = u

	stats := NewUserStatistic()
	stats.Username = "steven"
	stats.AvailableBalance = 0
	stats.ActiveLentBalance = 100
	stats.OnOrderBalance = 0
	stats.AverageActiveRate = .4
	stats.AverageOnOrderRate = .1
	stats.Time = time.Now()
	var _ = stats

	u.RecordData(stats)
	stats.AvailableBalance = 0
	stats.Time = stats.Time.Add(5 * time.Second)
	u.RecordData(stats)
	// u.RecordData(stats)

	ustats, _ := u.GetStatistics("steven", 1)
	da := GetDayAvg(ustats[0])
	if da.LendingPercent != 1 {
		t.Error("Should be 1")
	}
}

func TestStats(t *testing.T) {
	u, _ := NewUserStatisticsMapDB()
	var _ = u

	stats := NewUserStatistic()
	stats.Username = "steven"
	stats.AvailableBalance = 0
	stats.ActiveLentBalance = 100
	stats.OnOrderBalance = 0
	stats.AverageActiveRate = .4
	stats.AverageOnOrderRate = .1
	stats.Time = time.Now()
	var _ = stats

	u.RecordData(stats)
	stats.AvailableBalance = 0
	stats.Time = stats.Time.Add(5 * time.Second)
	u.RecordData(stats)
	// u.RecordData(stats)

	ustats, _ := u.GetStatistics("steven", 1)
	da := GetDayAvg(ustats[0])

	var _ = da
}

// func TestThisThing(t *testing.T) {
// 	thingy := func(i int, offset int) int {
// 		i += offset
// 		if i > 30 {
// 			overFlow := i - 30
// 			i = -1 + overFlow
// 		}

// 		if i < 0 {
// 			underFlow := i * -1
// 			i = 31 - underFlow
// 		}
// 		return i
// 	}

// 	for i := 0; i < 100; i++ {
// 		fmt.Println(thingy(1, -1*(i%30)))
// 	}

// }
