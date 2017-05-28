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

	sc := scraper.NewScraperWithMap(s)
	err := sc.Scrape("BTC")
	panicErr(err)

	sc.Scrape("BTC")

	w := sc.NewWalker()

	v, err := scraper.IndentReturn(w.ReadLast())
	panicErr(err)

	fmt.Println(v)

}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
