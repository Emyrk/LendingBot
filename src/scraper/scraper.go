package scraper

// protoc -I ./scraper/ ./scraper/scraperGRPC/scraper.proto --go_out=plugins=grpc:scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/database"
)

var _ = fmt.Println

var (
	LastScrape []byte = []byte("LastScrapeTime")
)

type Scraper struct {
	db       database.IDatabase
	State    *core.State
	Currency string

	Walker *Walker
}

func NewScraper(st *core.State, currency string) *Scraper {
	return newScraper(false, currency, st)
}

func NewScraperWithMap(st *core.State, currency string) *Scraper {
	return newScraper(true, currency, st)
}

func NewScraperFromExising(st *core.State, currency string, boltDB string) (*Scraper, error) {
	s := NewScraper(st, currency)
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

func newScraper(withMap bool, currency string, st *core.State) *Scraper {
	s := new(Scraper)
	if withMap {
		s.db = database.NewMapDB()
	} else {
		s.db = database.NewBoltDB("BTCScraper.db")
	}
	s.State = st
	s.Currency = currency

	s.Walker = s.NewWalker()

	return s
}

func (s *Scraper) Scrape() error {
	loans, err := s.State.PoloniexGetLoanOrders(s.Currency)
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
	return int(t.Unix()) / 86400
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
	s.Walker = w

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
	Day     []byte
	Second  []byte

	Index    int
	TodayDay [][]byte
}

func (w *Walker) SetDay(day []byte) {
	w.Day = day
}

func (w *Walker) SetSecond(second []byte) {
	w.Second = second
}

func (w *Walker) GetLastDayAndSecond() (day []byte, second []byte, err error) {
	db := w.scraper.db
	data, err := db.Get(LastScrape, LastScrape)
	if err != nil {
		return nil, nil, err
	}

	if len(data) != 8 {
		return nil, nil, fmt.Errorf("Expect 8 bytes, found %d", len(data))
	}

	return data[:4], data[4:], nil
}

type SortableValues [][]byte

func (slice SortableValues) Len() int {
	return len(slice)
}

func (slice SortableValues) Less(i, j int) bool {
	v1, _ := primitives.BytesToUint32(slice[i])
	v2, _ := primitives.BytesToUint32(slice[j])
	if v1 < v2 {
		return true
	}
	return false
}

func (slice SortableValues) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (w *Walker) LoadDay(day []byte) error {
	keys, err := w.scraper.db.ListAllKeys(day)
	if err != nil {
		return err
	}

	w.Day = day

	sort.Sort(SortableValues(keys))
	w.TodayDay = keys
	w.Index = 0
	return nil
}

func (w *Walker) LoadSecond(second []byte) ([]byte, error) {
	data, err := w.scraper.db.Get(w.Day, second)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadNext reads the next loan order book in line. It will go to the next
// day if we run out of data points on this day
func (w *Walker) ReadNext() ([]byte, error) {
	if w.Index >= len(w.TodayDay) {
		// Load a new day
		u, err := primitives.BytesToUint32(w.Day)
		if err != nil {
			return nil, err
		}

		lastDay, _, err := w.GetLastDayAndSecond()
		if err != nil {
			return nil, err
		}
		lastU, err := primitives.BytesToUint32(lastDay)
		if err != nil {
			return nil, err
		}

		for {
			if u > lastU {
				return nil, fmt.Errorf("Out of data to read from. On day %d, last day is %d", u, lastU)
			}
			u++
			b := primitives.Uint32ToBytes(u)
			keys, err := w.scraper.db.ListAllKeys(b)
			if err != nil {
				continue
			}

			if len(keys) > 0 {
				w.LoadDay(b)
				break
			}
		}
	}

	second := w.TodayDay[w.Index]
	w.Index++

	return w.scraper.db.Get(w.Day, second)
}

func (w *Walker) ReadLast() ([]byte, error) {
	day, sec, err := w.GetLastDayAndSecond()
	if err != nil {
		return nil, err
	}
	return w.scraper.db.Get(day, sec)
}
