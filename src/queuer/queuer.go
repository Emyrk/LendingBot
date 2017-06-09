package queuer

// For now, just keep users in memory. In future will need to be smarter
//	TODO:
//		- Measure queue rates and adjust for them
//		- Prioritize certain users

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/lender"
)

type SingleUser struct {
	Username        string
	MiniumumLoanAmt float64
	LendingStrategy uint32
}

type UserList []*SingleUser

func (slice UserList) Len() int {
	return len(slice)
}

func (slice UserList) Less(i, j int) bool {
	v := strings.Compare(slice[i].Username, slice[j].Username)
	if v < 0 {
		return true
	}
	return false
}

func (slice UserList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Queuer decides when to add Jobs to various queues
type Queuer struct {
	AllUsers []*SingleUser
	State    *core.State
	Lender   *lender.Lender

	quit chan struct{}
}

func NewQueuer(s *core.State, l *lender.Lender) *Queuer {
	q := new(Queuer)
	q.quit = make(chan struct{})
	q.State = s
	q.Lender = l

	return q
}

func (q *Queuer) Close() error {
	q.quit <- struct{}{}
	return nil
}

func (q *Queuer) Start() {
	ticker := time.NewTicker(time.Second * 5)
	interval := 0
	q.LoadUsers()

	last := time.Now()

	for {
		select {
		case <-q.quit:
			q.quit <- struct{}{}
			return
		case <-ticker.C:
			QueuerCycles.Inc()
			interval++
			if interval > 20 {
				err := q.LoadUsers()
				if err != nil {
					log.Println(err)
				}
				interval = 0
			}

			if time.Since(last).Seconds() > 60 {
				log.Printf("Have %d users to make jobs for", len(q.AllUsers))
				for _, us := range q.AllUsers {
					log.Printf("     %s", us.Username)
				}
				last = time.Now()
			}
			q.AddJobs()
		}
	}
}

func (q *Queuer) AddJobs() {
	if len(q.AllUsers) == 0 {
		j := lender.NewManualJob("", []float64{0, 0}, 0, []string{"BTC", "FCT"})
		q.Lender.AddJob(j)
		QueuerJobsMade.Inc()
	}
	for _, u := range q.AllUsers {
		j := lender.NewManualJob(u.Username, []float64{0.0008, 0.0002}, u.LendingStrategy, []string{"BTC", "FCT"})
		q.Lender.AddJob(j)
		QueuerJobsMade.Inc()
	}
}

func (q *Queuer) LoadUsers() error {
	all, err := q.State.FetchAllUsers()
	if err != nil {
		return err
	}

	var newAll []*SingleUser
	for _, u := range all {
		if u.PoloniexEnabled.BTC {
			newAll = append(newAll, &SingleUser{Username: u.Username, MiniumumLoanAmt: u.MiniumLend.BTC})
		}
	}

	sort.Sort(UserList(newAll))
	q.AllUsers = newAll

	QueuerTotalUsers.Set(float64(len(newAll)))
	return nil
}
