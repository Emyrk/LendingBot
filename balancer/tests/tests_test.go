package tests_test

// The longer tests
import (
	"fmt"
	"testing"

	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/bee"
)

var _, _ = balancer.Shutdown, bee.Online

func TestBalancerDisconnects(t *testing.T) {
	bal := balancer.NewBalancer()
	bal.Run(1151)

	beelist := make([]*bee.Bee, 0)
	for i := 0; i < 10; i++ {
		b := bee.NewBee("localhost:1151")
		beelist = append(beelist, b)
		b.Run()
		fmt.Printf("Launched Bee %d\n", i)
	}
}
