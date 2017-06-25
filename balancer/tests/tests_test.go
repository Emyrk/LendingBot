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
	time.Sleep(5 * time.Second)
	bal = balancer.NewBalancer()
	bal.Run(1151)

	time.Sleep(6 * time.Second)
	if len(bal.ConnetionPool.Slaves.GetAllBees()) != 10 {
		t.Errorf("Bees connected after disconnect is only: %d\n", len(bal.ConnetionPool.Slaves.GetAllBees()))
	}
}
