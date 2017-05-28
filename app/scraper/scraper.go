package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core"
	"github.com/DistributedSolutions/LendingBot/app/core/common/primitives"
	"github.com/DistributedSolutions/LendingBot/app/core/database"
)

var _ = fmt.Println

var (
	LastScrape []byte = []byte("LastScrapeTime")
)

type Scraper struct {
	db    database.IDatabase
	State *core.State
}

func NewScraper(st *core.State) *Scraper {
	return newScraper(false, st)
}

func NewScraperWithMap(st *core.State) *Scraper {
	return newScraper(true, st)
}

func NewScraperFromExising(st *core.State, boltDB string) (*Scraper, error) {
	s := NewScraper(st)
	db := database.NewBoltDB(boltDB)
	bucs, err := db.ListAllBuckets()
	if err != nil {
		return s, err
	}

	for _, b := range bucs {
		if bytes.Compare(b, LastScrape) == 0 {
			continue
		}
		keys, err := db.ListAllKeys(b)
		if err != nil {
			return s, err
		}

		for _, k := range keys {
			data, err := db.Get(b, k)
			if err != nil {
				return s, err
			}
			err = s.db.Put(b, k, data)
			if err != nil {
				return s, err
			}
		}
	}

	return s, nil
}

func newScraper(withMap bool, st *core.State) *Scraper {
	s := new(Scraper)
	if withMap {
		s.db = database.NewMapDB()
	} else {
		s.db = database.NewBoltDB("Scraper.db")
	}
	s.State = st

	return s
}

func (s *Scraper) Scrape(currency string) error {
	loans, err := s.State.PoloniecGetLoanOrders(currency)
	if err != nil {
		return err
	}

	data, err := json.Marshal(loans)
	if err != nil {
		return err
	}

	t := time.Now()
	day := GetDay(t)
	sec := GetSeconds(t)

	seconds := primitives.Uint32ToBytes(uint32(sec))
	days := primitives.Uint32ToBytes(uint32(day))

	last := append(days, seconds...)
	err = s.db.Put(LastScrape, LastScrape, last)
	if err != nil {
		return err
	}
	return s.db.Put(days, seconds, data)
}

func GetDay(t time.Time) int {
	return t.Day() * int(t.Month()) * t.Year()
}

func GetSeconds(t time.Time) int {
	// 3 units of time, in seconds from midnight
	sec := t.Second()
	min := t.Minute() * 60
	hour := t.Hour() * 3600

	return sec + min + hour

}

func (s *Scraper) NewWalker() *Walker {
	w := new(Walker)
	w.scraper = s

	return w
}

func IndentReturn(jsonData []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonData, "", "\t")
	if err != nil {
		return "", err
	}

	return string(prettyJSON.Next(prettyJSON.Len())), nil
}

type Walker struct {
	scraper *Scraper
	Day     int
	Second  int
}

func (w *Walker) SetDay(day int) {
	w.Day = day
}

func (w *Walker) SetSecond(second int) {
	w.Second = second
}

func (w *Walker) ReadLast() ([]byte, error) {
	db := w.scraper.db
	data, err := db.Get(LastScrape, LastScrape)
	if err != nil {
		return nil, err
	}

	if len(data) != 8 {
		return nil, fmt.Errorf("Expect 8 bytes, found %d", len(data))
	}

	return db.Get(data[:4], data[4:])
}
