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

func TestAddRemoveUser(t *testing.T) {
	bal := NewBalancer()
	bal.Run(9911)

	s, r := net.Pipe()
	fb := NewBee(s, bal.ConnetionPool)
	var _ = r
	bal.ConnetionPool.BaseSlave = fb

	bee := bee.NewBee("127.0.0.1:9911")
	bee.Run()

	// Wait for bee to connect
	for {
		if bal.ConnetionPool.Slaves.SwarmCount() == 0 {
			time.Sleep(5 * time.Millisecond)
		} else {
			break
		}
	}

	internalBee, ok := bal.ConnetionPool.Slaves.GetBee(bee.ID)
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

func TestAddRemoveUser(t *testing.T) {