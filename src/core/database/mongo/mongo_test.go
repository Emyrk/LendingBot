package mongo_test

import (
	. "github.com/Emyrk/LendingBot/src/core/database/mongo"

	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

var _ = bson.M{}
var _ = fmt.Sprintf
var db *MongoDB
var session *mgo.Session
var err error

// func Test_user_db_create(t *testing.T) {
// 	db, err = CreateTestUserDB("mongodb://localhost:27017", "", "")
// 	if err != nil {
// 		t.Errorf("Error creating test db: %s\n", err.Error())
// 		t.FailNow()
// 	}
// }

// func Test_user_create_session(t *testing.T) {
// 	s, err := db.CreateSession()
// 	if err != nil {
// 		t.Errorf("Error creating session: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	session = s

// 	s.Close()

// 	s, err = db.CreateSession()
// 	if err != nil {
// 		t.Errorf("Error creating second session: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	session = s
// }

// func Test_user_testing_how_to_insert(t *testing.T) {
// 	c := session.DB(USER_DB_TEST).C(C_USER)
// 	if err != c.DropCollection() {
// 		t.Errorf("Error dropping userdb test collection: %s\n", err.Error())
// 	}

// 	// first test1 insert
// 	u, err := userdb.NewUser("test1", "test1")
// 	if err != nil {
// 		t.Errorf("Error creating user test1: %s\n", err.Error())
// 	}
// 	err = c.Insert(u)
// 	if err != nil {
// 		t.Errorf("Error finding inserting test1: %s\n", err.Error())
// 	}

// 	// find test1 user
// 	var result *userdb.User
// 	err = c.Find(bson.M{"username": "test1"}).One(result)
// 	if err != nil {
// 		t.Errorf("Error finding user test1: %s\n", err.Error())
// 	}
// 	if result == nil {
// 		t.Error("Error user test2 is nil")
// 	}

// 	result.PoloniexKeys.SetEmptyIfBlank()

// 	if !u.IsSameAs(result) {
// 		t.Error("Test1 is not the same")
// 	}

// 	// insert test2 user
// 	u2, err := userdb.NewUser("test2", "test2")
// 	if err != nil {
// 		t.Errorf("Error creating user test2: %s\n", err.Error())
// 	}
// 	err = c.Insert(u2)
// 	if err != nil {
// 		t.Errorf("Error finding inserting test2: %s\n", err.Error())
// 	}

// 	// error finding all users
// 	var results []userdb.User
// 	iter := c.Find(nil).Sort("username").Limit(2).Iter()
// 	err = iter.All(&results)
// 	if err != nil {
// 		t.Errorf("Error finding test1 and test2: %s\n", err.Error())
// 	}
// 	results[0].PoloniexKeys.SetEmptyIfBlank()
// 	results[1].PoloniexKeys.SetEmptyIfBlank()
// 	if !u.IsSameAs(&results[0]) {
// 		t.Errorf("Error test1 is not the same")
// 	}
// 	if !u2.IsSameAs(&results[1]) {
// 		t.Errorf("Error test2 is not the same")
// 	}
// }

func Test_connect_prod(t *testing.T) {
	revel_pass := os.Getenv("MONGO_REVEL_PASS")
	if revel_pass == "" {
		t.Fatalf("Need prod env var MONGO_REVEL_PASS")
	}
	db, err = CreateUserDB("mongo1.hodl.zone:4000", "revel", revel_pass)
	if err != nil {
		t.Errorf("Error creating revel db: %s\n", err.Error())
		t.FailNow()
	}
}

func Test_user_userdb(t *testing.T) {
	db, err = CreateTestUserDB("mongodb://localhost:27017", "", "")
	if err != nil {
		t.Errorf("Error creating test db: %s\n", err.Error())
		t.FailNow()
	}
	s, c, err := db.GetCollection(C_USER)
	if err != nil {
		t.Errorf("Error getting collection: %s", err.Error())
	}
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Error dropping collection: %s", err.Error())
	}
	s.Close()

	db, err = CreateTestUserDB("mongodb://localhost:27017", "", "")
	if err != nil {
		t.Errorf("Error creating userdb: %s\n", err.Error())
		t.FailNow()
	}
	udb := userdb.NewMongoUserDatabaseGiven(db)

	//add user
	u, err := userdb.NewUser("t1", "t1")
	if err != nil {
		t.Errorf("Error creating new user: %s\n", err.Error())
	}

	err = udb.PutUser(u)
	if err != nil {
		t.Errorf("Error putting user: %s\n", err.Error())
	}

	var tempU *userdb.User
	if tempU, err = udb.FetchUser("t1"); err != nil {
		t.Errorf("Error grabbing user t1: %s\n", err.Error())
	}
	if !u.IsSameAs(tempU) {
		t.Errorf("Error comparing users: %s\n", err.Error())
	}

	//update user
	err = udb.SetUserLevel("t1", userdb.Moderator)
	if err != nil {
		t.Errorf("Error changing user level: %s\n", err.Error())
	}
	if tempU, err = udb.FetchUser("t1"); err != nil {
		t.Errorf("Error grabbing updated t1: %s\n", err.Error())
	}
	u.Level = userdb.Moderator
	if !u.IsSameAs(tempU) {
		t.Errorf("Error comparing users: %s\n", err.Error())
	}

	//fetch all users
	u2, err := userdb.NewUser("t2", "t2")
	if err != nil {
		t.Errorf("Error creating new user t2: %s\n", err.Error())
	}
	err = udb.PutUser(u2)
	if err != nil {
		t.Errorf("Error putting user t2: %s\n", err.Error())
	}
	all, err := udb.FetchAllUsers()
	if err != nil {
		t.Errorf("Error fetchign all users: %s\n", err.Error())
	}
	if len(all) != 2 {
		t.Errorf("Error wrong length of all users: %d\n", len(all))
	}
	//NOT sorted, try both cases
	if (!u.IsSameAs(&all[0]) && u2.IsSameAs(&all[1])) || (!u2.IsSameAs(&all[1]) && u.IsSameAs(&all[0])) {
		t.Errorf("Error all users do not match :(\n")
	}
}

