package core

import (
	"time"

	log "github.com/sirupsen/logrus"
)

var filog = plog.WithFields(log.Fields{"file": "collector"})

// DebtCollector will manage debt and payments
type DebtCollector struct {
	S *State
}

func NewDebtCollector(s *State) *DebtCollector {
	d := new(DebtCollector)
	d.S = s

	return d
}

func (dc *DebtCollector) Go() {
	go dc.PaymentRoutine()
}

// PaymentRoutine checks debts every hour and makes payments
func (dc *DebtCollector) PaymentRoutine() {
	flog := filog.WithFields(log.Fields{"func": "PaymentRoutine"})
	ticker := time.NewTicker(time.Hour)
	for _ = range ticker.C {
		users, err := dc.S.FetchAllUsers()
		if err != nil {
			flog.Errorf("%s", err.Error())
			continue
		}

		for _, u := range users {
			err := dc.S.RecalcStatus(u.Username)
			if err != nil {
				flog.Errorf("%s", err.Error())
				continue
			}
		}
	}
}
