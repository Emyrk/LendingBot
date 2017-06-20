package balancer

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
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
	b.Listener = ln
}

func (b *Balancer) Accept() {
	for {
		conn, _ := b.Listener.Accept()
		b.NewConnection(conn)
	}
}

func (m *Balancer) NewConnection(c net.Conn) {
	go m.ConnetionPool.FlyIn(c)
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

// FlyIn is like a gatekeeper for the swarm. Before they enter we need to know:
//		Are you already here? (Yea the metaphor falls apart here)
//		Are you a new guy?
//	Initialize answers that question, then we add to the swarm
func (h *Hive) FlyIn(c net.Conn) {
	id, e, d, p := Initialize(c)
	h.Slaves.AddBeeFromRaw(id, c, e, d, p)
}

// Initialize is called before the bee is added to the beemap. It is used to determine the
// ID of the bee to see if it's a bee that we have offline
func Initialize(c net.Conn) (string, *gob.Encoder, *gob.Decoder, *Parcel) {
	enc := gob.NewEncoder(c)
	dec := gob.NewDecoder(c)

	// Request the identity
	err := enc.Encode(NewRequestIDParcel())
	if err != nil {
		// This bee never got added to the map, just let it die
		c.Close()
	}

	var m Parcel
	err = dec.Decode(&m)
	if err != nil {
		// This bee never got added to the map, just let it die
		c.Close()
	}

	return m.ID, enc, dec, &m
}

type LoanRate struct {
	SimpleRate  float64
	AverageRate float64
}

type Bee struct {
	ID           string
	LastHearbeat time.Time
	Users        []User
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
	b.Status = Initializing
	return b
}

func NewBeeFromRaw(id string, c net.Conn, e *gob.Encoder, d *gob.Decoder) *Bee {
	b := new(Bee)
	b.ID = id
	b.Connection = c
	b.Encoder = e
	b.Decoder = d

	b.SendChannel = make(chan *Parcel)
	b.RecieveChannel = make(chan *Parcel)
	b.ErrorChannel = make(chan error)
	b.Status = Initializing
	return b
}

func (b *Bee) Runloop() {
	go b.HandleSends()
	go b.HandleReceieves()
	b.Initialize()
	for {
		// Handle Errors
		b.HandleError()

		// React on state changes
		switch b.Status {
		case Online:
		case Offline:
		case Shutdown:
		}
	}
}

// Close will close the connection, sending an EOF and making the slave reconnect
func (b *Bee) Close() {
	b.Connection.Close()
}

// ReconnectBee will repair the connection with the given bee.
// This is because the Bees dial us, meaning to repair them, we
// actually get a new bee. Instead of adding a new bee to the map,
// we can just repair the original. The IDs must match
func (a *Bee) ReconnectBee(b *Bee) error {
	if a.ID != b.ID {
		return fmt.Errorf("IDs of bees do not match. Found %s and %s", a.ID, b.ID)
	}

	a.Connection = b.Connection
	a.Encoder = b.Encoder
	a.Decoder = b.Decoder
}

func (b *Bee) HandleError() {
	for {
		select {
		case e := <-b.ErrorChannel:
			// Handle errors
			if e == io.EOF {
				// Reinit connection
			} else {

			}
		default:
			return
		}
	}
}

// HandleSends will act until shutdown
func (b *Bee) HandleSends() {
	for {
		if b.Status == Online {
			select {
			case p := <-b.SendChannel:
				err := b.Encoder.Encode(p)
				if err != nil {
					b.ErrorChannel <- err
				}
			}
		} else {
			time.Sleep(1 * time.Second)
		}
		if b.Status == Shutdown {
			return
		}
	}
}

func (b *Bee) HandleReceieves() {
	for {
		if b.Status == Online {
			select {
			case p := <-b.RecieveChannel:
			}
		} else {
			time.Sleep(1 * time.Second)
		}
		if b.Status == Shutdown {
			return
		}
	}
}
