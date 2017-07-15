package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	dat, err := ioutil.ReadFile("/home/jesse/go/src/github.com/Emyrk/LendingBot/reverse/email.txt")
	if err != nil {
		panic(err)
	}
	s := string(dat)
	Reverse(s)
}

func Reverse(s string) {
	strArr := strings.Split(s, "\n")
	for i, _ := range strArr {
		fmt.Println(strArr[len(strArr)-1-i])
	}
	fmt.Println("Lines: ", len(strArr))
}
