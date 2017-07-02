package balancer

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var Test = false

var instanceLogger = log.WithField("instancetype", "Balancer")
var balLogger = instanceLogger.WithField("package", "balancer")

const (
	Online int = iota
	Initializing
	Offline
	Shutdown
)

// Balancer is the Queen Bee
type Balancer struct {
	ConnectionPool *Hive
	RateCalculator *QueenBee
	Listener       net.Listener
	IRS            *Auditor

	quit      bool
	salt      []byte
	cipherKey [32]byte

	auditReportLock sync.Mutex
	lastReport      *AuditReport
}

func (b *Balancer) Close() {
	b.quit = true
	fmt.Printf("Close Balancer %x\n", b.salt)
	err := b.Listener.Close()
	if err != nil {
		fmt.Println(err)
	}
	b.Listener = nil
	b.ConnectionPool.Close()
}

func (b *Balancer) AddUser(u *User) error {
	return b.ConnectionPool.AddUser(u)
}

func (b *Balancer) RemoveUser(email string, exchange int) error {
	return b.ConnectionPool.RemoveUser(email, exchange)
}

func NewBalancer(cipherKey [32]byte, dba, dbu, dbp string) *Balancer {
	b := new(Balancer)
	b.cipherKey = cipherKey
	b.ConnectionPool = NewHive(b)
	b.RateCalculator = NewRateCalculator(b.ConnectionPool, dba, dbu, dbp)
	b.salt = make([]byte, 10)
	rand.Read(b.salt)
	b.IRS = NewAuditor(b.ConnectionPool, dba, dbu, dbp, b.cipherKey)
	return b
}
func (b *Balancer) Run(port int) {
	b.Listen(port)
	go b.Accept()
	go b.ConnectionPool.Run()
	go b.RateCalculator.Run()
	go b.Runloop()
}

func (b *Balancer) GetLastReportString() string {
	str := ""
	b.auditReportLock.Lock()
	if b.lastReport == nil {
		str = "No audit report."
	} else {
		str = b.lastReport.String()
	}
	b.auditReportLock.Unlock()
	return str
}

func (b *Balancer) Runloop() {
	ticker := time.NewTicker(time.Hour)

	// Run an audit in a minute
	go func() {
		time.Sleep(1 * time.Minute)
		ar := b.IRS.PerformAudit()
		b.auditReportLock.Lock()
		b.lastReport = ar
		b.auditReportLock.Unlock()
	}()

	for _ = range ticker.C {
		ar := b.IRS.PerformAudit()
		if ar == nil {
			balLogger.WithFields(log.Fields{"func": "Runloop", "task": "Auditing"}).Errorf("Audit perform came back <nil>")
			continue
		}
		b.auditReportLock.Lock()
		b.lastReport = ar
		b.auditReportLock.Unlock()
		err := b.IRS.SaveAudit(ar)
		if err != nil {
			balLogger.WithFields(log.Fields{"func": "Runloop", "task": "Auditing"}).Errorf("Audit was not saved: %s", err.Error())
		}
	}
}

// Listen will listen on a port and add connections
func (b *Balancer) Listen(port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	b.Listener = ln
}

func (b *Balancer) Accept() {
	for {
		if b.quit {
			return
		}
		conn, err := b.Listener.Accept()
		if err == nil {
			b.NewConnection(conn)
		}

	}
}

func (m *Balancer) NewConnection(c net.Conn) {
	go m.ConnectionPool.FlyIn(c)
}

const (
	ShutdownBeeCommand int = iota
)

type Command struct {
	ID     string
	Action int
}

// Hive controls all slave connections
type Hive struct {
	BaseSlave   *Bee
	Slaves      *Swarm
	CurrentRate map[string]LoanRate
	LastAudit   time.Time
	PublicKey   []byte

	// Send to bees
	SendChannel chan *Parcel
	// What the bees are buzzing about
	RecieveChannel chan *Parcel

	CommandChannel chan *Command

	quit chan bool

	parent *Balancer
}

func NewHive(parent *Balancer) *Hive {
	h := new(Hive)
	h.CurrentRate = make(map[string]LoanRate)
	h.SendChannel = make(chan *Parcel, 1000)
	h.RecieveChannel = make(chan *Parcel, 1000)
	h.CommandChannel = make(chan *Command, 1000)
	h.PublicKey = make([]byte, 32)
	rand.Read(h.PublicKey)
	h.quit = make(chan bool, 10)
	h.Slaves = NewSwarm()
	h.parent = parent

	return h
}

func (h *Hive) Close() {
	h.quit <- true
	bees := h.Slaves.GetAllBees()
	for _, b := range bees {
		b.Shutdown()
		b.Close()
	}
}

