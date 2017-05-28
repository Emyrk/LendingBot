package main

import (
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/scraper/client"
)

var _ = fmt.Println

type HexInput struct {
	Hex string `json:"hex"`
}

type StringResp struct {
	Message string `json:"message"`
}

func main() {
	sc := client.NewScraperClient("scraper", "localhost:50051")
	sc.Close()
}
