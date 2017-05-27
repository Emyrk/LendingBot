package core

import (
	"github.com/DistributedSolutions/LendingBot/app/core/userdb"
)

type State struct {
	UserDB *userdb.UserDatabase
}

func NewState() *State {
	s := new(State)
	s.UserDB = userdb.NewMapUserDatabase()

	return s
}

/*
func (ud *UserDatabase) PutUser(u *User) error {

func (ud *UserDatabase) FetchUser(username string) (*User, error) {

func (ud *UserDatabase) AuthenticateUser(username string, password string) (bool, *User, error) {



*/

func (s *State) NewUser() {

}

func (s *State) FetchUser(username string) (*userdb.User, error) {
	return s.UserDB.FetchUser(username)
}

func (s *State) AuthenticateUser(username string, password string) (bool, *userdb.User, error) {
	return s.UserDB.AuthenticateUser(username, password)
}
