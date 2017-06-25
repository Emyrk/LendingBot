package tests_test

// The longer tests
import (
	"fmt"
	"testing"
	"time"

	"github.com/Emyrk/LendingBot/balancer"
	. "github.com/Emyrk/LendingBot/balancer/tests"
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
	err := PopulateUserTestDB()
	if err != nil {
		t.Error(err)
	}

	bal = balancer.NewBalancer()
	bal.Run(1151)

	beelist := make([]*bee.Bee, 4, 4)
	for i := 0; i < 4; i++ {
		b := bee.NewBee("localhost:1151")
		beelist[i] = b
		b.Run()
		fmt.Printf("Launched Bee %d\n", i)
		fmt.Println("STATUS: ", b.Status)
	}

	for _, u := range BalUsersPOL {
		err = bal.AddUser(u)
		if err != nil {
			t.Errorf("Add user one: %s\n", err.Error())
		}
	}
	for _, u := range BalUsersBIT {
		err = bal.AddUser(u)
		if err != nil {
			t.Errorf("Add user one: %s\n", err.Error())
		}
	}

	time.Sleep(500 * time.Millisecond)

	for i, b := range bal.ConnetionPool.Slaves.GetAllBees() {
		b.UserLock.RLock()
		if len(b.Users) != 50 {
			t.Errorf("Local bee should have quarter of all users: %d", len(b.Users))
		}
		b.UserLock.RUnlock()

		if len(beelist[i].Users) != 50 {
			t.Errorf("Remote bee should have quarter of all users: %d", len(beelist[i].Users))
		}
	}

	time.Sleep(500 * time.Millisecond)

	//close off bees half of bees
	beelist[0].Shutdown()
	beelist[1].Shutdown()

	//check for rebalance
	localb1 := bal.ConnetionPool.Slaves.GetBee(beelist[0].ID)
	localb2 := bal.ConnetionPool.Slaves.GetBee(beelist[1].ID)
	if len(localb1.Users) != 100 {
		t.Errorf("Local bee should have half of all users: %d", len(b.Users))
	}
	if len(b.Users) != 100 {
		t.Errorf("Local bee should have half of all users: %d", len(b.Users))
	}
	if len(b.Users) != 100 {
		t.Errorf("Local bee should have half of all users: %d", len(b.Users))
	}
	if len(b.Users) != 100 {
		t.Errorf("Local bee should have half of all users: %d", len(b.Users))
	}
}
