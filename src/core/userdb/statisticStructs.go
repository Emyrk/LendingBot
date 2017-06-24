//go:generate msgp
package userdb

import (
	"fmt"
	"time"
)

type MongoAllUserStatistics struct {
	Username       string             `bson:"_id"`
	UserStatistics []AllUserStatistic `bson:"userstats"`
}

type AllUserStatistic struct {
	Currencies map[string]*UserStatistic `json:"currencies",msg:"currencies"`

	Username         string             `bson:"-" json:"username",msg:"username"`
	TotalCurrencyMap map[string]float64 `bson:"currencymap" json:"currencymap",msg:"currencymap"`
	Time             time.Time          `json:"time",msg:"time"`
	day              int                `json:"day",msg:"day"`
}

type UserStatistic struct {
	BTCRate float64

	AvailableBalance   float64   `bson:"availbal" json:"availbal",msg:"availbal"`
	ActiveLentBalance  float64   `bson:"availlent" json:"availlent",msg:"availlent"`
	OnOrderBalance     float64   `bson:"onorder"json:"onorder",msg:"onorder"`
	AverageActiveRate  float64   `bson:"activerate" json:"activerate",msg:"activerate"`
	AverageOnOrderRate float64   `bson:"onorderrate" json:"onorderrate",msg:"onorderrate"`
	HighestRate        float64   `json:"highestrate",msg:"highestrate"`
	LowestRate         float64   `json:"lowestrate",msg:"lowestrate"`
	Currency           string    `json:"currency",msg:"currency"`
	Time               time.Time `json:"time",msg:"time"`
}

type AllLendingHistoryEntry struct {
	Data      map[string]*LendingHistoryEntry `json:"data",msg:"data"`
	Time      time.Time                       `json:"time",msg:"time"`
	ShortTime string                          `json:"shorttime",msg:"shorttime"`
	Username  string                          `bson:"email" json:"username",msg:"username"`
}

type LendingHistoryEntry struct {
	Earned      float64 `json:"earned",msg:"earned"`
	Fees        float64 `json:"fees",msg:"fees"`
	AvgDuration float64 `json:"avgduration",msg:"avgduration"`
	Currency    string  `json:"currency",msg:"currency"`
	LoanCounts  int     `json:"loancount",msg:"loancount"`
}

type PoloniexStat struct {
	Time     time.Time `bson:"_id"`
	Rate     float64
	Currency string
}

func (l *AllLendingHistoryEntry) SetTime(t time.Time) {
	l.Time = t
	l.ShortTime = t.Format("Mon Jan 02")
}

func (l *AllLendingHistoryEntry) String() string {
	str := fmt.Sprintf("[%s] %s: \n", l.Username, l.ShortTime)
	for _, v := range l.Data {
		str += v.String()
	}
	return str
}

func (l *LendingHistoryEntry) String() string {
	return fmt.Sprintf("  [%s] E: %f, F: %f, D: %f, LC: %d\n", l.Currency, l.Earned, l.Fees, l.AvgDuration, l.LoanCounts)
}
