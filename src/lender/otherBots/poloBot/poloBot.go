package poloBot

import (
	"encoding/json"
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
	Time time.Time   `json:"time"`
	BTC  PoloBotCoin `json:"BTC"`
	ETH  PoloBotCoin `json:"ETH"`
	XMR  PoloBotCoin `json:"XMR"`
	XRP  PoloBotCoin `json:"XRP"`
	DASH PoloBotCoin `json:"DASH"`
	LTC  PoloBotCoin `json:"LTC"`
	DOGE PoloBotCoin `json:"DOGE"`
	BTS  PoloBotCoin `json:"BTS"`
}

type PoloBotCoin struct {
	AvgLoadHoldingTime string `json:"averageLoanHoldingTime"`
	BestDuration       string `json:"bestDuration"`
	BestReturnRate     string `json:"bestReturnRate"`
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

	err = client.On("send:loanOfferParameters", func(h *gosocketio.Channel, args interface{}) {
		llog.Info("PoloBot received loanOffersParams")

		data, err := json.Marshal(args)
		if err != nil {
			llog.Error("Error marshalling: " + err.Error())
		}

		var temp PoloBotParams
		err = json.Unmarshal(data, &temp)
		if err != nil {
			llog.Error("Error umarshalling: " + err.Error())
		}
		channel <- &temp
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
