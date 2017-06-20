package balancer

import (
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

func (s *Swarm) Lock() {
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
