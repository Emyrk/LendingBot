package lender

import (
	user "github.com/Emyrk/LendingBot/src/core/userdb"
)

type Job struct {
	Username    string
	MinimumLend float64
	Strategy    uint32
	Currency    string
}

func NewManualBTCJob(username string, min float64, strat uint32) *Job {
	return newJob(username, min, "BTC", strat)
}

func NewManualJob(username string, min float64, strat uint32, currency string) *Job {
	return newJob(username, min, Currency, strat)
}

func NewBTCJob(u *user.User) *Job {
	return NewJob(u, "BTC", u.LendingStrategy)
}

func NewJob(u *user.User, currency string, strat uint32) *Job {
	return &Job{Username: u.Username, MinimumLend: u.MiniumLend.BTC, Currency: currency, Strategy: strat}
}

func newJob(u string, l float64, cur string, strat uint32) *Job {
	return &Job{Username: u, MinimumLend: l, Currency: cur, Strategy: strat}
}