// func Test_user_close_session(t *testing.T) {
// 	session.Close()
// }

// var usdb *userdb.UserStatisticsDB

// func Test_stat_db_create(t *testing.T) {
// 	db, err = CreateTestStatDB("mongodb://localhost:27017", "", "")
// 	if err != nil {
// 		t.Errorf("Error creating test db: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	if db == nil {
// 		t.Errorf("Error db is null")
// 		t.FailNow()
// 	}

// 	s, err := db.CreateSession()
// 	if err != nil {
// 		t.Errorf("Error creating session: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	defer s.Close()
// 	err = s.DB(STAT_DB_TEST).DropDatabase()
// 	if err != nil {
// 		t.Errorf("Error dropping database: %s\n", err.Error())
// 		t.FailNow()
// 	}
// }

// func Test_stat_user(t *testing.T) {
// 	usdb, err = userdb.NewUserStatisticsMongoDBGiven(db)
// 	if err != nil {
// 		t.Errorf("Error creating new stat mongodb: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	stats := userdb.NewAllUserStatistic()
// 	b := userdb.NewUserStatistic("BTC", 1)
// 	// stats.Username = "steven"
// 	b.AvailableBalance = 0
// 	b.ActiveLentBalance = 100
// 	b.OnOrderBalance = 0
// 	b.AverageActiveRate = .4
// 	b.AverageOnOrderRate = .1
// 	stats.Currencies["BTC"] = b
// 	stats.Time = time.Now().UTC()
// 	stats.Username = "bob"

// 	//add
// 	err = usdb.RecordData(stats)
// 	if err != nil {
// 		t.Errorf("Error recording user stat data: %s\n", err.Error())
// 	}

// 	stats.Time = stats.Time.Add(-5 * time.Second).UTC()
// 	err = usdb.RecordData(stats)
// 	if err != nil {
// 		t.Errorf("Error recording user stat data: %s\n", err.Error())
// 	}

// 	stats.Time = stats.Time.Add(-24 * time.Hour).UTC()

// 	err = usdb.RecordData(stats)
// 	if err != nil {
// 		t.Errorf("Error recording user stat data: %s\n", err.Error())
// 	}
// 	// end/add

// 	//get stats 2
// 	polExch := userdb.UserExchange("pol")
// 	ustats, err := usdb.GetStatistics(stats.Username, 1, &polExch)
// 	if err != nil {
// 		t.Errorf("Error getting user stat data: %s\n", err.Error())
// 	}

