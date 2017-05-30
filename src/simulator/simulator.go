package main

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/Emyrk/LendingBot/src/lender"
)

var _, _ = time.Now, fmt.Println

var (
	DayWithData uint32 = 282370
)

type Simulator struct {
	State  *core.State
	Polo   *poloniex.FakePoloniex
	Lender *lender.Lender

	SimUser     string
	SimUserPass string
	APIKey      string
	APISecret   string
	User        *userdb.User
}

func NewSimulator() *Simulator {
	sim := new(Simulator)
	sim.SimUser = "User"
	sim.SimUserPass = "Pass"
	sim.APIKey = "Key"
	sim.APISecret = "Secret"

	s := core.NewFakePoloniexState()
	sim.State = s
	sim.State.NewUser(sim.SimUser, sim.SimUserPass)
	sim.State.SetUserKeys(sim.SimUser, sim.APIKey, sim.APISecret)
	u, err := sim.State.FetchUser(sim.SimUser)
	if err != nil {
		panic(err)
	}
	sim.User = u

	polo := s.PoloniexAPI.(*poloniex.FakePoloniex)
	sim.Polo = polo
	polo.LoadDay(primitives.Uint32ToBytes(DayWithData))

	sim.Lender = lender.NewLender(s)
	sim.Lender.CalculateInterval = 0

	return sim
}

func (s *Simulator) newJob() *lender.Job {
	return lender.NewBTCJob(s.User)
}

func (s *Simulator) AddJob() {
	s.Lender.AddJob(s.newJob())
}

func (s *Simulator) Start() {
	s.Polo.AddFunds("BTC", 10)

	ticker := time.NewTicker(time.Second)
	go s.Lender.Start()
	for _ = range ticker.C {
		s.AddJob()
		time.Sleep(10 * time.Millisecond)
		fmt.Println(s.Polo.String())
	}
}
