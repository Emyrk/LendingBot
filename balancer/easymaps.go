package balancer

import (
	// "encoding/gob"
	// "net"
	"sync"
)

type Swarm struct {
	swarm map[string]*Bee
	sync.RWMutex
}

func NewSwarm() *Swarm {
	s := new(Swarm)
	s.swarm = make(map[string]*Bee)
	return s
}

func (s *Swarm) Rlock() {
	s.RLock()
}

func (s *Swarm) Lock() {
	s.Lock()
}

func (s *Swarm) Unlock() {
	s.Unlock()
}

func (s *Swarm) RUnlock() {
	s.RUnlock()
}

func (s *Swarm) GetBee(id string) (*Bee, bool) {
	s.RLock()
	v, ok := s.swarm[id]
	s.RUnlock()
	return v, ok
}

func (s *Swarm) GetAndLockBee(id string, readonly bool) (*Bee, bool) {
	if readonly {
		s.Rlock()
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
			return b, true
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
	return b, false
}

func (s *Swarm) AddBee(b *Bee) {
	s.Lock()
	s.swarm[b.ID] = b
	s.Unlock()
}

func (s *Swarm) GetAllBees() []*Bee {
	var all []*Bee

	s.RLock()
	for _, b := range s.swarm {
		all = append(all, b)
	}
	s.RUnlock()
	return all
}

func (s *Swarm) SwarmCount() int {
	s.RLock()
	l := len(s.swarm)
	s.RUnlock()
	return l
}
