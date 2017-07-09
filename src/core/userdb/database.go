package userdb

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/database"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	db  database.IDatabase
	mdb *mongo.MongoDB
}

func NewMapUserDatabase() *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewMapDB()
	u.mdb = nil

	return u
}

func NewBoltUserDatabase(path string) *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewBoltDB(path)
	u.mdb = nil

	return u
}

func NewMongoUserDatabase(uri string, dbu string, dbp string) (*UserDatabase, error) {

	mdb, err := mongo.CreateUserDB(uri, dbu, dbp)
	if err != nil {
		return nil, err
	}
	u := NewMongoUserDatabaseGiven(mdb)

	return u, nil
}

func NewMongoUserDatabaseGiven(mdb *mongo.MongoDB) *UserDatabase {
	u := new(UserDatabase)
	u.db = database.NewMapDB()
	u.mdb = mdb

	return u
}

func (ud *UserDatabase) Close() error {
	if ud.mdb == nil {
		return ud.db.Close()
	}
	return nil
}

func (ud *UserDatabase) PutUser(u *User) error {
	if ud.mdb != nil {
		s, c, err := ud.mdb.GetCollection(mongo.C_USER)
		if err != nil {
			return fmt.Errorf("PutUser: getCol: %s", err.Error())
		}
		defer s.Close()

		upsertAction := bson.M{"$set": u}
		_, err = c.UpsertId(u.Username, upsertAction)
		if err != nil {
			return fmt.Errorf("PutUser: upsert: %s", err.Error())
		}
		return nil
	}

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
	if ud.mdb != nil {
		s, c, err := ud.mdb.GetCollection(mongo.C_USER)
		if err != nil {
			return nil, fmt.Errorf("PutUser: getCol: %s", err.Error())
		}
		defer s.Close()

		var result User
		err = c.FindId(username).One(&result)
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("PutUser: find: %s", err.Error())
		}
		result.PoloniexKeys.SetEmptyIfBlank()
		return &result, nil
	}

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

	if ud.mdb != nil {
		s, c, err := ud.mdb.GetCollection(mongo.C_USER)
		if err != nil {
			return nil, fmt.Errorf("FetchAllUsers: getCol: %s", err.Error())
		}
		defer s.Close()

		err = c.Find(nil).All(&all)
		if err != nil {
			return nil, fmt.Errorf("FetchAllUsers: find: %s", err.Error())
		}
		for _, o := range all {
			o.PoloniexKeys.SetEmptyIfBlank()
		}

		return all, nil
	}

	keys, err := ud.db.ListAllKeys(UsersBucket)
	if err != nil {
		return all, err
	}

	for _, k := range keys {
		u := NewBlankUser()

		data, err := ud.db.Get(UsersBucket, k)
		if err != nil {
			continue
		}

		if data == nil {
			continue
		}

		fmt.Printf("%s: %x", k, data)

		err = u.SafeUnmarshal(data)
		if err != nil {
			continue
		}
		all = append(all, *u)
		// f, err := ud.get(UsersBucket, k, u)
		// if f && err == nil {
		// 	all = append(all, *u)
		// 	continue
		// }
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
