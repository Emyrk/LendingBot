package main

// Usage
//		userdb-cli -u USERNAME -l admin

import (
	"flag"
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

func main() {
	var (
		username = flag.String("u", "", "Username to change level of")
		level    = flag.String("l", "", "Level to set user, 'admin', 'sysadmin'")
	)

	flag.Parse()
	if *username == "" {
		panic("No username chosen")
	}
	fmt.Println("Asd")

	db := userdb.NewBoltUserDatabase("UserDatabase.db")
	if db == nil {
		panic("DB not opened")
	}

	u, err := db.FetchUserIfFound(*username)
	if err != nil {
		fmt.Printf("Error when loading user: %s\n", *username)
		panic(err)
	}

	fmt.Println("-- User Found --")
	fmt.Println(u)

	if *level == "admin" {
		db.SetUserLevel(*username, userdb.Admin)
	} else if *level == "sysadmin" {
		db.SetUserLevel(*username, userdb.SysAdmin)
	} else {
		panic("No level detected")
	}

	u, err = db.FetchUserIfFound(*username)
	if err != nil {
		fmt.Printf("Error when loading user: %s\n", *username)
		panic(err)
	}
	fmt.Println("-- User Is Now --")
	fmt.Println(u)
}
