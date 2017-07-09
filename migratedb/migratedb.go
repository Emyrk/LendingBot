package main

import (
	"fmt"
	"os"

	"github.com/Emyrk/LendingBot/src/core/userdb"
)

type UserMigrateDB struct {
	userMongoDB    *userdb.UserDatabase
	userEmbeddedDB *userdb.UserDatabase
}

type UserStatMigrateDB struct {
	userStatMongoDBBalacner *userdb.UserStatisticsDB
	userStatMongoDBBee      *userdb.UserStatisticsDB
	userStatEmbeddedDB      *userdb.UserStatisticsDB
}

func SetUpUserDB() *UserMigrateDB {
	var err error
	userMigrateDB := new(UserMigrateDB)

	uri := "mongo.hodl.zone:27017"
	mongoRevelPass := os.Getenv("MONGO_REVEL_PASS")
	if mongoRevelPass == "" {
		panic("Running in prod, but no revel pass given in env var 'MONGO_REVEL_PASS'")
	}
	userMigrateDB.userMongoDB, err = userdb.NewMongoUserDatabase(uri, "revel", mongoRevelPass)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to user mongodb: %s\n", err.Error()))
	}

	v := os.Getenv("USER_DB")
	if len(v) == 0 {
		v = "UserDatabase.db"
	}
	userMigrateDB.userEmbeddedDB = userdb.NewBoltUserDatabase(v)
	return userMigrateDB
}

func SetUpUserStatMigrateDB() *UserStatMigrateDB {
	var err error
	userStatMigrateDB := new(UserStatMigrateDB)

	uri := "mongo.hodl.zone:27017"
	mongoBalancerPass := os.Getenv("MONGO_BAL_PASS")
	if mongoBalancerPass == "" {
		panic("Running in prod, but no revel pass given in env var 'MONGO_BAL_PASS'")
	}
	userStatMigrateDB.userStatMongoDBBalacner, err = userdb.NewUserStatisticsMongoDB(uri, "balancer", mongoBalancerPass)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to userstat balancer mongodb: %s\n", err.Error()))
	}
	mongoBeePass := os.Getenv("MONGO_BEE_PASS")
	if mongoBeePass == "" {
		panic("Running in prod, but no revel pass given in env var 'MONGO_BEE_PASS'")
	}
	userStatMigrateDB.userStatMongoDBBee, err = userdb.NewUserStatisticsMongoDB(uri, "bee", mongoBeePass)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to userstat bee mongodb: %s\n", err.Error()))
	}

	v := os.Getenv("USER_DB")
	if len(v) == 0 {
		v = "UserDatabase.db"
	}
	userStatMigrateDB.userStatEmbeddedDB, err = userdb.NewUserStatisticsDB()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to user embedded: %s\n", err.Error()))
	}
	return userStatMigrateDB
}

func main() {
	// fmt.Printf("---------STARTED MIRGATE USER DB---------\n")
	// userMigrateDB := SetUpUserDB()

	// users, err := userMigrateDB.userEmbeddedDB.FetchAllUsers()
	// if err != nil {
	// 	panic(fmt.Sprintf("Error retrieving users: %s\n", err.Error()))
	// } else {
	// 	fmt.Printf("Successfully retrieved %d users\n", len(users))
	// 	for i, _ := range users {
	// 		err = userMigrateDB.userMongoDB.PutUser(&users[i])
	// 		if err != nil {
	// 			fmt.Printf("ERROR: adding user: %s\n", users[i].Username)
	// 		} else {
	// 			fmt.Printf("Success: adding user: %s\n", users[i].Username)
	// 		}
	// 	}
	// }
	// fmt.Printf("---------FINISHED MIRGATE USER DB---------\n\n")

	fmt.Printf("---------STARTED MIRGATE USERSTATS DB---------\n")
	userStatMigrateDB := SetUpUserStatMigrateDB()
	if err != nil {
		panic(fmt.Sprintf("Error retrieving userstat: %s\n", err.Error()))
	} else {
		// lendingHist, err := userStatMigrateDB.userStatEmbeddedDB.GetLendHistorySummary(username, time.Now().UTC())
		// if err != nil {
		// 	fmt.Printf("ERROR: adding user: %s\n", u.Username)
		// }
		// fmt.Printf("---------START MIRGATE LENDING HISTORY---------\n")
		// for _, u := range users {
		// 	n := time.Now().UTC()
		// 	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
		// 	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
		// 	// Must start 2 days back to ensure all loans covered
		// 	top = top.Add(-24 * time.Hour)
		// 	curr := top.Add(time.Hour * -72).Add(1 * time.Second)
		// 	for i := 0; i < 28; i++ {
		// 		v, err := userStatMigrateDB.userStatEmbeddedDB.GetLendHistorySummary(u.Username, top)
		// 		if err != nil {
		// 			fmt.Printf("ERROR: retrieving lending history for day[%s]for user: %s\n", top.String(), u.Username)
		// 		} else if v != nil {
		// 			err := userStatMigrateDB.userStatMongoDBBalacner.SaveLendingHistory(v)
		// 			if err != nil {
		// 				fmt.Printf("ERROR: saving lending history for day[%s] for user: %s\n", top.String(), u.Username)
		// 			}
		// 		}

		// 		top = top.Add(-24 * time.Hour)
		// 		curr = curr.Add(-24 * time.Hour)
		// 	}
		// 	fmt.Printf("Success: Adding user lending history: %s\n", u.Username)
		// }

		// fmt.Printf("---------FINISHED MIRGATE LENDING HISTORY---------\n")
		// fmt.Printf("---------START MIRGATE EXCHANGE---------\n")

		// poloCoins := []string{"BTC", "BTS", "CLAM", "DOGE", "DASH", "LTC", "MAID", "STR", "XMR", "XRP", "ETH", "FCT"}
		// for _, coin := range poloCoins {
		// 	psArr, err := userStatMigrateDB.userStatEmbeddedDB.GetAllPoloniexStatistics(coin)
		// 	if err != nil {
		// 		fmt.Printf("ERROR: retrieving polo stats")
		// 	} else {
		// 		fmt.Printf("Exchange info found: %d\n", len(*psArr))
		// 		for _, ps := range *psArr {
		// 			err = userStatMigrateDB.userStatMongoDBBalacner.RecordPoloniexStatisticTime(coin, ps.Rate, ps.Time)
		// 			if err != nil {
		// 				fmt.Printf("ERROR: saving poloniex stats: %s\n", err.Error())
		// 			}
		// 		}
		// 	}
		// }

		// fmt.Printf("---------FINISHED MIRGATE EXCHANGE---------\n")
		fmt.Printf("---------START MIRGATE USERSTATS---------\n")

		for _, u := range users {
			for i := 0; i < 29; i++ {
				stats := userStatMigrateDB.userStatEmbeddedDB.GetStatisticsOneDay(u.Username, i)
				fmt.Printf("userstat info found for [%s] Day [%d]: %d\n", u.Username, i, len(stats))
				for i, _ := range stats {
					err = userStatMigrateDB.userStatMongoDBBee.RecordData(&stats[i])
					if err != nil {
						fmt.Printf("Error saving user %s userStat: %s\n", u.Username, err.Error())
					}
				}
			}
		}
		fmt.Printf("---------FINISHED MIRGATE USERSTATS---------\n")

	}
	fmt.Printf("---------FINISHED MIRGATE USERSTATS DB---------\n\n")
}
