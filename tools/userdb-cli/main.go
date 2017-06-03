package main

// Usage
//		userdb-cli -u USERNAME -l admin

import (
	"flag"
	"fmt"
	"os"

	"github.com/Emyrk/LendingBot/src/core/userdb"
)

func main() {
	var (
		username = flag.String("u", "", "Username to change level of")
		level    = flag.String("l", "", "Level to set user, 'admin', 'sysadmin', user")
		auth     = flag.String("a", "", "2fa auth")
		listall  = flag.Bool("la", false, "List all users")
	)

	flag.Parse()
	la := *listall
	if *username == "" && !la {
		panic("No username chosen")
	}

	db := userdb.NewBoltUserDatabase("UserDatabase.db")
	if db == nil {
		panic("DB not opened")
	}

	if *listall {
		users, err := db.FetchAllUsers()
		if err != nil {
			fmt.Printf("Error when loading users: %s\n", err.Error())
			panic(err)
		}

		for _, u := range users {
			fmt.Println(u.Username)
		}
		return
	}

	u, err := db.FetchUserIfFound(*username)
	if err != nil {
		fmt.Printf("Error when loading user: %s\n", *username)
		panic(err)
	}

	fmt.Println("-- User Found --")
	fmt.Println(u)

	if *auth != "" {
		err := u.Validate2FA(*auth)
		if err != nil {
			p, _ := u.User2FA.OTP()
			fmt.Println("Error:", err.Error())
			fmt.Printf("Should be: %s\n", p)
			f, _ := os.OpenFile("qr.png", os.O_CREATE|os.O_RDWR, 0777)
			b, _ := u.User2FA.QR()
			f.Write(b)
			f.Close()
		} else {
			fmt.Println("Successfully authenticated via 2fa!")
		}
		return
	}

	switch *level {
	case "admin":
		db.SetUserLevel(*username, userdb.Admin)
	case "sysadmin":
		db.SetUserLevel(*username, userdb.SysAdmin)
	case "user":
		db.SetUserLevel(*username, userdb.CommonUser)
	default:
		fmt.Println("No level detected. Expect: 'sysadmin', admin', or 'user'")
		return
	}

	u, err = db.FetchUserIfFound(*username)
	if err != nil {
		fmt.Printf("Error when loading user: %s\n", *username)
		panic(err)
	}
	fmt.Println("-- User Is Now --")
	fmt.Println(u)
}