// 	if len(ustats[0]) != 2 {
// 		t.Errorf("Incorrect number of user stats TEST 1: %d", len(ustats[0]))
// 	}
// 	// end/get stats 2

// 	da := userdb.GetCombinedDayAverage(ustats[0])
// 	if da.LendingPercent != 1 {
// 		t.Errorf("Should be 1 is: %f\n", da.LendingPercent)
// 	}

// 	//get stats 3
// 	ustats, err = usdb.GetStatistics(stats.Username, 2, &polExch)
// 	if err != nil {
// 		t.Errorf("Error getting user stat data: %s\n", err.Error())
// 	}

// 	if len(ustats[0]) != 2 || len(ustats[1]) != 1 {
// 		t.Errorf("Incorrect number of user stats TEST 2: %d, %d", len(ustats[0]), len(ustats[1]))
// 	}
// 	// end/get stats 3
// }

// func Test_stat_purge(t *testing.T) {
// 	db, err = CreateTestStatDB("mongodb://localhost:27017", "", "")
// 	if err != nil {
// 		t.Errorf("Error creating test db: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	if db == nil {
// 		t.Errorf("Error db is null")
// 		t.FailNow()
// 	}

// 	s, c, err := db.GetCollection(C_UserStat)
// 	if err != nil {

// 		t.Errorf("Error creating session: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	defer s.Close()
// 	err = s.DB(STAT_DB_TEST).DropDatabase()
// 	if err != nil {
// 		t.Errorf("Error dropping database: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	usdb, err = userdb.NewUserStatisticsMongoDBGiven(db)
// 	if err != nil {
// 		t.Errorf("Error creating new stat mongodb: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	stats := userdb.NewAllUserStatistic()
// 	stats.Username = "tot"

// 	year, month, day := time.Now().Date()
// 	timeToday := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
// 	timeYesterday := timeToday.Add(-24 * time.Hour)
// 	timeArr := []time.Time{
// 		timeToday.Add(-1 * time.Minute), //0
// 		timeToday.Add(-2 * time.Minute),
// 		timeToday.Add(-3 * time.Minute),
// 		timeToday.Add(-4 * time.Minute), //3

// 		timeYesterday.Add(-1 * time.Minute), //4
// 		timeYesterday.Add(-2 * time.Minute),
// 		timeYesterday.Add(-3 * time.Minute), //6
// 		timeYesterday.Add(-4 * time.Minute), //gone
// 		timeYesterday.Add(-5 * time.Minute), //8
// 		timeYesterday.Add(-6 * time.Minute),
// 		timeYesterday.Add(-7 * time.Minute), //10
// 		timeYesterday.Add(-8 * time.Minute), //gone
// 	}
// 	for _, o := range timeArr {
// 		stats.Time = o
// 		err = usdb.RecordData(stats)
// 		if err != nil {
// 			t.Errorf("Error recording user stat data: %s\n", err.Error())
// 		}
// 	}

// 	err = usdb.PurgeMin("tot", 0)
// 	if err != nil {
// 		t.Errorf("Error purge: %s\n", err.Error())
// 	}

// 	// o1 := bson.D{{"$match", bson.M{"_id": "tot"}}}
// 	// o2 := bson.D{{"$unwind", "$userstats"}}
// 	// o4 := bson.D{{"$project", bson.M{"_id": 0, "userstats.time": 1}}}
// 	// o5 := bson.D{{"$sort", bson.M{"userstats.time": -1}}}
// 	// ops := []bson.D{o1, o2, o4, o5}
// 	// var results []bson.M
// 	// err = c.Pipe(ops).All(&results)
// 	// if err != nil {
// 	// 	t.Errorf("Error recording user stat data: %s\n", err.Error())
// 	// }

// 	var results []bson.M
// 	err = c.Find(nil).Sort("-time").All(&results)
// 	if err != nil {
// 		t.Errorf("Error recording user stat data: %s\n", err.Error())
// 	}

// 	//1
// 	tempTime := results[0]["time"].(time.Time).UTC()
// 	tempTime2 := timeArr[0]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//2
// 	tempTime = results[1]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[1]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//3
// 	tempTime = results[2]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[2]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//1
// 	tempTime = results[3]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[4]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//2
// 	tempTime = results[4]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[5]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//3
// 	tempTime = results[5]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[6]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//4
// 	tempTime = results[6]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[8]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//5
// 	tempTime = results[7]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[9]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}

