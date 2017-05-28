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
	ACCESS_KEY string = "GOQI8AZ3-RO5444AI-RM27HI48-AWQ38XSF"
	SECRET_KEY string = "14a5c36f61d26eb46aed2ad910d7f6f7e7085b36d8ffd2d183074b236e2743fc436942b53b18c06a2d30a271e4e42e683684aeaf62f84ded616747cf9bcc57fd"
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

	// go l.Start()
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(1 * time.Second)
	// l.AddJob(lender.NewBTCJob(u))
	// time.Sleep(1 * time.Second)
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
