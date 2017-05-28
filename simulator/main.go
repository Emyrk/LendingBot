package main

import (
	"fmt"
)

var _ = fmt.Println

func main() {
	s := NewSimulator()
	orders, err := s.Polo.GetLoanOrders("BTC")
	fmt.Println(orders, err)

	var _ = s
}