// 	//6
// 	tempTime = results[8]["time"].(time.Time).UTC()
// 	tempTime2 = timeArr[10]
// 	if tempTime != tempTime2 {
// 		t.Errorf("Time not matching: [%s], [%s]\n", tempTime.String(), tempTime2.String())
// 	}
// }

// func Test_lending_history_db_create(t *testing.T) {
// 	db, err = CreateTestStatDB("mongodb://localhost:27017", "", "")
// 	if err != nil {
// 		t.Errorf("Error creating test db: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	if db == nil {
// 		t.Errorf("Error db is null")
// 		t.FailNow()
// 	}

// 	s, err := db.CreateSession()
// 	if err != nil {
// 		t.Errorf("Error creating session: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	defer s.Close()
// 	err = s.DB(STAT_DB_TEST).C(C_LendHist).DropCollection()
// 	if err != nil {
// 		t.Errorf("Error dropping collection: %s\n", err.Error())
// 		t.FailNow()
// 	}
// }

// func Test_lending_history_stat(t *testing.T) {
// 	usdb, err = userdb.NewUserStatisticsMongoDBGiven(db)
// 	if err != nil {
// 		t.Errorf("Error creating new stat mongodb: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	lendHist := userdb.NewAllLendingHistoryEntry()
// 	lendHist.Time = time.Now().Add(-24 * time.Hour)
// 	lendHist.Username = "ted"

// 	err = usdb.SaveLendingHistory(lendHist)
// 	if err != nil {
// 		t.Errorf("Error saving lending hist %s\n", err.Error())
// 	}

// 	_, err := usdb.GetLendHistorySummary("ted", lendHist.Time)
// 	if err != nil {
// 		t.Errorf("Error getting temp lending summary %s\n", err.Error())
// 	}
// }

// func Test_exchange_poloniex_db_create(t *testing.T) {
// 	db, err = CreateTestStatDB("mongodb://localhost:27017", "", "")
// 	if err != nil {
// 		t.Errorf("Error creating test db: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	if db == nil {
// 		t.Errorf("Error db is null")
// 		t.FailNow()
// 	}

// 	s, err := db.CreateSession()
// 	if err != nil {
// 		t.Errorf("Error creating session: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	defer s.Close()
// 	_, err = s.DB(STAT_DB_TEST).C(C_Exchange_POL).RemoveAll(bson.M{})
// 	if err != nil {
// 		t.Errorf("Error dropping removeAll: %s\n", err.Error())
// 		t.FailNow()
// 	}
// }

// func Test_exchange_poloniex_stat(t *testing.T) {
// 	usdb, err = userdb.NewUserStatisticsMongoDBGiven(db)
// 	if err != nil {
// 		t.Errorf("Error creating new stat mongodb: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	tempTime := time.Now().UTC() //5min
// 	err = usdb.RecordPoloniexStatisticTime("BTC", 0.005, tempTime)
// 	if err != nil {
// 		t.Errorf("Error adding statistic: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	tempTime = time.Now().UTC().Add(-6 * time.Minute) //hr
// 	err = usdb.RecordPoloniexStatisticTime("BTC", 0.005, tempTime)
// 	if err != nil {
// 		t.Errorf("Error adding statistic: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	tempTime = time.Now().UTC().Add(-2 * time.Hour) //day
// 	err = usdb.RecordPoloniexStatisticTime("BTC", 0.005, tempTime)
// 	if err != nil {
// 		t.Errorf("Error adding statistic: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	tempTime = time.Now().UTC().Add(-30 * time.Hour) //week
// 	err = usdb.RecordPoloniexStatisticTime("BTC", 0.005, tempTime)
// 	if err != nil {
// 		t.Errorf("Error adding statistic: %s\n", err.Error())
// 		t.FailNow()
// 	}
// 	tempTime = time.Now().UTC().Add(-8 * 24 * time.Hour) //month
// 	err = usdb.RecordPoloniexStatisticTime("BTC", 0.005, tempTime)
// 	if err != nil {
// 		t.Errorf("Error adding statistic: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	pol, err := usdb.GetPoloniexStatistics("BTC")
// 	if err != nil {
// 		t.Errorf("Error retrieving pol stats: %s\n", err.Error())
// 		t.FailNow()
// 	}

// 	if pol.FiveMinAvg != 0.005 {
// 		t.Errorf("5 Min average is incorrect: %f\n", pol.FiveMinAvg)
// 	}

