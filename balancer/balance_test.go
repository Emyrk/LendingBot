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

func TestInit(t *testing.T) {
	bal := NewBalancer()
	bal.Run(9911)

	bee := bee.NewBee("127.0.0.1:9911")
	err := bee.FlyIn()
	if err != nil {
		t.Error(err)
	}
	bee.Run()

	bal.ConnetionPool.AddUser(&User{Username: "Peter"})
	time.Sleep(2 * time.Second)
	fmt.Println(bee.Users)
}
