package userdb

import (
	"fmt"
	"time"

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

func (ud *UserDatabase) AuthenticateUser(username string, password string, token string) (bool, *User, error) {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return false, nil, err
	}

	if !u.AuthenticatePassword(password) {
		return false, nil, nil
	}

	// Passed Password Auth
	if u.Enabled2FA {
		err = ud.validate2FA(u, token)
		if err != nil {
			return false, nil, err
		}
	}

	return false, nil, nil
}

func (ud *UserDatabase) Add2FA(username string, password string) (qr []byte, err error) {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return nil, err
	}

	if u.Enabled2FA {
		return nil, fmt.Errorf("2FA is already enabled. Disable to generate a new barcode")
	}

	if !u.AuthenticatePassword(password) {
		return nil, fmt.Errorf("Invalid password")
	}

	qr, err = u.Create2FA()
	if err != nil {
		return nil, err
	}

	err = ud.PutUser(u)
	if err != nil {
		return nil, err
	}

	return qr, nil
}

func (ud *UserDatabase) Enable2FA(username string, password string, token string, enabled bool) error {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	if !u.AuthenticatePassword(password) {
		return fmt.Errorf("Invalid password or 2FA")
	}

	valid := ud.validate2FA(u, token)
	if valid != nil {
		return err
	}

	u.Enabled2FA = enabled
	return ud.PutUser(u)
}

func (ud *UserDatabase) validate2FA(u *User, token string) error {
	return u.Validate2FA(token)
}

func (ud *UserDatabase) UpdateJWTTime(username string, t time.Time) error {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.JWTTime = t

	return ud.PutUser(u)
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