// 	if pol.HrAvg != 0.005 {
// 		t.Errorf("Hour average is incorrect: %f\n", pol.HrAvg)
// 	}

// 	if pol.DayAvg != 0.005 {
// 		t.Errorf("Day average is incorrect: %f\n", pol.DayAvg)
// 	}

// 	if pol.WeekAvg != 0.005 {
// 		t.Errorf("Week average is incorrect: %f\n", pol.WeekAvg)
// 	}

// 	if pol.MonthAvg != 0.005 {
// 		t.Errorf("Month average is incorrect: %f\n", pol.MonthAvg)
// 	}
// }

// func Test_connect_to_remote_database(t *testing.T) {
// 	db, err := CreateTestUserDB("mongo.hodl.zone:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
// 	if err != nil {
// 		t.Error(err.Error())
// 		t.FailNow()
// 	}
// 	_, _, err = db.GetCollection(C_USER)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// }

// func Test_lendhist_call_time(t *testing.T) {
// 	ua, err := CreateTestStatDB("mongo.hodl.zone:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
// 	if err != nil {
// 		t.Error(err.Error())
// 		t.FailNow()
// 	}

// 	s, c, err := ua.GetCollection(C_LendHist)
// 	if err != nil {
// 		t.Errorf("createSession: %s", err.Error())
// 		t.FailNow()
// 	}
// 	n := time.Now().UTC()
// 	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
// 	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
// 	//CAN OPTIMIZE LATER
// 	qasdf := bson.M{
// 		"$and": []bson.M{
// 			bson.M{"email": "stevenmasley@gmail.com"},
// 			bson.M{"_id": top},
// 		},
// 	}
// 	m := bson.M{}
// 	err = c.Find(qasdf).Explain(m)
// 	if err != nil {
// 		t.Errorf("find: %s", err)
// 	} else {
// 		t.Logf("Explain: %#v\n", m)
// 	}
// 	s.Close()

// 	start := time.Now()
// 	for i := 0; i < 30; i++ {
// 		s, c, err := ua.GetCollection(C_LendHist)
// 		if err != nil {
// 			t.Errorf("createSession: %s", err.Error())
// 			t.FailNow()
// 		}

// 		query := bson.M{
// 			"$and": []bson.M{
// 				bson.M{"email": "stevenmasley@gmail.com"},
// 				bson.M{"_id": top},
// 			},
// 		}

// 		x := userdb.NewAllLendingHistoryEntry()
// 		err = c.Find(query).One(x)
// 		if err != nil {
// 			t.Errorf("find: %s", err)
// 		}
// 		s.Close()
// 		top = top.Add(-24 * time.Hour)
// 	}
// 	t.Logf("Took %fs", time.Since(start).Seconds())
// }

