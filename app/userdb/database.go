package userdb

import (
	"github.com/DistributedSolutions/LendingBot/app/database"
)

type UserDatabase struct {
	db database.IDatabase
}

func NewMapUserDatabase() *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewMapDB()

	return u
}
