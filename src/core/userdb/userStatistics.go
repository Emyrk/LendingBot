package userdb

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/database"
)

var (
	CurrentDayBucket []byte = []byte("CurrentDay")
)

type UserStatisticsDB struct {
	db database.IDatabase

	CurrentDay   int
	CurrentIndex int // 0 to 30
}

func NewUserStatisticsDB() *UserStatisticsDB {
	u := new(UserStatisticsDB)

	userStatsPath := os.Getenv("USER_STATS_DB")
	if userStatsPath == "" {
		userStatsPath = "UserStats.db"
	}

	u.db = database.NewBoltDB(userStatsPath)
	return u
}

type UserStatistic struct {
	Username          string
	AvailableBalance  float64
	ActiveLentBalance float64
	OnOrderBalance    float64
	Time              time.Time
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

	b, err = s.Time.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

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

	td := newData[:15]
	newData = newData[15:]
	err = s.Time.UnmarshalBinary(td)
	if err != nil {
		return nil, err
	}

	return
}

func (us *UserStatisticsDB) RecordData(stats *UserStatistic) {
	//seconds := GetSeconds(stats.Time)
	//data := nil
	//us.putStats(stats.Username, seconds, data)
}

func (us *UserStatisticsDB) putStats(username string, seconds int, data []byte) error {
	key := primitives.Uint32ToBytes(uint32(seconds))
	return us.db.Put(us.getBucket(username), key, data)
}

func (us *UserStatisticsDB) getBucket(username string) []byte {
	hash := GetUsernameHash(username)
	index := primitives.Uint32ToBytes(uint32(us.CurrentIndex))
	return append(hash[:], index...)
}

func GetDay(t time.Time) int {
	return t.Day() * int(t.Month()) * t.Year()
}

func GetSeconds(t time.Time) int {
	// 3 units of time, in seconds from midnight
	sec := t.Second()
	min := t.Minute() * 60
	hour := t.Hour() * 3600

	return sec + min + hour

}
