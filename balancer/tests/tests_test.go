package tests_test

// The longer tests
import (
	"fmt"
	"testing"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/bee"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
)

var _, _ = balancer.Shutdown, bee.Online
var bal *balancer.Balancer

func Test_balancer_and_bee_disconnect(t *testing.T) {
	bal = balancer.NewBalancer()
	bal.Run(1151)

	beelist := make([]*bee.Bee, 0)
	for i := 0; i < 10; i++ {
		b := bee.NewBee("localhost:1151")
		beelist = append(beelist, b)
		b.Run()
		fmt.Printf("Launched Bee %d\n", i)
	}

	bal.Close()
	time.Sleep(1 * time.Second)

	for _, o := range beelist {
		o.Shutdown()
	}
	time.Sleep(1 * time.Second)

	for _, b := range beelist {
		if b.Status != bee.Shutdown {
			t.Errorf("Bee not shutdown: %s", b.ID)
		}
	}
}
