package queuer

// For now, just keep users in memory. In future will need to be smarter
//	TODO:
//		- Measure queue rates and adjust for them
//		- Prioritize certain users

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/lender"

	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

type SingleUser struct {
	Username          string
	EnablesCurrencies []string
	MiniumumLoanAmts  []float64
	LendingStrategy   uint32
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

	Status string

	quit chan struct{}
}

func NewQueuer(s *core.State, l *lender.Lender) *Queuer {
	q := new(Queuer)
	q.quit = make(chan struct{})
	q.State = s
	q.Lender = l
	q.Status = "Initiated"

	return q
}

func (q *Queuer) Close() error {
	q.quit <- struct{}{}
	return nil
}

func (q *Queuer) Start() {
	ticker := time.NewTicker(time.Second * 10)
	interval := 0
	q.LoadUsers()

	last := time.Now().Add(time.Second * -70)
	// lastCalc := time.Now()

	for {
		select {
		case <-q.quit:
			q.quit <- struct{}{}
			return
		case <-ticker.C:
			QueuerCycles.Inc()
			interval++
			//if interval > 20 {
			err := q.LoadUsers()
			if err != nil {
				log.Println(err)
			}
			interval = 0
			//}

			if time.Since(last).Seconds() > 60 {
				var str = ""
				str += fmt.Sprintf("Have %d users to make jobs for\n", len(q.AllUsers))
				for _, us := range q.AllUsers {
					str += fmt.Sprintf("     %s, %v\n", us.Username, us.EnablesCurrencies)
				}
				q.Status = str
				last = time.Now()
			}
			// if time.Since(lastCalc).Seconds() > 30 {
			// 	q.calcStats()
			// }
			q.AddJobs()
		}
	}
}

var CryptoList = []string{"BTC", "BTS", "CLAM", "DOGE", "DASH", "LTC", "MAID", "STR", "XMR", "XRP", "ETH", "FCT"}

func (q *Queuer) calcStats() {
	j := lender.NewManualJob("", []float64{0}, 0, CryptoList)
	q.Lender.AddJob(j)
	QueuerJobsMade.Inc()
}

func (q *Queuer) AddJobs() {
	for _, u := range q.AllUsers {
		j := lender.NewManualJob(u.Username, u.MiniumumLoanAmts, u.LendingStrategy, u.EnablesCurrencies)
		if j.Currency == nil || j.MinimumLend == nil {
			continue
		}
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
		if !u.PoloniexKeys.APIKeyEmpty() {
			keys := u.PoloniexEnabled.Keys()
			var mins []float64
			for _, k := range keys {
				r := u.MiniumLend.Get(k)
				mins = append(mins, r)
			}
			newAll = append(newAll, &SingleUser{Username: u.Username, MiniumumLoanAmts: mins, EnablesCurrencies: keys})
		}
	}

	sort.Sort(UserList(newAll))
	q.AllUsers = newAll

	QueuerTotalUsers.Set(float64(len(newAll)))
	return nil
}
