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
var CK [32]byte

func TestWaitFor(t *testing.T) {
	return
	s := time.Now()
	WaitFor(func() bool {
		return false
	}, time.Millisecond*45)
	if time.Since(s) < time.Millisecond*45 {
		t.Error("Should be at least 45ms")
	}
}

func Test_balancer_and_bee_disconnect(t *testing.T) {
	bal = balancer.NewBalancer(CK, "mongodb://localhost:27017", "", "")
	bal.Run(1151)

	beelist := make([]*bee.Bee, 0)
	for i := 0; i < 10; i++ {
		b := bee.NewBee("localhost:1151", "localhost:27017", "", "", true)
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

	bal = balancer.NewBalancer(CK, "mongodb://localhost:27017", "", "")
	bal.Run(1151)

	beelist := make(map[string]*bee.Bee)
	var keys []string
	for i := 0; i < 4; i++ {
		b := bee.NewBee("localhost:1151", "", "", "", true)
		beelist[b.ID] = b
		keys = append(keys, b.ID)
		b.Run()
		fmt.Printf("Launched Bee %d\n", i)
		fmt.Println("STATUS: ", b.Status)
	}

	WaitFor(func() bool {
		if beelist[keys[3]].Status == balancer.Online {
			return true
		}
		return false
	}, time.Second*3)
	time.Sleep(100 * time.Millisecond)

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

	// time.Sleep(500 * time.Millisecond)

	for _, b := range bal.ConnectionPool.Slaves.GetAllBees() {
		b.UserLock.RLock()
		WaitFor(func() bool {
			if len(b.Users) == 50 {
				return true
			}
			return false
		}, time.Second*1)

		if len(b.Users) != 50 {
			t.Errorf("Local bee should have quarter of all users: %d", len(b.Users))
		}
		b.RebalanceDuration = time.Second
		b.UserLock.RUnlock()

		WaitFor(func() bool {
			if len(beelist[b.ID].Users) == 50 {
				return true
			}
			return false
		}, time.Second*1)

		if len(beelist[b.ID].Users) != 50 {
			t.Errorf("Remote bee should have quarter of all users: %d", len(beelist[b.ID].Users))
		}
	}
	time.Sleep(500 * time.Millisecond)

	beelist[keys[0]].Shutdown()
	beelist[keys[1]].Shutdown()

	//check for rebalance
	localb1, _ := bal.ConnectionPool.Slaves.GetBee(beelist[keys[2]].ID)
	localb2, _ := bal.ConnectionPool.Slaves.GetBee(beelist[keys[3]].ID)
	WaitFor(func() bool {
		fmt.Println(len(localb1.Users))
		if len(localb1.Users) == 100 {
			return true
		}
		return false
	}, time.Second*3)

	if len(localb1.Users) != 100 {
		t.Errorf("Local bee should have half of all users: %d", len(localb1.Users))
	}

	WaitFor(func() bool {
		fmt.Println(len(localb1.Users))
		if len(localb2.Users) == 100 {
			return true
		}
		return false
	}, time.Second*3)
	if len(localb2.Users) != 100 {
		t.Errorf("Local bee should have half of all users: %d", len(localb2.Users))
	}
	// if len(b.Users) != 100 {
	// 	t.Errorf("Local bee should have half of all users: %d", len(b.Users))
	// }
	// if len(b.Users) != 100 {
	// 	t.Errorf("Local bee should have half of all users: %d", len(b.Users))
	// }
}
