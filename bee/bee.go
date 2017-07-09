package bee

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/slack"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"

	"github.com/Emyrk/LendingBot/balancer/security"
	log "github.com/sirupsen/logrus"
)

var _ = io.EOF
var generalBeeLogger = log.WithField("instancetype", "Bee")
var beeLogger = generalBeeLogger.WithField("Package", "Bee")

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

	// LendingBot
	LendingBot *Lender

	// We need reference to the master hive to send it messages
	HiveAddress string
	HivePublic  []byte

	//db
	userStatDB *userdb.UserStatisticsDB
	userDB     *userdb.UserDatabase
}

func NewBee(hiveAddress string, dba string, dbu string, dbp string, test bool) *Bee {
	var err error

	b := new(Bee)
	b.SendChannel = make(chan *balancer.Parcel, 1000)
	b.RecieveChannel = make(chan *balancer.Parcel, 1000)
	b.ErrorChannel = make(chan error, 1000)
	b.HiveAddress = hiveAddress
	b.PublicKey = make([]byte, 32)
	rand.Read(b.PublicKey)
	idbytes := make([]byte, 10)
	rand.Read(idbytes)
	b.ID = fmt.Sprintf("%x", idbytes)
	b.HearbeatDuration = time.Minute
	b.Users = make([]*balancer.User, 0)
	b.LendingBot = NewLender(b)
	userStatDBRaw, err := mongo.CreateStatDB(dba, dbu, dbp)
	if err != nil {
		if test {
			slack.SendMessage(":rage:", b.ID, "test", fmt.Sprintf("@channel Bee %s: Oy!.. failed to connect to the userstat mongodb, I am panicing! Error: %s", b.ID, err.Error()))
		} else {
			slack.SendMessage(":rage:", b.ID, "alerts", fmt.Sprintf("@channel Bee %s: Oy!.. failed to connect to the userstat mongodb, I am panicing! Error: %s", b.ID, err.Error()))
		}
		panic(fmt.Sprintf("Failed to connect to userstat db: %s", err.Error()))
	}

	b.userStatDB, err = userdb.NewUserStatisticsMongoDBGiven(userStatDBRaw)
	if err != nil {
		panic(fmt.Sprintf("Failed to wrap userstatsdb: %s", err.Error()))
	}

	b.userDB, err = userdb.NewMongoUserDatabase(dba, dbu, dbp) //, "")
	if err != nil {
		if test {
			slack.SendMessage(":rage:", b.ID, "test", fmt.Sprintf("@channel Bee %s: Oy!.. failed to connect to the user mongodb, I am panicing! Error: %s", b.ID, err.Error()))
		} else {
			slack.SendMessage(":rage:", b.ID, "alerts", fmt.Sprintf("@channel Bee %s: Oy!.. failed to connect to the user mongodb, I am panicing! Error: %s", b.ID, err.Error()))
		}
		panic(fmt.Sprintf("Failed to connect to user db: %s", err.Error()))
	}

	return b
}

