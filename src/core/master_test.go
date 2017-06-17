package core_test

import (
	"testing"
	"time"

	. "github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/slave"
)

var _ = time.ANSIC

func TestMaster(t *testing.T) {
	m := NewMaster()
	go m.Run(1081)

	s := slave.NewSlave("localhost:1081")
	go s.Run()

	time.Sleep(1 * time.Second)

	p := poloniex.StartPoloniex()
	req, err := p.ConstructAuthenticatedLendingHistoryRequest("", "", "", "", "")
	if err != nil {
		t.Error(err)
	}

	resp, err := m.SendConstructedCall(req)
	if err != nil || resp.Response != `{"error":"Invalid API key\/secret pair."}` {
		t.Error("Did not work")
	}
}
