package balancer

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var _ = io.EOF
var _ = log.Panic

type Bee struct {
	ID                string
	LastHearbeat      time.Time
	RebalanceDuration time.Duration
	ApiRate           float64
	LoanJobRate       float64
	PublicKey         []byte

	UserLock      sync.RWMutex
	Users         []*User
	exchangeCount map[int]int

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

	// We need reference to the master hive to send it messages
	MasterHive *Hive
}

func NewBee(c net.Conn, h *Hive) *Bee {
	b := new(Bee)
	b.ID = "unknown"
	b.Connection = c
	b.Encoder = gob.NewEncoder(c)
	b.Decoder = gob.NewDecoder(c)

	b.SendChannel = make(chan *Parcel, 1000)
	b.RecieveChannel = make(chan *Parcel, 1000)
	b.ErrorChannel = make(chan error, 100)
	b.MasterHive = h
	b.RebalanceDuration = time.Minute * 7

	b.Status = Initializing
	return b
}

func NewBeeFromWingleess(wb *WinglessBee) *Bee {
	b := new(Bee)
	b.ID = wb.ID
	b.Connection = wb.Connection
	b.Encoder = wb.Encoder
	b.Decoder = wb.Decoder

	b.SendChannel = make(chan *Parcel, 1000)
	b.RecieveChannel = make(chan *Parcel, 1000)
	b.ErrorChannel = make(chan error, 100)
	b.Status = Initializing
	b.MasterHive = wb.ControllingHive
	b.exchangeCount = make(map[int]int)
	return b
}

func (b *Bee) Runloop() {
	b.Recount()
	go b.HandleSends()
	go b.HandleReceieves()
	for {
		time.Sleep(100 * time.Millisecond)
		// Handle Errors
		b.HandleErrors()

		// React on state changes
		switch b.Status {
		case Online:
			// Process Received Parcels
			b.ProcessParcels()
		case Offline:
			// Offline for 7min+
			if time.Since(b.LastHearbeat) > b.RebalanceDuration {
				b.Shutdown()
			} else {
				time.Sleep(250 * time.Millisecond)
			}

		case Shutdown:
			// Shutdown means we close up shop and call it a day
			b.Close()
			return
		}
	}
}

// Shutdown will send a rebalance command to balancer
func (b *Bee) Shutdown() {
	b.Status = Shutdown
	b.UserLock.RLock()
	for _, u := range b.Users {
		p := NewRebalanceUserParcel(b.ID, *u)
		b.MasterHive.RecieveChannel <- p
	}
	b.UserLock.RUnlock()

	b.MasterHive.CommandChannel <- &Command{ID: b.ID, Action: ShutdownBeeCommand}
}

func (b *Bee) Recount() {
	b.UserLock.Lock()
	b.exchangeCount = make(map[int]int)
	for _, u := range b.Users {
		b.exchangeCount[u.Exchange] += 1
	}
	b.UserLock.Unlock()
}

func (b *Bee) GetUnsafeExchangeCount(exch int) (int, bool) {
	v, ok := b.exchangeCount[exch]
	return v, ok
}

func (b *Bee) GetExchangeCount(exch int) (int, bool) {
	b.UserLock.RLock()
	v, ok := b.exchangeCount[exch]
	b.UserLock.RUnlock()
	return v, ok
}

func (b *Bee) ProcessParcels() {
	for {
		select {
		case p := <-b.RecieveChannel:
			if p.ID != b.ID {
				fmt.Println("Bee ID does not match ID in parcel")
				break
			}
			switch p.Type {
			case HeartbeatParcel:
				h := new(Heartbeat)
				err := json.Unmarshal(p.Message, h)
				if err != nil {
					fmt.Println("Type of parcel is Heartbeat, but failed to cast")
					break
				}
				b.HandleHeartbeat(*h)
			}
		default:
			return
		}
	}
}

// Close will close the connection, sending an EOF and making the slave reconnect
func (b *Bee) Close() {
	b.Connection.Close()
}

func (b *Bee) ChangeUser(us *User, add, active bool) {
	b.UserLock.Lock()
	defer b.UserLock.Unlock()
	b.SendChannel <- NewChangeUserParcel(b.ID, *us, add, active)

	index := -1
	for i, u := range b.Users {
		if u.Username == us.Username && u.Exchange == us.Exchange {
			index = i
			break
		}
	}

	if !add {
		// Remove
		if index >= 0 {
			b.Users[index] = b.Users[len(b.Users)-1]
			b.Users = b.Users[:len(b.Users)-1]
		}
	} else {
		// Add
		if index == -1 {
			b.Users = append(b.Users, us)
		} else {
			b.Users[index].Active = active
		}
	}
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

// HandleErrors will clear all the errors and act appropriately
func (b *Bee) HandleErrors() bool {
	alreadyKilled := false
	for {
		select {
		case e := <-b.ErrorChannel:
			// Handle errors
			fmt.Println("INTERNAL-BEE", e)
			// if e == io.EOF {
			// 	continue
			// }

			if !alreadyKilled {
				b.Status = Offline
				b.Close()
			}
		default:
			return alreadyKilled
		}
	}
	return alreadyKilled
}

// HandleSends will act until shutdown
func (b *Bee) HandleSends() {
	for {
		if b.Status == Online {
			select {
			case p := <-b.SendChannel:
				err := b.Encoder.Encode(p)
				if err != nil {
					b.ErrorChannel <- fmt.Errorf("[HandleSends] %s", err)
					b.Status = Offline
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
			var p Parcel
			err := b.Decoder.Decode(&p)
			if err != nil {
				b.ErrorChannel <- fmt.Errorf("[HandleReceieves] %s", err)
				b.Status = Offline
			}
			b.RecieveChannel <- &p
		} else {
			time.Sleep(1 * time.Second)
		}
		if b.Status == Shutdown {
			return
		}
	}
}

func (b *Bee) HandleHeartbeat(h Heartbeat) {
	b.ApiRate = h.ApiRate
	b.LoanJobRate = h.LoanJobRate

	b.CorrectRemoteList(h.Users)
	b.LastHearbeat = time.Now()
}

func (b *Bee) CorrectRemoteList(list []*User) {
	m := make(map[string]map[int]*User)
	for _, u := range list {
		if u == nil {
			continue
		}
		if _, ok := m[u.Username]; !ok {
			m[u.Username] = make(map[int]*User)
		}
		m[u.Username][u.Exchange] = u
	}

	correctionList := make([]NewChangeUser, 0)
	b.UserLock.Lock()
	for _, u := range b.Users {
		cu := m[u.Username][u.Exchange]
		// The user does not exist on the bee, but it should
		if cu == nil {
			correctionList = append(correctionList, NewChangeUser{U: *u, Add: true, Active: u.Active})
		} else {
			// The user does exists, check the active
			if u.Active != cu.Active {
				correctionList = append(correctionList, NewChangeUser{U: *u, Add: true, Active: u.Active})
			}
		}

		// Remove from the map to signal done
		delete(m[u.Username], u.Exchange)
	}
	b.UserLock.Unlock()

	// Users that should not exist on Bee, but do
	for _, submap := range m {
		for _, v := range submap {
			correctionList = append(correctionList, NewChangeUser{U: *v, Add: false, Active: false})
		}
	}

	// Send out Corrections
	for _, c := range correctionList {
		p := NewChangeUserParcelFromStruct(b.ID, c)
		b.SendChannel <- p
	}
}
