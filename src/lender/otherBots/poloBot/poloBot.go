package poloBot

import (
	// "encoding/json"
	"fmt"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	log "github.com/sirupsen/logrus"
)

var clog = log.WithField("package", "poloBot")

type PoloBotClient struct {
	Client *gosocketio.Client
}

type PoloBotParams struct {
	Time               time.Time `json:"time"`
	AvgLoadHoldingTime string    `json:"averageLoanHoldingTime"`
	BestReturnRate     float32   `json:"bestReturnRate"`
	BestDuration       int32     `json:"bestDuration"`
}

func NewPoloBot(channel chan *PoloBotParams) (*PoloBotClient, error) {
	p := new(PoloBotClient)
	llog := clog.WithField("method", "NewPoloBot")

	client, err := gosocketio.Dial(
		gosocketio.GetUrl("safe-hollows.crypto.zone", 80, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		return nil, fmt.Errorf("Error opening client to poloBoy: %s", err.Error())
	}

	err = client.On("send:loanOfferParameters", func(h *gosocketio.Channel, args PoloBotParams) {
		llog.Info("PoloBot received loanOffersParams")

		// var pbp PoloBotParams
		// err := json.Unmarshal(args, &pbp)
		// if err != nil {
		// 	llog.Error("Unable to unmarshal loadofferparameters.")
		// }

		channel <- &args
	})
	if err != nil {
		llog.Error("PoloBot received error on send:loanOfferParameters")
		return nil, err
	}

	err = client.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		llog.Info("PoloBot Disconnected")
	})
	if err != nil {
		llog.Error("PoloBot received error on disconnect")
		return nil, err
	}

	err = client.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		llog.Info("PoloBot Connected")
	})
	if err != nil {
		llog.Error("PoloBot received error on connected")
		return nil, err
	}

	p.Client = client
	return p, nil
}

func (c *PoloBotClient) Close() {
	c.Client.Close()
}
