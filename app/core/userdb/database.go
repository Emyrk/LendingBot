package userdb

import (
	"github.com/DistributedSolutions/LendingBot/app/database"
)

type BinaryMarshalable interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) (err error)
	UnmarshalBinaryData(data []byte) (newData []byte, err error)
}

// Buckets
var (
	UsersBucket []byte = []byte("UserByHash")
)

type UserDatabase struct {
	db database.IDatabase
}

func NewMapUserDatabase() *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewMapDB()

	return u
}

func (ud *UserDatabase) PutUser(u *User) error {
	return ud.put(UsersBucket, GetUsernameHash(u.Username)[:], u)
}

func (ud *UserDatabase) FetchUser(username string) (*User, error) {
	u := new(User)
	err := ud.get(UsersBucket, GetUsernameHash(username)[:], u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ud *UserDatabase) AuthenticateUser(username string, password string) (bool, *User, error) {
	u, err := ud.FetchUser(username)
	if err != nil {
		return false, nil, err
	}

	if u.Authenticate(password) {
		return true, u, nil
	}

	return false, nil, nil
}

func (ud *UserDatabase) put(bucket []byte, key []byte, obj BinaryMarshalable) error {
	data, err := obj.MarshalBinary()
	if err != nil {
		return err
	}

	return ud.db.Put(bucket, key, data)
}

func (ud *UserDatabase) get(bucket []byte, key []byte, obj BinaryMarshalable) error {
	data, err := ud.db.Get(bucket, key)
	if err != nil {
		return err
	}

	err = obj.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	return nil
}
