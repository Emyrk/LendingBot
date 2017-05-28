package main

import (
	"fmt"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core"
	"github.com/DistributedSolutions/LendingBot/app/scraper"
)

var _ = fmt.Println
var _ = time.Now

func main() {
	s := core.NewStateWithMap()

	sc := scraper.NewScraper(s)
	err := sc.Scrape("BTC")
	panicErr(err)

	ticker := time.NewTicker(2 * time.Second)
	for t := range ticker.C {
		sc.Scrape("BTC")
		fmt.Println("Scraped ", t)
	}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