func Test_botactivity(t *testing.T) {
	nsNotFoundErr := errors.New("ns not found")

	givenUa, err := CreateTestStatDB("127.0.0.1:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	s, c, err := givenUa.GetCollection(C_BotActivity)
	if err != nil && err.Error() != nsNotFoundErr.Error() {
		t.Errorf("createSession: %s", err.Error())
		t.FailNow()
	}
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Error dropping collection: %s", err.Error())
	}
	s.Close()

	ua, err := userdb.NewUserStatisticsMongoDBGiven(givenUa)
	if err != nil {
		t.Errorf("Error creating new stat mongodb: %s\n", err.Error())
		t.FailNow()
	}

	now := time.Now().UTC()
	bals := make([]userdb.BotActivityLogEntry, 99, 99)
	for i := 0; i < 99; i++ {
		bals[i] = userdb.BotActivityLogEntry{
			fmt.Sprintf("%d", i),
			now.Add(time.Duration(-i) * time.Minute),
		}
	}

	err = ua.AddBotActivityLogEntry("test", &bals)
	if err != nil {
		t.Errorf("Error adding bot activity: %s\n", err.Error())
	}

	b, err := ua.GetBotActivity("test")
	if err != nil {
		t.Errorf("Error getting bot activity: %s\n", err.Error())
	}
	if len(*b.ActivityLog) != 99 {
		t.Errorf("Error incorrect number of activity logs: %d should be 99", len(*b.ActivityLog))
	}

	bals = make([]userdb.BotActivityLogEntry, 2, 2)
	for i := 0; i < 2; i++ {
		bals[i] = userdb.BotActivityLogEntry{
			fmt.Sprintf("new %d", i),
			now.Add(time.Duration(1+i) * time.Minute),
		}
	}

	err = ua.AddBotActivityLogEntry("test", &bals)
	if err != nil {
		t.Errorf("Error adding bot activity: %s\n", err.Error())
	}
	b, err = ua.GetBotActivity("test")
	if err != nil {
		t.Errorf("Error getting bot activity again: %s\n", err.Error())
	}
	if len(*b.ActivityLog) != 100 {
		t.Errorf("Error incorrect number of activity logs: %d should be 100", len(*b.ActivityLog))
		t.FailNow()
	}

	if (*b.ActivityLog)[0].Log != "new 1" {
		t.Errorf("0 Error with log='%s' time='%s'\n", (*b.ActivityLog)[0].Log, (*b.ActivityLog)[0].Time)
	}
	if (*b.ActivityLog)[1].Log != "new 0" {
		t.Errorf("1 Error with log='%s' time='%s'\n", (*b.ActivityLog)[1].Log, (*b.ActivityLog)[1].Time)
	}
	if (*b.ActivityLog)[2].Log != "0" {
		t.Errorf("2 Error with log='%s' time='%s'\n", (*b.ActivityLog)[2].Log, (*b.ActivityLog)[2].Time)
	}
	if (*b.ActivityLog)[99].Log != "97" {
		t.Errorf("99 Error with log='%s' time='%s'\n", (*b.ActivityLog)[3].Log, (*b.ActivityLog)[3].Time)
	}

	balsV2, err := ua.GetBotActivityTimeGreater("test", now)
	if err != nil {
		t.Errorf("Error getting time greater: %s\n", err.Error())
	}

	if len(*balsV2) < 2 {
		t.Errorf("Error length of balls should be 2 is %d\n", len(*balsV2))
		t.FailNow()
	}

	if (*balsV2)[0].Log != "new 1" {
		t.Errorf("GetBoatActivityTimeGreater error with log='%s' time='%s'\n", (*balsV2)[0].Log, (*balsV2)[0].Time)
	}
	if (*balsV2)[1].Log != "new 0" {
		t.Errorf("GetBoatActivityTimeGreater error with log='%s' time='%s'\n", (*balsV2)[0].Log, (*balsV2)[0].Time)
	}
}

var _ = errors.New

