package core

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

const (
	Shutdown int = iota
)

var plog = log.WithFields(log.Fields{"package": "core", "subpackage": "master"})

type Command struct {
	ID     string
	Action int
}

// Master can forward our api calls out to another person to call for us
type Master struct {
	Listener    net.Listener
	Connections map[string]*Slave
	ConMap      sync.RWMutex

	Responses       chan *poloniex.ResponseHolder
	SingleResponses map[int64]chan *poloniex.ResponseHolder
	SR              sync.RWMutex
	Commands        chan Command
}

func NewMaster() *Master {
	m := new(Master)
	m.Connections = make(map[string]*Slave)
	m.Responses = make(chan *poloniex.ResponseHolder, 100)
	m.Commands = make(chan Command, 100)
	m.SingleResponses = make(map[int64]chan *poloniex.ResponseHolder)

	return m
}

func (m *Master) Run(port int) {
	m.Listen(port)
	go m.Accept()
	go m.HandleReturns()
}

// Listen will listen on a port and add connections
func (m *Master) Listen(port int) {
	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	m.Listener = ln
}

func (m *Master) Accept() {
	for {
		conn, _ := m.Listener.Accept()
		plog.WithFields(log.Fields{"address": conn.RemoteAddr().String()}).Infof("A new slave has been connected.")
		m.NewConnection(conn)
	}
}

func (m *Master) NewConnection(con net.Conn) {
	s := NewSlave(con, m)
	go s.Run()

	m.ConMap.Lock()
	m.Connections[s.ID] = s
	m.ConMap.Unlock()
	NumberOfSlaves.Set(float64(len(m.Connections)))
}

func (m *Master) HandleCommands() {
	for {
		select {
		case c := <-m.Commands:
			switch c.Action {
			case Shutdown:
				m.ConMap.Lock()
				delete(m.Connections, c.ID)
				NumberOfSlaves.Set(float64(len(m.Connections)))
				m.ConMap.Unlock()
			}
		}
	}
}

func (m *Master) HandleReturns() {
	for {
		select {
		case resp := <-m.Responses:
			m.SR.RLock()
			v, ok := m.SingleResponses[resp.ID]
			if ok {
				v <- resp
			}
			m.SR.RUnlock()
		}
	}
}

func (m *Master) SendConstructedCall(req *poloniex.RequestHolder) (*poloniex.ResponseHolder, error) {
	start := time.Now()
	defer SlaveCallTime.Observe(float64(time.Since(start).Nanoseconds()))
	m.ConMap.RLock()
	if len(m.Connections) == 0 {
		return nil, fmt.Errorf("No slaves to use")
	}
	i := rand.Intn(len(m.Connections))
	c := 0
	var slave *Slave
	for _, s := range m.Connections {
		if c == i {
			slave = s
			break
		}
		c++
	}
	m.ConMap.RUnlock()
	retChan := make(chan *poloniex.ResponseHolder)
	m.SR.Lock()
	m.SingleResponses[req.ID] = retChan
	m.SR.Unlock()
	slave.Requests <- req
	timeout := make(chan bool)
	slave.Take()
	go func() {
		time.Sleep(130 * time.Second)
		timeout <- true
	}()
	SlaveCalls.Inc()

	select {
	case resp := <-m.SingleResponses[req.ID]:
		m.SR.Lock()
		delete(m.SingleResponses, req.ID)
		m.SR.Unlock()
		return resp, nil
	case <-timeout:
		slave.Close()
		m.ConMap.Lock()
		delete(m.Connections, slave.ID)
		m.ConMap.Unlock()
		NumberOfSlaves.Set(float64(len(m.Connections)))
		SlaveTimeouts.Inc()
		return nil, fmt.Errorf("Timedout trying to send to slave")
	}
	return nil, fmt.Errorf("Should never hit this")
}

type Slave struct {
	Connection net.Conn
	ID         string
	limiter    ratelimit.Limiter

	Encoder  *gob.Encoder
	Decoder  *gob.Decoder
	Requests chan *poloniex.RequestHolder
	// Responses chan *poloniex.ResponseHolder

	Master *Master
}

func (s *Slave) Close() {
	s.Connection.Close()
}

func (s *Slave) Take() {
	s.limiter.Take()
}

func NewSlave(con net.Conn, m *Master) *Slave {
	s := new(Slave)
	s.Connection = con
	salt := fmt.Sprintf("%s%d", con.RemoteAddr().String(), rand.Int())
	id := sha256.Sum256([]byte(salt))
	idStr := fmt.Sprintf("%x", id)
	s.ID = idStr
	s.limiter = ratelimit.New(6)

	s.Requests = make(chan *poloniex.RequestHolder, 100)
	s.Encoder = gob.NewEncoder(s.Connection)
	s.Decoder = gob.NewDecoder(s.Connection)
	s.Master = m

	return s
}

func (s *Slave) Run() {
	go s.HandleSends()
	go s.HandleRecieves()
}

// HandleSends will take any requests in the channel and send them to the slave
func (s *Slave) HandleSends() {
	hLog := plog.WithFields(log.Fields{"func": "HandleSends()", "worker": "Slave"})
	for {
		select {
		case req := <-s.Requests:
			s.limiter.Take()
			err := s.Encoder.Encode(req)
			if err != nil {
				hLog.Errorf("Error encoding request: %s", err.Error())
				resp := new(poloniex.ResponseHolder)
				resp.Err = err
				resp.ID = req.ID
				s.Master.Responses <- resp
				s.Master.Commands <- Command{ID: s.ID, Action: Shutdown}
			}
		}
	}
}

// HandleRecieves takes anything from the slave and sends it to the master
func (s *Slave) HandleRecieves() {
	hLog := plog.WithFields(log.Fields{"func": "HandleRecs()", "worker": "Slave"})
	for {
		var m poloniex.ResponseHolder
		err := s.Decoder.Decode(&m)
		if err != nil && err != io.EOF {
			hLog.Errorf("Error decoding request: %s", err.Error())
			s.Master.Commands <- Command{ID: s.ID, Action: Shutdown}
			continue
		}
		s.Master.Responses <- &m
	}
}
