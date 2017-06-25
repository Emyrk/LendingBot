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

	bal.ConnetionPool.AddUser(&User{Username: "Peter", Exchange: BitfinexExchange})

	start := time.Now()
	for len(bee.Users) == 0 {
		if time.Since(start).Seconds() > 2 {
			t.Error("Timeout adding user")
			return
		}
	}

	err := bal.ConnetionPool.RemoveUser("Peter", BitfinexExchange)
	if err != nil {
		t.Error(err)
	}
	start = time.Now()
	for len(bee.Users) != 0 {
		if time.Since(start).Seconds() > 2 {
			t.Error("Timeout removing user")
			return
		}
	}
}
