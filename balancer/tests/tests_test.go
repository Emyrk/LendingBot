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
var err error

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

func Test_balancer_rebalance(t *testing.T) {
	populateUserTestDB(t)

	bal = balancer.NewBalancer()
	bal.Run(1151)

	beelist := make([]*bee.Bee, 4, 4)
	for i := 0; i < 4; i++ {
		b := bee.NewBee("localhost:1151")
		beelist = append(beelist, b)
		b.Run()
		fmt.Printf("Launched Bee %d\n", i)
		fmt.Println("STATUS: ", b.Status)
	}

	time.Sleep(5000 * time.Millisecond)

	for i, u := range balUsersPOL {
		err = bal.AddUser(&u)
		if err != nil {
			t.Errorf("Add user one: %s\n", err.Error())
		}
		err = bal.AddUser(&balUsersBIT[i])
		if err != nil {
			t.Errorf("Add user one: %s", err.Error())
		}
	}

	time.Sleep(10000 * time.Millisecond)

	for i, b := range bal.ConnetionPool.Slaves.GetAllBees() {
		b.UserLock.RLock()
		if len(b.Users) != 25 {
			t.Errorf("Local bee should have quarter of all users: %d", len(b.Users))
		}
		fmt.Println(len(b.Users))
		b.UserLock.RUnlock()

		fmt.Println(i)
		fmt.Println(beelist[i])
		fmt.Println(beelist[i].Users)
		if len(beelist[i].Users) != 25 {
			t.Errorf("Remote bee should have quarter of all users: %d", len(beelist[i].Users))
		}

	}
}
