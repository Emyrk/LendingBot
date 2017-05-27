package lender

import (
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core"
)

type Lender struct {
	State *core.State

	CurrentLoanRate       float64
	LastCalculateLoanRate time.Time
}

func NewLender(s *core.State) *Lender {
	l := new(Lender)
	l.State = s
	l.CurrentLoanRate = 1

	return l
}

func (l *Lender) CaluclateLoanRate() {

}

// ProcessJob will calculate the newest loan rate, then it create a loan for 0.1 btc at that rate
// for the user in the Job
func (l *Lender) ProcessJob(j *Job) {

}
