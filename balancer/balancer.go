package balancer

import (
	"encoding/gob"
	"io"
	"net"
	"sync"
	"time"
)

const (
	Online int = iota
	Initializing
	Offline
	Shutdown
)

// Balancer is the Queen Bee
type Balancer struct {
	ConnetionPool *Hive
	Listener      net.Listener
}

func NewBalancer() *Balancer {
	b := new(Balancer)
	return b
}

// Listen will listen on a port and add connections
func (b *Balancer) Listen(port int) {
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	m.Listener = ln
}

func (m *Balancer) Accept() {
	for {
		conn, _ := m.Listener.Accept()
		m.NewConnection(conn)
	}
}

func (m *Balancer) NewConnection(c net.Conn) {
	m.ConnetionPool.FlyIn(c)
}

// Hive controls all slave connections
type Hive struct {
	BaseSlave   *Bee
	Slaves      *Swarm
	CurrentRate map[string]LoanRate
	LastAudit   time.Time

	// Send to bees
	SendChannel chan *Parcel
	// What the bees are buzzing about
	RecieveChannel chan *Parcel
}

func NewHive() *Hive {
	h := new(Hive)
	h.CurrentRate = make(map[string]LoanRate)
	h.SendChannel = make(chan *Parcel, 1000)
	h.RecieveChannel = make(chan *Parcel, 1000)

	return h
}

func (h *Hive) FlyIn(c net.Conn) {
	h.Slaves.AddBee(NewBee(c))
}

type LoanRate struct {
	SimpleRate  float64
	AverageRate float64
}

type Bee struct {
	ID           string
	LastHearbeat time.Time
	Users        []string
	ApiRate      float64
	LoanJobRate  float64

	// Send to Bee
	SendChannel chan *Parcel
	// Receieve from bee
	RecieveChannel chan *Parcel

	// Error Channel
	ErrorChannel chan error

	Connection net.Conn
	Encoder    *gob.Encoder
	Decoder    *gob.Decoder

	Status int
}

func NewBee(c net.Conn) *Bee {
	b := new(Bee)
	b.ID = "unknown"
	b.Connection = c
	b.Encoder = gob.NewEncoder(c)
	b.Decoder = gob.NewDecoder(c)

	b.SendChannel = make(chan *Parcel)
	b.RecieveChannel = make(chan *Parcel)
	b.ErrorChannel = make(chan error)
	return b
}

func (b *Bee) Runloop() {
	for {
		// Handle Errors

		// Handle Sends

		// Handle Recieves
	}
}

func (b *Bee) HandleError() {
	select {
	case e := <-b.ErrorChannel:
		// Handle errors
		if e == io.EOF {
			// Reinit connection
		} else {

		}
	default:
	}
}

func (b *Bee) HandleSends() {
	select {
	case p := <-b.SendChannel:
	}
}

func (b *Bee) HandleReceieves() {
	select {
	case p := <-b.RecieveChannel:
	}
}
