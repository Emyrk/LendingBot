package bee

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
)

const (
	Online  int = iota
	Offline int = iota
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
		b.ProcessParcels()

		if time.Since(b.LastHearbeat) > b.HearbeatDuration {
			// Send Hearbeat
		}

	}
}

func (b *Bee) SendHearbeat() {
	h := new(balancer.Heartbeat)
	b.userlock.RLock()
	u2 := make([]*balancer.User, len(b.Users))
	for i := range b.Users {
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
						b.userlock.Unlock()
						break
					}
					// Add them
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
			var _ = e
			// Handle errors
			// if e == io.EOF {
			// 	// Reinit connection
			// 	if !alreadyKilled {
			// 		alreadyKilled = true
			// 		b.Status = Offline
			// 		b.Close()
			// 	}
			// }

			// if !alreadyKilled {
			// 	b.Status = Offline
			// 	b.Close()
			// }
		default:
			return
		}
	}
	var _ = alreadyKilled
}

func (b *Bee) NewParcel() *balancer.Parcel {
	p := new(balancer.Parcel)
	p.ID = b.ID
	return p
}
