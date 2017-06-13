package poloBot_test

import (
	// "github.com/oguzbilgic/socketio"
	. "github.com/Emyrk/LendingBot/src/lender/otherBots/poloBot"
	"testing"
	"time"
)

func TestInviteCode(t *testing.T) {

	poloBotChannel := make(chan *PoloBotParams)
	poloBotClient, err := NewPoloBot(poloBotChannel)
	if err != nil {
		t.Errorf("Error: %s\n", err)
		return
	}

	time.Sleep(60 * time.Second)

	poloBotClient.Close()
}
