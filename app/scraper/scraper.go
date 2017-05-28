package scraper

import (
	"github.com/DistributedSolutions/LendingBot/app/core"
	"github.com/DistributedSolutions/LendingBot/app/core/database"
)

type Scraper struct {
	db    database.IDatabase
	State *core.State
}

func NewScraper(st *core.State) *Scraper {
	s := new(Scraper)
	s.db = database.NewMapDB()
	s.State = st

	return s
}

func (s *Scraper) Scrape(currency string) error {
	loans, err := s.State.PoloniecGetLoanOrders(currency)
	if err != nil {
		return err
	}
}
