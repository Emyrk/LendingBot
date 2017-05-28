package main

import (
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core"
)

var _ = fmt.Println

var (
	ACCESS_KEY string = "GOQI8AZ3-RO5444AI-RM27HI48-AWQ38XSF"
	SECRET_KEY string = "14a5c36f61d26eb46aed2ad910d7f6f7e7085b36d8ffd2d183074b236e2743fc436942b53b18c06a2d30a271e4e42e683684aeaf62f84ded616747cf9bcc57fd"
)

func main() {
	s := core.NewState()
	/*err := s.NewUser("hello", "123")
	panicErr(err)

	err = s.SetUserKeys("hello", ACCESS_KEY, SECRET_KEY)
	panicErr(err)*/

	bal, err := s.PoloniecGetInactiveLoans("hello")
	panicErr(err)

	fmt.Println(bal["BTC"])

	l, err := s.PoloniecGetActiveLoans("hello")
	panicErr(err)

	fmt.Println(l.Provided)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
