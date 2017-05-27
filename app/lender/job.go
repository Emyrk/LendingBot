package lender

import (
	user "github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

type Job struct {
	Username    string
	MinimumLend float64
	Currency    string
}

func NewBTCJob(u *user.User) *Job {
	return NewJob(u, "BTC")
}

func NewJob(u *user.User, currency string) *Job {
	return &Job{Username: u.Username, MinimumLend: u.MiniumLend, Currency: currency}
}
