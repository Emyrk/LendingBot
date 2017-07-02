package balancer_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	. "github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/bee"
	log "github.com/sirupsen/logrus"
)

var (
	_ = net.FilePacketConn
	_ = time.Now
	_ = fmt.Println
)

func init() {
	log.SetLevel(log.DebugLevel)
}

var CK [32]byte

func TestAddRemoveUser(t *testing.T) {
	Test = true

	bal := NewBalancer(CK, "mongodb://localhost:27017", "", "")
	bal.Run(9911)

	s, r := net.Pipe()
	fb := NewBee(s, bal.ConnectionPool)
	var _ = r
	bal.ConnectionPool.BaseSlave = fb

	bee := bee.NewBee("127.0.0.1:9911", "mongodb://localhost:27017", "", "", true)
	bee.Run()

	// Wait for bee to connect
	for {
		if bal.ConnectionPool.Slaves.SwarmCount() == 0 {
			time.Sleep(5 * time.Millisecond)
		} else {
			break
		}
	}

	internalBee, ok := bal.ConnectionPool.Slaves.GetBee(bee.ID)
	if !ok {
		t.Error("No bee found!")
		t.FailNow()
	}

	start := time.Now()
	for internalBee.Status != Online {
		if time.Since(start).Seconds() > 2 {
			t.Error("Timeout adding user")
			t.FailNow()
		}
	}

	bal.AddUser(&User{Username: "Peter", Exchange: BitfinexExchange})

	start = time.Now()
	for len(bee.Users) == 0 {
		if time.Since(start).Seconds() > 2 {
			t.Error("Timeout adding user")
			t.FailNow()
		}
	}

	err := bal.RemoveUser("Peter", BitfinexExchange)
	if err != nil {
		t.Error(err)
	}
	start = time.Now()
	for len(bee.Users) != 0 {
		if time.Since(start).Seconds() > 2 {
			t.Error("Timeout removing user")
			t.FailNow()
		}
	}
}
