package userdb

import (
	"fmt"

	"github.com/Emyrk/LendingBot/app/core/database"
)

var _ = fmt.Println

type BinaryMarshalable interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) (err error)
	UnmarshalBinaryData(data []byte) (newData []byte, err error)
}

// Buckets
var (
	UsersBucket  []byte = []byte("UserByHash")
	VerifyBucket []byte = []byte("VerifyEmails")
)

type UserDatabase struct {
	db database.IDatabase
}

func NewMapUserDatabase() *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewMapDB()

	return u
}

func NewBoltUserDatabase(path string) *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewBoltDB(path)

	return u
}

func (ud *UserDatabase) Close() error {
	return ud.db.Close()
}

func (ud *UserDatabase) PutUser(u *User) error {
	hash := GetUsernameHash(u.Username)
	return ud.put(UsersBucket, hash[:], u)
}

func (ud *UserDatabase) FetchUserIfFound(username string) (*User, error) {
	u, err := ud.FetchUser(username)
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, fmt.Errorf("Not found")
	}
	return u, nil
}

func (ud *UserDatabase) FetchUser(username string) (*User, error) {
	u := NewBlankUser()
	hash := GetUsernameHash(username)
	f, err := ud.get(UsersBucket, hash[:], u)
	if err != nil {
		return nil, err
	}

	if !f {
		return nil, nil
	}

	return u, nil
}

func (ud *UserDatabase) FetchAllUsers() ([]User, error) {
	var all []User

	keys, err := ud.db.ListAllKeys(UsersBucket)
	if err != nil {
		return all, err
	}

	for _, k := range keys {
		u := NewBlankUser()
		f, err := ud.get(UsersBucket, k, u)
		if f && err == nil {
			all = append(all, *u)
			continue
		}
	}

	return all, nil
}

func (ud *UserDatabase) SetUserLevel(username string, level UserLevel) error {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.Level = level

	return ud.PutUser(u)
}

func (ud *UserDatabase) put(bucket []byte, key []byte, obj BinaryMarshalable) error {
	data, err := obj.MarshalBinary()
	if err != nil {
		return err
	}

	return ud.db.Put(bucket, key, data)
}

func (ud *UserDatabase) get(bucket []byte, key []byte, obj BinaryMarshalable) (found bool, err error) {
	data, err := ud.db.Get(bucket, key)
	if err != nil {
		return false, err
	}

	if data == nil || len(data) == 0 {
		return false, nil
	}

	err = obj.UnmarshalBinary(data)
	if err != nil {
		return true, err
	}
	return true, nil
}
