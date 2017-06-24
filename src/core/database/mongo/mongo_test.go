package mongo_test

import (
	. "github.com/Emyrk/LendingBot/src/core/database/mongo"

	"time"
	// "fmt"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var db *MongoDB
var session *mgo.Session
var err error

func Test_user_db_create(t *testing.T) {
	db, err = CreateTestUserDB("mongodb://localhost:27017")
	if err != nil {
		t.Errorf("Error creating test db: %s\n", err.Error())
		t.FailNow()
	}
}

func Test_user_create_session(t *testing.T) {
	s, err := db.CreateSession()
	if err != nil {
		t.Errorf("Error creating session: %s\n", err.Error())
		t.FailNow()
	}
	session = s

	s.Close()

	s, err = db.CreateSession()
	if err != nil {
		t.Errorf("Error creating second session: %s\n", err.Error())
		t.FailNow()
	}
	session = s

	c := session.DB(USER_DB_TEST).C(USER_DB_C_USER)
	if err != c.DropCollection() {
		t.Errorf("Error dropping userdb test collection: %s\n", err.Error())
	}

	// first test1 insert
	u, err := userdb.NewUser("test1", "test1")
	if err != nil {
		t.Errorf("Error creating user test1: %s\n", err.Error())
	}
	err = c.Insert(u)
	if err != nil {
		t.Errorf("Error finding inserting test1: %s\n", err.Error())
	}

	// find test1 user
	result := new(userdb.User)
	err = c.Find(bson.M{"username": "test1"}).One(result)
	if err != nil {
		t.Errorf("Error finding user test1: %s\n", err.Error())
	}
	if result == nil {
		t.Error("Error user test2 is nil")
	}

	result.PoloniexKeys.SetEmptyIfBlank()

	if !u.IsSameAs(result) {
		t.Error("Test1 is not the same")
	}

	// insert test2 user
	u2, err := userdb.NewUser("test2", "test2")
	if err != nil {
		t.Errorf("Error creating user test2: %s\n", err.Error())
	}
	err = c.Insert(u2)
	if err != nil {
		t.Errorf("Error finding inserting test2: %s\n", err.Error())
	}

	// error finding all users
	var results []userdb.User
	iter := c.Find(nil).Sort("username").Limit(2).Iter()
	err = iter.All(&results)
	if err != nil {
		t.Errorf("Error finding test1 and test2: %s\n", err.Error())
	}
	results[0].PoloniexKeys.SetEmptyIfBlank()
	results[1].PoloniexKeys.SetEmptyIfBlank()
	if !u.IsSameAs(&results[0]) {
		t.Errorf("Error test1 is not the same")
	}
	if !u2.IsSameAs(&results[1]) {
		t.Errorf("Error test2 is not the same")
	}
}

func Test_user_close_session(t *testing.T) {
	session.Close()
}

var usdb *userdb.UserStatisticsDB

func Test_stat_db_create(t *testing.T) {
	db, err = CreateTestStatDB("mongodb://localhost:27017")
	if err != nil {
		t.Errorf("Error creating test db: %s\n", err.Error())
		t.FailNow()
	}

	if db == nil {
		t.Errorf("Error db is null")
		t.FailNow()
	}

	s, err := db.CreateSession()
	if err != nil {
		t.Errorf("Error creating session: %s\n", err.Error())
		t.FailNow()
	}
	defer s.Close()
	err = s.DB(STAT_DB_TEST).DropDatabase()
	if err != nil {
		t.Errorf("Error dropping database: %s\n", err.Error())
		t.FailNow()
	}
}

func Test_stat(t *testing.T) {
	usdb, err = userdb.NewUserStatisticsMongoDB(db)
	if err != nil {
		t.Errorf("Error creating new stat mongodb: %s\n", err.Error())
		t.FailNow()
	}

	stats := userdb.NewAllUserStatistic()
	b := userdb.NewUserStatistic("BTC", 1)
	// stats.Username = "steven"
	b.AvailableBalance = 0
	b.ActiveLentBalance = 100
	b.OnOrderBalance = 0
	b.AverageActiveRate = .4
	b.AverageOnOrderRate = .1
	stats.Currencies["BTC"] = b
	stats.Time = time.Now().Add(-20 * time.Hour)
	stats.Username = "bob"

	//add
	err = usdb.RecordData(stats)
	if err != nil {
		t.Errorf("Error recording user stat data: %s\n", err.Error())
	}

	stats.Time = stats.Time.Add(5 * time.Second)
	err = usdb.RecordData(stats)
	if err != nil {
		t.Errorf("Error recording user stat data: %s\n", err.Error())
	}

	stats.Time = stats.Time.Add(-24 * time.Hour)
	t.Log(stats.Time)
	err = usdb.RecordData(stats)
	if err != nil {
		t.Errorf("Error recording user stat data: %s\n", err.Error())
	}
	// end/add

	//get stats 2
	ustats, err := usdb.GetStatistics(stats.Username, 1)
	if err != nil {
		t.Errorf("Error getting user stat data: %s\n", err.Error())
	}

	if len(ustats[0]) != 2 {
		t.Errorf("Incorrect number of user stats TEST 1: %d", len(ustats[0]))
	}
	// end/get stats 2

	da := userdb.GetCombinedDayAverage(ustats[0])
	if da.LendingPercent != 1 {
		t.Errorf("Should be 1 is: %f\n", da.LendingPercent)
	}

	//get stats 3
	ustats, err = usdb.GetStatistics(stats.Username, 2)
	if err != nil {
		t.Errorf("Error getting user stat data: %s\n", err.Error())
	}

	if len(ustats[0]) != 2 || len(ustats[1]) != 1 {
		t.Errorf("Incorrect number of user stats TEST 2: %d, %d", len(ustats[0]), len(ustats[1]))
	}
	// end/get stats 3
}

func Test_stat_close_session(t *testing.T) {
	session.Close()
}
