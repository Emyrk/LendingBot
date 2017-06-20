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
//		Who do you think you should be lending for?
//	Initialize answers that question, then we add to the swarm
func (h *Hive) FlyIn(c net.Conn) {
	// The protocol is:
	//		1. Send IdentityRequest
	//		2. Wait for IdentityResponse
	//		3. Check if bee exists in map
	//		4. Send Assignment confirmation
	//		5. Wait for assignment confirmation
	//		6. Launch bee

	// 1,2. Send Identity Request, Wait for IdentityResponse
	wb, p, err := Initialize(c)
	if err != nil {
		// Probably a timeout. Closing the bee, it will try to flyin again if it is alive
		return
	}

	// 3. Check if bee exists
	b, newbee := h.Slaves.AttachWings(wb)
	b.Status = Initializing

	resp, ok := p.Message.(*IDResponse)
	if !ok {
		// Ok, this is the wrong message. What a dumb fucking bee
		// Also how? We check for this.
		c.Close()
		return
	}

	// 4. TODO: Fix this assignment, currently just responds with the same as given
	assignment := Assignment{Users: resp.Users}
	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))
	a := NewAssignment(b.ID, assignment)
	err = b.Encoder.Encode(a)
	if err != nil {
		b.Close()
		return
	}

	var m Parcel
	// 5. Confirm their list
	err = b.Decoder.Decode(&m)
	if err != nil {
		b.Close()
		return
	}

	resp, ok = m.Message.(*IDResponse)
	if !ok { // The wrong message. Comon man, the process is documented
		b.Close()
		return
	}

	//		See if the lists are the same
	if !CompareUserList(assignment.Users, resp.Users) {
		b.Close()
		return
	}

	// 6. Go buzzing bee!
	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))
	b.Status = Online
	if newbee {
		go b.Runloop()
	}
}

type WinglessBee struct {
	ID         string
	Encoder    *gob.Encoder
	Decoder    *gob.Decoder
	Connection net.Conn
}

func NewWinglessBee(id string, c net.Conn, e *gob.Encoder, d *gob.Decoder) *WinglessBee {
	wb := new(WinglessBee)
	wb.ID = id
	wb.Encoder = e
	wb.Decoder = d
	wb.Connection = c

	return wb
}

// Initialize is called before the bee is added to the beemap. It is used to determine the
// ID of the bee to see if it's a bee that we have offline
func Initialize(c net.Conn) (*WinglessBee, *Parcel, error) {
	enc := gob.NewEncoder(c)
	dec := gob.NewDecoder(c)

	// Deadlines to prevent deadlocks
	c.SetDeadline(time.Now().Add(60 * time.Second))

	// Request the identity
	err := enc.Encode(NewRequestIDParcel())
	if err != nil {
		// This bee never got added to the map, just let it die
		c.Close()
		return nil, nil, err
	}

	var m Parcel
	tries := 0
	for {
		c.SetDeadline(time.Now().Add(60 * time.Second))
		// We will allow 10 messages that are not the correct type.
		if tries > 10 {
			c.Close()
			return nil, nil, err
		}
		err = dec.Decode(&m)
		if err != nil {
			// This bee never got added to the map, just let it die
			c.Close()
			return nil, nil, err
		}
		// If the message is not the correct type, keep listening. Maybe he will send us
		// the right one eventually.
		if m.Type != ResponseIdentityParcel {
			tries++
			continue
		} else {
			break
		}
	}

	// Remove deadline
	c.SetDeadline(time.Time{})
	return NewWinglessBee(m.ID, c, enc, dec), &m, nil
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

func NewBeeFromWingleess(wb *WinglessBee) *Bee {
	b := new(Bee)
	b.ID = wb.ID
	b.Connection = wb.Connection
	b.Encoder = wb.Encoder
	b.Decoder = wb.Decoder

	b.SendChannel = make(chan *Parcel)
	b.RecieveChannel = make(chan *Parcel)
	b.ErrorChannel = make(chan error)
	b.Status = Initializing
	return b
}

func (b *Bee) Runloop() {
	go b.HandleSends()
	go b.HandleReceieves()
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
	return nil
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
				var _ = p
			}
		} else {
			time.Sleep(1 * time.Second)
		}
		if b.Status == Shutdown {
			return
		}
	}
}
