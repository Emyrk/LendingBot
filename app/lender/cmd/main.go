package main

import (
	"fmt"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core"
	"github.com/DistributedSolutions/LendingBot/app/lender"
)

var _ = time.Second
var _ = fmt.Println

var (
	ACCESS_KEY string = ""
	SECRET_KEY string = ""
)

func main() {
	s := core.NewStateWithMap()
	err := s.NewUser("hello", "123")
	panicErr(err)

	err = s.SetUserKeys("hello", ACCESS_KEY, SECRET_KEY)
	panicErr(err)

	bal, err := s.PoloniexGetBalances("hello")
	panicErr(err)

	var _ = bal

	u, _ := s.FetchUser("hello")
	var _ = u

	l := lender.NewLender(s)
	l.CalculateLoanRate()
	fmt.Println(l.CurrentLoanRate)

	go l.Start()
	l.AddJob(lender.NewBTCJob(u))
	time.Sleep(1 * time.Second)
	l.AddJob(lender.NewBTCJob(u))
	time.Sleep(1 * time.Second)
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(1 * time.Second)
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(1 * time.Second)
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(5 * time.Second)
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(1 * time.Second)
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(5 * time.Second)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
