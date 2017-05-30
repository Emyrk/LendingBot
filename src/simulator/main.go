package main

import (
	"fmt"
)

var _ = fmt.Println

func main() {
	s := NewSimulator()
	s.Start()

	var _ = s
}