func Test_user_session(t *testing.T) {
	nsNotFoundErr := errors.New("ns not found")
	db, err = CreateTestUserDB("127.0.0.1:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	s, c, err := db.GetCollection(C_Session)
	if err != nil {
		t.Errorf("createSession: %s", err.Error())
		t.FailNow()
	}
	err = c.DropCollection()
	if err != nil && err.Error() != nsNotFoundErr.Error() {
		t.Errorf("Error dropping collection: %s", err.Error())
	}
	s.Close()

	db, err = CreateTestUserDB("127.0.0.1:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	usdb := userdb.NewMongoUserDatabaseGiven(db)

	testSessionId := "apples"
	testEmail := "test"
	testIp := net.ParseIP("216.14.49.184")
	testTime := time.Now().UTC()
	err = usdb.UpdateUserSession(testSessionId, testEmail, testTime, testIp, true)
	if err != nil {
		t.Errorf("Error updating user session: %s", err.Error())
	}

	//get one session that is open
	allSessions, err := usdb.GetAllUserSessions(testEmail, 0, 100)
	if err != nil {
		t.Errorf("Error getting all user sessions: %s", err.Error())
	}
	if len(*allSessions) != 1 {
		t.Errorf("Error with length of all sessions should be 1 is %d", len(*allSessions))
		t.FailNow()
	}

	sessionIP := userdb.SessionIP{testIp, testTime}
	sesRetArr := []userdb.SessionIP{sessionIP}
	sessionSt := userdb.SessionState{userdb.OPENED, testTime}
	sesStateRetArr := []userdb.SessionState{sessionSt}
	sesRet := userdb.Session{testSessionId, testEmail, testTime, testIp, true, sesRetArr, sesStateRetArr}
	if (*allSessions)[0].IsSameAs(&sesRet) == false {
		t.Error("Error sessions not equal: ", JsonPrettyHelper((*allSessions)[0]), JsonPrettyHelper(sesRet))
	}

	//add another open session
	test2SessionId := "pears"
	test2Email := "test"
	test2Ip := net.ParseIP("216.14.49.185")
	test2Time := time.Now().UTC()
	err = usdb.UpdateUserSession(test2SessionId, test2Email, test2Time, test2Ip, true)
	if err != nil {
		t.Errorf("Error updating user2 session: %s", err.Error())
	}
	allSessions, err = usdb.GetAllUserSessions(testEmail, 0, 100)
	if err != nil {
		t.Errorf("Error getting all user sessions2: %s", err.Error())
	}
	if len(*allSessions) != 2 {
		t.Errorf("Error with length of all sessions should be 2 is %d", len(*allSessions))
		t.FailNow()
	}

	sessionIP = userdb.SessionIP{test2Ip, test2Time}
	sesRetArr2 := []userdb.SessionIP{sessionIP}
	sessionSt = userdb.SessionState{userdb.OPENED, test2Time}
	sesStateRetArr2 := []userdb.SessionState{sessionSt}
	sesRet2 := userdb.Session{test2SessionId, test2Email, test2Time, test2Ip, true, sesRetArr2, sesStateRetArr2}
	if (*allSessions)[1].IsSameAs(&sesRet) == false {
		t.Error("Error sessions 2 not equal: ", JsonPrettyHelper((*allSessions)[1]), JsonPrettyHelper(sesRet))
	}
	if (*allSessions)[0].IsSameAs(&sesRet2) == false {
		t.Error("Error sessions 3 not equal: ", "\n", JsonPrettyHelper((*allSessions)[0]), "\n", JsonPrettyHelper(sesRet2))
	}

	//increment one and set to off
	err = usdb.UpdateUserSession(testSessionId, testEmail, testTime, testIp, false)
	if err != nil {
		t.Errorf("Error updating user3 session: %s", err.Error())
	}

	allSessions, err = usdb.GetAllUserSessions(testEmail, 2, 100)
	if err != nil {
		t.Errorf("Error getting all user sessions3: %s", err.Error())
	}
	if len(*allSessions) != 1 {
		t.Errorf("Error with length of all sessions 1 should be 1 is %d", len(*allSessions))
		t.FailNow()
	}
	sesRet.Open = false
	sesRet.ChangeState = append(sesRet.ChangeState, userdb.SessionState{userdb.CLOSED, testTime})
	if (*allSessions)[0].IsSameAs(&sesRet) == false {
		t.Error("Error sessions 4 not equal: ", "\n", JsonPrettyHelper((*allSessions)[0]), "\n", JsonPrettyHelper(sesRet))
	}

	allSessions, err = usdb.GetAllUserSessions(testEmail, 1, 100)
	if err != nil {
		t.Errorf("Error getting all user sessions3: %s", err.Error())
	}
	if len(*allSessions) != 1 {
		t.Errorf("Error with length of all sessions 2 should be 1 is %d", len(*allSessions))
		t.FailNow()
	}
	if (*allSessions)[0].IsSameAs(&sesRet2) == false {
		t.Error("Error sessions 5 not equal: ", JsonPrettyHelper((*allSessions)[0]), JsonPrettyHelper(sesRet2))
	}

	rTime := time.Now().UTC()
	//test update renewal time
	err = usdb.UpdateUserSession(test2SessionId, test2Email, rTime, test2Ip, true)
	if err != nil {
		t.Errorf("Error updating user4 session: %s", err.Error())
	}
	allSessions, err = usdb.GetAllUserSessions(testEmail, 1, 100)
	if err != nil {
		t.Errorf("Error getting all user sessions4: %s", err.Error())
	}
	if len(*allSessions) != 1 {
		t.Errorf("Error with length of all sessions 3 should be 1 is %d", len(*allSessions))
		t.FailNow()
	}
	sesRet2.LastRenewalTime = rTime
	if (*allSessions)[0].IsSameAs(&sesRet2) == false {
		t.Error("Error sessions 6 not equal: ", "\n", JsonPrettyHelper((*allSessions)[0]), "\n", JsonPrettyHelper(sesRet2))
	}
}

func JsonPrettyHelper(i userdb.Session) string {
	b, _ := json.Marshal(i)
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	return string(out.Bytes())
}

func JsonPrettyHelperArr(i []userdb.Session) string {
	b, _ := json.Marshal(i)
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	return string(out.Bytes())
}
