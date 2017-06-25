package tests_test

import (
	"fmt"
	"github.com/Emyrk/LendingBot/balancer"
	// "github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	// "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
	"testing"
)

// var users []userdb.User
var balUsersPOL []balancer.User
var balUsersBIT []balancer.User

func populateUserTestDB(t *testing.T) {
	// db, err := mongo.CreateTestUserDB("mongodb://localhost:27017")
	// if err != nil {
	// 	t.Errorf("Could not create dbs: %s", err.Error())
	// }

	// s, c, err := db.GetCollection(mongo.C_USER)
	// if err != nil {
	// 	t.Errorf("Error opening connection: %s\n", err.Error())
	// }
	// defer s.Close()

	// err = c.DropCollection()
	// if err != nil {
	// 	t.Errorf("Error dropping collection: %s\n", err.Error())
	// }

	// users := make([]userdb.User, 100, 100)
	balUsersPOL := make([]balancer.User, 100, 100)
	balUsersBIT := make([]balancer.User, 100, 100)
	for i := 0; i < 100; i++ {
		n := fmt.Sprintf("jimbo_%d", i)
		u, err := userdb.NewUser(n, n)
		if err != nil {
			t.Errorf("Error creating new user: %s", err.Error())
		}

		// upsertAction := bson.M{"$set": u}
		// _, err = c.UpsertId(u.Username, upsertAction)
		// if err != nil {
		// 	t.Errorf("upsert failed to add user: %s", err.Error())
		// }
		// users[i] = *u
		balUsersPOL[i] = balancer.User{
			Username: u.Username,
			Exchange: balancer.PoloniexExchange,
		}
		balUsersBIT[i] = balancer.User{
			Username: u.Username,
			Exchange: balancer.PoloniexExchange,
		}
		fmt.Println("one", i, balUsersPOL[i], balUsersBIT[i])
	}
}
