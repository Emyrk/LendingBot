package balancer

import (
	// "encoding/gob"
	// "net"
	"fmt"
	"runtime/debug"
	"sync"
)

var _ = fmt.Sprintf

type Swarm struct {
	swarm map[string]*Bee

	// Quick lookup to find a user
	usermap map[string]map[int]string

	locckkerr sync.RWMutex
}

func NewSwarm() *Swarm {
	s := new(Swarm)
	s.swarm = make(map[string]*Bee)
	s.usermap = make(map[string]map[int]string)
	return s
}

func (s *Swarm) RLock() {
	s.locckkerr.RLock()
	fmt.Printf("GREP RLock: %s\n", string(debug.Stack()))
}

func (s *Swarm) Lock() {
	s.locckkerr.Lock()
	fmt.Printf("GREP Lock: %s\n", string(debug.Stack()))
}

func (s *Swarm) Unlock() {
	s.locckkerr.Unlock()
	fmt.Printf("GREP Unlock: %s\n", string(debug.Stack()))
}

func (s *Swarm) RUnlock() {
	s.locckkerr.RUnlock()
	fmt.Printf("GREP RUnlock: %s\n", string(debug.Stack()))
}

func (s *Swarm) GetBeeUnsafe(id string) (*Bee, bool) {
	v, ok := s.swarm[id]
	return v, ok
}

func (s *Swarm) GetBee(id string) (*Bee, bool) {
	s.RLock()
	v, ok := s.swarm[id]
	s.RUnlock()
	return v, ok
}

func (s *Swarm) SendParcelToUnsafe(id string, p *Parcel) bool {
	if id == "ALL" {
		for _, b := range s.swarm {
			b.SendChannel <- p
		}
		return true
	}

	b, ok := s.swarm[id]
	if !ok {
		return false
	}
	b.SendChannel <- p
	return true
}

func (s *Swarm) SendParcelTo(id string, p *Parcel) bool {
	s.Lock()
	defer s.Unlock()
	return s.SendParcelToUnsafe(id, p)
}

func (s *Swarm) GetAndLockBee(id string, readonly bool) (*Bee, bool) {
	if readonly {
		s.RLock()
	} else {
		s.Lock()
	}
	v, ok := s.swarm[id]
	return v, ok
}

// AddUnsafe does not do any locking. BE SURE TO LOCK THIS SHIT!
func (s *Swarm) AddUnsafe(b *Bee) {
	s.swarm[b.ID] = b
}

func (s *Swarm) SwamCountUnsafe() int {
	return len(s.swarm)
}

// AttachWings will add a bee to our swarm.
//		true --> New bee
//		false --> Exisitng bee
func (s *Swarm) AttachWings(wb *WinglessBee) (*Bee, bool) {
	// TODO: Deal with parcel for setting up Bee
	s.Lock()
	defer s.Unlock()
	// If this bee already exists, check if it's offline
	if b, ok := s.swarm[wb.ID]; ok {
		switch b.Status {
		case Offline: // Ahh cool, it's reconnecting. Let's help him out
			b.Connection = wb.Connection
			b.Encoder = wb.Encoder
			b.Decoder = wb.Decoder
			// Go buzzing buddy!
			b.Status = Online
			return b, false
		case Initializing:
			// Umm... This is bizarre. They called twice quickly, we should replace the underlying
			fallthrough
		case Online:
			// Also bizarre. Close it up, and let it fall through to the replacement
			b.Close()
		case Shutdown:
			// Let it fall through to just replace the bee in the map
		}
	}

	// This bee does not currently exist. Welcome to the swarm buddy!
	b := NewBeeFromWingleess(wb)
	s.swarm[b.ID] = b
	return b, true
}

func (s *Swarm) AddBee(b *Bee) {
	s.Lock()
	s.swarm[b.ID] = b
	s.Unlock()
}

func (s *Swarm) AddUserUnsafe(email string, exchange int, bee string) {
	if s.usermap[email] == nil {
		s.usermap[email] = make(map[int]string)
	}
	s.usermap[email][exchange] = bee
}

func (s *Swarm) AddUser(email string, exchange int, beeID string) {
	s.Lock()
	s.AddUserUnsafe(email, exchange, beeID)
	s.Unlock()
}

func (s *Swarm) GetUser(email string, exchange int) (string, bool) {
	s.RLock()
	v, ok := s.GetUserUnsafe(email, exchange)
	s.RUnlock()

	return v, ok
}

func (s *Swarm) GetUserUnsafe(email string, exchange int) (string, bool) {
	v, ok := s.usermap[email][exchange]
	return v, ok
}

func (s *Swarm) RemoveUserUnsafe(email string, exchange int) bool {
	_, ok := s.usermap[email][exchange]
	if ok {
		delete(s.usermap[email], exchange)
	}
	return ok
}

func (s *Swarm) RemoveUser(email string, exchange int) bool {
	s.Lock()
	v := s.RemoveUserUnsafe(email, exchange)
	s.Unlock()
	return v
}

func (s *Swarm) SquashBee(id string) {
	s.Lock()
	delete(s.swarm, id)
	s.Unlock()
}

func (s *Swarm) GetAndLockAllBees(readonly bool) []*Bee {
	var all []*Bee

	if readonly {
		s.RLock()
	} else {
		s.Lock()
	}
	for _, b := range s.swarm {
		all = append(all, b)
	}
	return all
}

func (s *Swarm) GetAllBees() []*Bee {
	all := s.GetAndLockAllBees(true)
	s.RUnlock()
	return all
}

func (s *Swarm) SwarmCount() int {
	s.RLock()
	l := len(s.swarm)
	s.RUnlock()
	return l
}
