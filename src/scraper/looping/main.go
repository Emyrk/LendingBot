package main

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/scraper"
)

var _ = fmt.Println
var _ = time.Now

func main() {
	s := core.NewStateWithMap()

	sc := scraper.NewScraper(s, "BTC")
	err := sc.Scrape()
	panicErr(err)

	sc.Serve()

	ticker := time.NewTicker(2 * time.Second)
	for t := range ticker.C {
		sc.Scrape()
		fmt.Println("Scraped ", t)
	}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
