//go:generate msgp
package userdb

import (
	"time"
)

type AllUserStatistic struct {
	Currencies map[string]*UserStatistic `json:"currencies",msg:"currencies"`

	Username         string             `json:"username",msg:"username"`
	TotalCurrencyMap map[string]float64 `json:"currencymap",msg:"currencymap"`
	Time             time.Time          `json:"time",msg:"time"`
	day              int                `json:"day",msg:"day"`
}

type UserStatistic struct {
	BTCRate float64

	AvailableBalance   float64   `json:"availbal",msg:"availbal"`
	ActiveLentBalance  float64   `json:"availlent",msg:"availlent"`
	OnOrderBalance     float64   `json:"onorder",msg:"onorder"`
	AverageActiveRate  float64   `json:"activerate",msg:"activerate"`
	AverageOnOrderRate float64   `json:"onorderrate",msg:"onorderrate"`
	Currency           string    `json:"currency",msg:"currency"`
	Time               time.Time `json:"time",msg:"time"`
}
