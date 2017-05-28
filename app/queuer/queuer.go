package queuer

// For now, just keep users in memory. In future will need to be smarter
//	TODO:
//		- Measure queue rates and adjust for them
//		- Prioritize certain users

import (
	"time"

	"github.com/Emyrk/LendingBot/app/core"
)

type UserList []SingleUser

type SingleUser struct {
	Username string
}

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

// Queuer decides when to add Jobs to various queues
type Queuer struct {
	Users []SingleUser
	State *core.State

	quit chan struct{}
}

func NewQueuer(s *core.State) *Queuer {
	q := new(Queuer)
	q.quit = make(chan struct{})
	q.State = s

	return q
}

func (q *Queuer) Close() error {
	q.quit <- struct{}{}
	return nil
}

func (q *Queuer) StartQueuer() {
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-q.quit:
			q.quit <- struct{}{}
			return
		case <-ticker.C:

		}
	}
}

func (q *Queuer) LoadUsers() {

}
