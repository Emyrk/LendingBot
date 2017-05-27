package main

import (
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core"
)

var _ = fmt.Println

var (
	ACCESS_KEY string = ""
	SECRET_KEY string = ""
)

func main() {
	s := core.NewState()
	err := s.NewUser("hello", "123")
	panicErr(err)

	err = s.SetUserKeys("hello", ACCESS_KEY, SECRET_KEY)
	panicErr(err)

	bal, err := s.PoloniexGetBalances("hello")
	panicErr(err)

	fmt.Println(bal)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