func (h *Hive) Run() {
	go h.HandleReceives()
	go h.HandleSends()
}

func (h *Hive) HandleReceives() {
	for {
		select {
		case p := <-h.RecieveChannel:
			switch p.Type {
			case RebalanceUserParcel:
				ru := new(RebalanceUser)
				err := json.Unmarshal(p.Message, ru)
				if err != nil {
					// Log
					break
				}

				// Add the user to another Bee
				h.AddUser(&ru.U)
			}
			var _ = p
		case c := <-h.CommandChannel:
			switch c.Action {
			case ShutdownBeeCommand:
				h.Slaves.SquashBee(c.ID)
			}
			var _ = c
		case <-h.quit:
			h.quit <- true
			return
		}
	}
}

func (h *Hive) HandleSends() {
	for {
		select {
		case p := <-h.SendChannel:
			sent := h.Slaves.SendParcelTo(p.ID, p)
			if !sent {
				fmt.Println("Parcel could not be sent")
			}
		case <-h.quit:
			h.quit <- true
			return
		}
	}
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
	wb, p, err := h.Initialize(c)
	if err != nil {
		// Probably a timeout. Closing the bee, it will try to flyin again if it is alive
		return
	}

	// 3. Check if bee exists
	b, newbee := h.Slaves.AttachWings(wb)
	b.Status = Initializing

	resp := new(IDResponse)
	err = json.Unmarshal(p.Message, resp)
	if err != nil {
		// Ok, this is the wrong message. What a dumb fucking bee
		// Also how? We check for this.
		log.Error("Could not unmarshal IDResp")
		c.Close()
		return
	}
	b.PublicKey = resp.PublicKey

	// 4. TODO: Fix this assignment
	var correctList []*User
	h.Slaves.RLock()
	for _, u := range b.Users {
		gb, ok := h.Slaves.GetUserUnsafe(u.Username, u.Exchange)
		if ok {
			// Another bee has this user
			if b.ID != gb {
				continue
			}
		}
		correctList = append(correctList, u)
	}
	h.Slaves.RUnlock()

	// Ensure the users have their api keys
	for i, cu := range correctList {
		if cu.AccessKey == "" {
			gu, err := h.parent.IRS.GetFullUser(cu.Username, cu.Exchange)
			if err == nil {
				correctList[i] = gu
			} else {
				correctList[i] = correctList[len(correctList)-1] // Replace it with the last one.
				correctList = correctList[:len(correctList)-1]   // Chop off the last one.
			}
		}
	}

	assignment := Assignment{Users: correctList}

	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))
	a := NewAssignment(b.ID, assignment)
	err = b.Encoder.Encode(a)
	if err != nil {
		log.Error("2", err.Error())
		b.Close()
		return
	}

	var m Parcel
	// 5. Confirm their list
	err = b.Decoder.Decode(&m)
	if err != nil {
		log.Error("3", err.Error())
		b.Close()
		return
	}

	resp = new(IDResponse)
	err = json.Unmarshal(m.Message, resp)
	if err != nil { // The wrong message. Comon man, the process is documented
		log.Error("4", err.Error())
		b.Close()
		return
	}

	//		See if the lists are the same
	if !CompareUserList(assignment.Users, resp.Users) {
		log.Error("5", err.Error())
		b.Close()
		return
	}

	// 6. Go buzzing bee!
	b.Connection.SetDeadline(time.Time{})
	if newbee {
		go b.Runloop()
	}
	b.Status = Online
}

type WinglessBee struct {
	ID              string
	Encoder         *gob.Encoder
	Decoder         *gob.Decoder
	Connection      net.Conn
	ControllingHive *Hive
}

func NewWinglessBee(id string, c net.Conn, e *gob.Encoder, d *gob.Decoder, h *Hive) *WinglessBee {
	wb := new(WinglessBee)
	wb.ID = id
	wb.Encoder = e
	wb.Decoder = d
	wb.Connection = c
	wb.ControllingHive = h

	return wb
}

// Initialize is called before the bee is added to the beemap. It is used to determine the
// ID of the bee to see if it's a bee that we have offline
func (h *Hive) Initialize(c net.Conn) (*WinglessBee, *Parcel, error) {
	enc := gob.NewEncoder(c)
	dec := gob.NewDecoder(c)

	// Deadlines to prevent deadlocks
	c.SetDeadline(time.Now().Add(60 * time.Second))

	// Request the identity
	err := enc.Encode(NewRequestIDParcel(h.PublicKey))
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
	return NewWinglessBee(m.ID, c, enc, dec, h), &m, nil
}

type LoanRate struct {
	SimpleRate  float64
	AverageRate float64
}
