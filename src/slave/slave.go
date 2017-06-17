package slave

import (
	"encoding/gob"
	"fmt"
	//"io"
	"net"
	// "net/http"
	"time"

	"github.com/Emyrk/LendingBot/src/core/poloniex"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

var _ = fmt.Println

var limiter ratelimit.Limiter

var dialer = net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

var plog = log.WithField("package", "SingleSlave")

// Ensure we keep the polo limit
func init() {
	limiter = ratelimit.New(6)
	log.SetLevel(log.InfoLevel)
}

type Slave struct {
	Address string

	Connection net.Conn
	Encoder    *gob.Encoder
	Decoder    *gob.Decoder

	Requests  chan *poloniex.RequestHolder
	Responses chan *poloniex.ResponseHolder
}

func NewSlave(address string) *Slave {
	s := new(Slave)
	s.Address = address
	s.Requests = make(chan *poloniex.RequestHolder, 100)
	s.Responses = make(chan *poloniex.ResponseHolder, 100)

	return s
}

func (s *Slave) Connect() error {
	conn, err := dialer.Dial("tcp", s.Address)
	if err != nil {
		plog.Errorf("Failed to connect to master")
		return err
	} else {
		plog.Infof("Connected to master")
	}

	e := gob.NewEncoder(conn)
	d := gob.NewDecoder(conn)

	s.Encoder = e
	s.Decoder = d
	return nil
}

func (s *Slave) Reconnect() {
	rlog := log.WithFields(log.Fields{"package": "Slave", "func": "Reconnect"})
	for {
		err := s.Connect()
		if err != nil {
			rlog.Errorf("Error in reconnect(): %s", err.Error())
			time.Sleep(5 * time.Second)
		} else {
			return
		}
	}
}

func (s *Slave) Run() {
	rlog := log.WithFields(log.Fields{"package": "Slave", "func": "Run"})
	s.Reconnect()
	go s.HandleRequests()
	go s.HandleResponses()
	for {
		var req poloniex.RequestHolder
		err := s.Decoder.Decode(&req)
		if err != nil {
			rlog.Errorf("Error in read(): %s", err.Error())
			s.Reconnect()
			continue
		}
		s.Requests <- &req
	}
}

func (s *Slave) HandleRequests() {
	rlog := log.WithFields(log.Fields{"package": "Slave", "func": "HandleRequests"})
	for {
		select {
		case req := <-s.Requests:
			resp, err := poloniex.SendConstructedRequest(req)
			if err != nil {
				rlog.Errorf("Error in sending polo request: %s", err.Error())
				break
			}
			var respHolder poloniex.ResponseHolder
			respHolder.ID = req.ID
			respHolder.Response = resp
			respHolder.Err = nil
			s.Responses <- &respHolder
		}
	}
}

func (s *Slave) HandleResponses() {
	rlog := log.WithFields(log.Fields{"package": "Slave", "func": "HandleResponses"})
	for {
		select {
		case resp := <-s.Responses:
			err := s.Encoder.Encode(resp)
			if err != nil {
				rlog.Errorf("Error in sending polo response: %s", err.Error())
				break
			}
		}
	}
}
