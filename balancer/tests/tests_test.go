package tests_test

// The longer tests
import (
	"fmt"
	"testing"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/bee"
)

var _, _ = balancer.Shutdown, bee.Online
var bal *balancer.Balancer

func TestBalancerDisconnects(t *testing.T) {
	bal = balancer.NewBalancer()
	bal.Run(1151)

	beelist := make([]*bee.Bee, 0)
	for i := 0; i < 10; i++ {
		b := bee.NewBee("localhost:1151")
		beelist = append(beelist, b)
		b.Run()
		fmt.Printf("Launched Bee %d\n", i)
	}
}

func Test_hive_disconnect(t *testing.T) {
	bal.Close()
	time.Sleep(500 * time.Nanosecond)
	bal = balancer.NewBalancer()
	bal.Run(1151)

	time.Sleep(3 * time.Second)
	if bal.ConnetionPool.Slaves.SwarmCount() != 10 {
		t.Errorf("Bees connected after disconnect is only: %d\n", bal.ConnetionPool.Slaves.SwarmCount())
	}
}
