package main

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/app/core"
	"github.com/Emyrk/LendingBot/app/core/common/primitives"
	"github.com/Emyrk/LendingBot/app/scraper"
)

var _ = fmt.Println
var _ = time.Now
var _ = primitives.RandXORKey

func main() {
	s := core.NewStateWithMap()

	// sc := scraper.NewScraperWithMap(s)
	sc := scraper.NewScraper(s, "BTC")

	w := sc.NewWalker()

	/*
		v, err := scraper.IndentReturn(w.ReadLast())
		panicErr(err)
		var _ = v*/

	// fmt.Println(v)

	day, _, err := w.GetLastDayAndSecond()
	panicErr(err)

	fmt.Println(day)
	err = w.LoadDay(day)
	panicErr(err)

	for _, d := range w.TodayDay {
		ret, _ := primitives.BytesToUint32(d)
		fmt.Println(ret)
	}

}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