func (b *Bee) Report() string {
	str := fmt.Sprintf("==== Bee [%s] ====\n", b.ID)
	str += fmt.Sprintf("  %-15s : %s\n", "Status", balancer.StatusToString(b.Status))
	str += fmt.Sprintf("  %-15s : %s\n", "LastHeartbeat", b.LastHearbeat)
	str += fmt.Sprintf("  %-15s : %d/%d\n", "SendChannel", len(b.SendChannel), cap(b.SendChannel))
	str += fmt.Sprintf("  %-15s : %d/%d\n", "RecieveChannel", len(b.RecieveChannel), cap(b.RecieveChannel))
	str += fmt.Sprintf("  %-15s : %d/%d\n", "ErrorChannel", len(b.ErrorChannel), cap(b.ErrorChannel))
	str += fmt.Sprintf("==== Lender [%s] ====\n", b.ID)
	str += fmt.Sprintf("%s\n", b.LendingBot.Report())
	str += fmt.Sprintf("==== Users [%s] ====\n", b.ID)
	b.userlock.RLock()
	for _, u := range b.Users {
		str += u.String() + "\n"
	}
	b.userlock.RUnlock()
	return str
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
	b.HivePublic = public

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
	tlsConf, err := security.GetClientTLSConfig()
	if err != nil {
		return err
	}

	c, err := tls.Dial("tcp", b.HiveAddress, tlsConf)
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
			beeLogger.WithField("func", "FlyIn").Errorf("Error in initial FlyIn: %s", err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	go b.HandleSends()
	go b.HandleRecieves()
	go b.Runloop()
	go b.LendingBot.Runloop()
}

func (b *Bee) Runloop() {
	for {
		time.Sleep(100 * time.Millisecond)

		b.HandleErrors()

		switch b.Status {
		case Offline:
			err := b.FlyIn()
			if err != nil {
				beeLogger.WithField("func", "Runloop").Errorf("Error in reconnect FlyIn: %s", err.Error())
				time.Sleep(2 * time.Second)
			} else {
				beeLogger.WithField("func", "Runloop").Infof("Successfully recconnected to Balancer by FlyIn")
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
			if p.ID != b.ID && p.ID != "ALL" {
				beeLogger.Warningf("Bee ID does not match ID in parcel. Found ID %s, exp %s", p.ID, b.ID)
				// break
			}
			switch p.Type {
			case balancer.ChangeUserParcel:
				m := new(balancer.NewChangeUser)
				err := json.Unmarshal(p.Message, m)
				if err != nil {
					beeLogger.WithFields(log.Fields{"message": "ChangeUserParcel", "func": "ProcessParcels"}).Errorf("ChangeUserParcel failed to unmarshal: %s", err.Error())
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
						b.Users[newU].Active = true // m.Active
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
			case balancer.LendingRatesParcel:
				m := new(balancer.LendingRatesArray)
				err := json.Unmarshal(p.Message, m)
				if err != nil {
					beeLogger.WithFields(log.Fields{"message": "LendingRatesParcel", "func": "ProcessParcels"}).Errorf("LendingRatesParcel failed to unmarshal: %s", err.Error())
					break
				}

				lendingRates, ticker := m.ToMaps()
				b.LendingBot.LendingRatesChannel <- lendingRates
				b.LendingBot.TickerChannel <- ticker
			case balancer.UpdateUserNotesParcel:
				m := new(balancer.UpdateUserNotes)
				err := json.Unmarshal(p.Message, m)
				if err != nil {
					beeLogger.WithFields(log.Fields{"message": "LendingRatesParcel", "func": "ProcessParcels"}).Errorf("LendingRatesParcel failed to unmarshal: %s", err.Error())
					break
				}

				b.userlock.Lock()

				na := time.Time{}
				for _, u := range b.Users {
					if u.Username == m.Username && u.Exchange == m.Exchange {
						if u.LastTouch.Before(m.LastTouch) {
							u.LastTouch = m.LastTouch
						}
						if m.SaveMonth != na {
							u.LastHistorySaved = m.SaveMonth
						}
						if m.Notes != "" {
							u.Notes = m.Notes
						}
					}
				}
				b.userlock.Unlock()
			}
		default:
			return
		}
	}
}

func (b *Bee) updateUser(user string, exch int, notes string, lasttouch, savemonth time.Time) {
	parcel := balancer.NewUpdateUserNotesParcel(b.ID, user, exch, notes, lasttouch, savemonth)
	b.RecieveChannel <- parcel
}

func (b *Bee) HandleSends() {
	for {
		if b.Status == Online {
			select {
			case p := <-b.SendChannel:
				err := b.Encoder.Encode(&p)
				if err != nil {
					b.ErrorChannel <- fmt.Errorf("[HandleSends] Error: %s", err.Error())
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

func (b *Bee) HandleRecieves() {
	for {
		if b.Status == Online {

			var p balancer.Parcel
			err := b.Decoder.Decode(&p)
			if err != nil {
				b.ErrorChannel <- fmt.Errorf("[HandleRecieves] Error: %s", err.Error())
				b.Status = Offline
			} else {
				if len(b.RecieveChannel) >= cap(b.RecieveChannel)-1 {
					<-b.RecieveChannel
				}
				b.RecieveChannel <- &p
			}
		} else {
			time.Sleep(1 * time.Second)
		}

		if b.Status == Shutdown {
			return
		}
	}
}

func (b *Bee) HandleErrors() {
	alreadyKilled := false
	var e error
	for {
		select {
		case e = <-b.ErrorChannel:
			// if e == io.EOF {
			// 	continue
			// }
			beeLogger.WithField("func", "HandleErrors").Errorf("Going offline due to error: %s", e.Error())
			if !alreadyKilled {
				b.Status = Offline
				b.goOffline()
			}
		default:
			return
		}
	}
	var _ = alreadyKilled
	var _ = e
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
