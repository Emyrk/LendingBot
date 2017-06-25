package bee

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
)

var _ = io.EOF

const (
	Online int = iota
	Offline
	Shutdown
)

type Hive struct {
	PublicKey []byte
}

type Bee struct {
	ID               string
	LastHearbeat     time.Time
	HearbeatDuration time.Duration

	userlock    sync.RWMutex
	Users       []*balancer.User
	ApiRate     float64
	LoanJobRate float64
	PublicKey   []byte

	// Send to Hive
	SendChannel chan *balancer.Parcel
	// Receieve from Hive
	RecieveChannel chan *balancer.Parcel

	// Error Channel
	ErrorChannel chan error

	Connection net.Conn
	Encoder    *gob.Encoder
	Decoder    *gob.Decoder

	Status int
	// We need reference to the master hive to send it messages
	// MasterHive *Hive

	HiveAddress string
	Home        *Hive
}

func NewBee(hiveAddress string) *Bee {
	b := new(Bee)
	b.SendChannel = make(chan *balancer.Parcel, 1000)
	b.RecieveChannel = make(chan *balancer.Parcel, 1000)
	b.ErrorChannel = make(chan error, 1000)
	b.Home = new(Hive)
	b.HiveAddress = hiveAddress
	b.PublicKey = make([]byte, 32)
	rand.Read(b.PublicKey)
	idbytes := make([]byte, 10)
	rand.Read(idbytes)
	b.ID = fmt.Sprintf("%x", idbytes)
	b.HearbeatDuration = time.Minute

	return b
}

func (b *Bee) FlyIn() error {
	err := b.PhoneHome()
	if err != nil {
		return err
	}

	err = b.Initialize()
	if err != nil {
		return err
	}

	err = b.ConfirmAssignment()
	if err != nil {
		return err
	}

	b.Status = Online
	b.Connection.SetDeadline(time.Time{})
	return nil
}

// Initialize counteracts balancer initialize. We will reponse to the requests given
func (b *Bee) Initialize() error {
	// Deadlines to prevent deadlocks
	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))

	// Get ID Req
	var p balancer.Parcel
	err := b.Decoder.Decode(&p)
	if err != nil {
		return err
	}

	public := p.Message
	b.Home.PublicKey = public

	// Send ID Resp
	resp := balancer.NewResponseIDParcel(b.ID, b.Users, b.PublicKey)

	// Deadlines to prevent deadlocks
	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))
	err = b.Encoder.Encode(resp)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bee) ConfirmAssignment() error {
	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))
	// Get Assignments
	var p balancer.Parcel
	err := b.Decoder.Decode(&p)
	if err != nil {
		return err
	}

	a := new(balancer.Assignment)
	err = json.Unmarshal(p.Message, a)
	if err != nil || p.Type != balancer.AssignmentParcel {
		return fmt.Errorf("Was not given an assignment type")
	}
	b.userlock.Lock()
	b.Users = a.Users
	b.userlock.Unlock()

	resp := balancer.NewResponseIDParcel(b.ID, a.Users, b.PublicKey)
	// Deadlines to prevent deadlocks
	b.Connection.SetDeadline(time.Now().Add(60 * time.Second))
	err = b.Encoder.Encode(resp)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bee) PhoneHome() error {
	c, err := net.Dial("tcp", b.HiveAddress)
	if err != nil {
		return err
	}
	b.Connection = c
	b.Encoder = gob.NewEncoder(c)
	b.Decoder = gob.NewDecoder(c)
	return nil
}

func (b *Bee) Run() {
	for {
		err := b.FlyIn()
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	go b.HandleSends()
	go b.HandleRecieves()
	go b.Runloop()
}

func (b *Bee) Runloop() {
	for {
		time.Sleep(100 * time.Millisecond)

		b.HandleErrors()

		switch b.Status {
		case Offline:
			err := b.FlyIn()
			if err != nil {
				time.Sleep(2 * time.Second)
			}
		case Online:
			b.ProcessParcels()

			if time.Since(b.LastHearbeat) > b.HearbeatDuration {
				b.SendHearbeat()
			}
		}
	}
}

func (b *Bee) SendHearbeat() {
	h := new(balancer.Heartbeat)
	b.userlock.RLock()
	u2 := make([]*balancer.User, len(b.Users))
	for i := range h.Users {
		tmp := *h.Users[i]
		u2[i] = &tmp
	}
	b.userlock.RUnlock()

	h.Users = u2
	h.SentTime = time.Now()
	b.LastHearbeat = time.Now()

	p := balancer.NewHeartbeat(b.ID, *h)
	b.SendChannel <- p
}

func (b *Bee) ProcessParcels() {
	for {
		select {
		case p := <-b.RecieveChannel:
			if p.ID != b.ID {
				fmt.Println("Bee ID does not match ID in parcel")
				// break
			}
			switch p.Type {
			case balancer.ChangeUserParcel:
				m := new(balancer.NewChangeUser)
				err := json.Unmarshal(p.Message, m)
				if err != nil {
					break
				}
				// A new user
				newU := -1
				b.userlock.Lock()
				for i, mu := range b.Users {
					if mu.Username == m.U.Username && mu.Exchange == m.U.Exchange {
						// Same user, get out
						newU = i
						break
					}
				}

				// Adding
				if m.Add {
					// Found the user, set the active flag
					if newU > -1 {
						b.Users[newU].Active = m.Active
						b.Users[newU].AccessKey = m.U.AccessKey
						b.Users[newU].SecretKey = m.U.SecretKey
						b.userlock.Unlock()
						break
					}
					// Add them
					m.U.Notes += fmt.Sprintf("%s [INFO] User added to lending server %s. Active: %t, Exchange: %s\n", time.Now().String(), b.ID, m.U.Active, balancer.GetExchangeString(m.U.Exchange)) + m.U.Notes
					b.Users = append(b.Users, &m.U)
				} else {
					// Removing the user
					if newU > -1 {
						// Found user
						b.Users[newU] = b.Users[len(b.Users)-1]
						b.Users = b.Users[:len(b.Users)-1]
					}

					// Not found? No need to remove
				}
				b.userlock.Unlock()
			}
		default:
			return
		}
	}
}

func (b *Bee) HandleSends() {
	for {
		if b.Status == Online {
			select {
			case p := <-b.SendChannel:
				err := b.Encoder.Encode(&p)
				if err != nil {
					b.ErrorChannel <- err
					b.Status = Offline
				}
			}
		} else {
			time.Sleep(1 * time.Second)
		}
		if b.Status == Offline {
			// return
		}
	}
}

func (b *Bee) HandleRecieves() {
	for {
		if b.Status == Online {

			var p balancer.Parcel
			err := b.Decoder.Decode(&p)
			if err != nil {
				b.ErrorChannel <- err
				b.Status = Offline
			} else {
				b.RecieveChannel <- &p
			}
		} else {
			time.Sleep(1 * time.Second)
		}

		if b.Status == Offline {
			// return
		}
	}
}

func (b *Bee) HandleErrors() {
	alreadyKilled := false
	for {
		select {
		case e := <-b.ErrorChannel:
			// if e == io.EOF {
			// 	continue
			// }

			if !alreadyKilled {
				b.Status = Offline
				b.goOffline()
			}
		default:
			return
		}
	}
	var _ = alreadyKilled
}

func (b *Bee) goOffline() {
	b.Status = Offline
	b.Connection.Close()
}

func (b *Bee) Shutdown() {
	b.Status = Shutdown
	b.Connection.Close()
}

func (b *Bee) NewParcel() *balancer.Parcel {
	p := new(balancer.Parcel)
	p.ID = b.ID
	return p
}
