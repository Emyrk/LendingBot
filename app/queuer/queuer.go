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

	"github.com/Emyrk/LendingBot/app/core"
	"github.com/Emyrk/LendingBot/app/lender"
)

type SingleUser struct {
	Username        string
	MiniumumLoanAmt float64
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
	ticker := time.NewTicker(time.Second * 1)
	interval := 0
	q.LoadUsers()

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

			log.Printf("Have %d users to make jobs for", len(q.AllUsers))
			q.AddJobs()
		}
	}
}

func (q *Queuer) AddJobs() {
	for _, u := range q.AllUsers {
		j := lender.NewManualBTCJob(u.Username, u.MiniumumLoanAmt)
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
		newAll = append(newAll, &SingleUser{Username: u.Username, MiniumumLoanAmt: u.MiniumLend})
	}

	sort.Sort(UserList(newAll))
	q.AllUsers = newAll

	QueuerTotalUsers.Set(float64(len(newAll)))
	return nil
}
