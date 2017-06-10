package userdb

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/database"
)

var (
	InviteCodeBucket []byte = []byte("Invite_Code_Bucket")
)

type InviteDB struct {
	db database.IDatabase
}

type InviteEntry struct {
	RawCode string `json:"rawc"`

	// Is a hash. Allows for user-readable codes
	Code [32]byte `json:"c"`

	// How many times it can be used
	Capacity int `json:"cap"`

	// Usernames who claimed
	Users []UserEntry `json:"users"`

	// Invite code will expire at this time
	Expires time.Time `json:"time"`
}

func (ie *InviteEntry) String() string {
	return fmt.Sprintf("Code: %s, Capacity %d, Claimed: %d, Expires %s\n", ie.RawCode, ie.Capacity, len(ie.Users), ie.Expires.String())
}

func NewInviteDB() *InviteDB {
	return newInviteDB(false)
}

func NewInviteMapDB() *InviteDB {
	return newInviteDB(true)
}

func newInviteDB(mapDB bool) *InviteDB {
	i := new(InviteDB)
	boltpath := os.Getenv("INVITE_DB")
	if boltpath == "" {
		boltpath = "InviteUserDB.db"
	}

	if mapDB {
		i.db = database.NewMapDB()
	} else {
		i.db = database.NewBoltDB(boltpath)
	}

	return i
}

func (ie *InviteDB) ListAll() ([]InviteEntry, error) {
	data, _, err := ie.db.GetAll(InviteCodeBucket)
	if err != nil {
		return nil, err
	}
	var list []InviteEntry

	for _, d := range data {
		var i InviteEntry
		err := i.UnmarshalBinary(d)
		if err != nil {
			fmt.Println(err)
			continue
		}
		list = append(list, i)
	}

	return list, nil
}

func (ie *InviteDB) DeleteInvite(hash string) error {
	//do not delete the code, just make the claim date in the past.
	return nil
}

func (ie *InviteDB) ClaimInviteCode(username string, code string) (bool, error) {
	key := sha256.Sum256([]byte(code))
	data, err := ie.db.Get(InviteCodeBucket, key[:])
	if err != nil {
		return false, err
	}

	if data == nil {
		return false, fmt.Errorf("not a valid invite code")
	}

	i := new(InviteEntry)
	err = i.UnmarshalBinary(data)
	if err != nil {
		return false, err
	}

	if len(i.Users) >= i.Capacity {
		return false, fmt.Errorf("invite code has been used up")
	}

	if i.Expires.Before(time.Now()) {
		return false, fmt.Errorf("invite code has expired")
	}

	i.Users = append(i.Users, UserEntry{Username: username, ClaimTime: time.Now()})
	return true, ie.putInviteCode(i)
}

func (ie *InviteDB) CreateInviteCode(code string, capacity int, expires time.Time) error {
	if v, _ := ie.getRawInviteCode(code); v != nil {
		return fmt.Errorf("Code %s already exists", code)
	}
	i := NewInviteCode(code, capacity, expires)
	return ie.putInviteCode(i)
}

func (ie *InviteDB) ExpireInviteCode(code string) error {
	c, err := ie.getInviteCode(code)
	if err != nil {
		return err
	}

	c.Expires = time.Now().Add(-1 * time.Hour)
	return ie.putInviteCode(c)
}

func (ie *InviteDB) putInviteCode(ic *InviteEntry) error {
	data, err := ic.MarshalBinary()
	if err != nil {
		return err
	}

	return ie.db.Put(InviteCodeBucket, ic.Code[:], data)
}

func (ie *InviteDB) getInviteCode(raw string) (*InviteEntry, error) {
	rawData, err := ie.getRawInviteCode(raw)
	if err != nil {
		return nil, err
	}

	e := new(InviteEntry)
	err = e.UnmarshalBinary(rawData)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (ie *InviteDB) getRawInviteCode(raw string) ([]byte, error) {
	key := sha256.Sum256([]byte(raw))
	return ie.db.Get(InviteCodeBucket, key[:])
}

func NewInviteCode(code string, capacity int, expires time.Time) *InviteEntry {
	i := new(InviteEntry)
	i.Code = sha256.Sum256([]byte(code))
	i.Capacity = capacity
	i.Expires = expires
	i.RawCode = code

	return i
}

func (ie *InviteEntry) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	data, err := primitives.MarshalStringToBytes(ie.RawCode, 100)
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.Write(ie.Code[:])

	buf.Write(primitives.Uint32ToBytes(uint32(ie.Capacity)))

	buf.Write(primitives.Uint32ToBytes(uint32(len(ie.Users))))

	for _, u := range ie.Users {
		data, err := u.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	data, err = ie.Expires.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Next(buf.Len()), nil
}

func (ie *InviteEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = ie.UnmarshalBinaryData(data)
	return
}

func (ie *InviteEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[UserEntry] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data

	var raw string
	raw, newData, err = primitives.UnmarshalStringFromBytesData(newData, 100)
	if err != nil {
		return data, err
	}
	ie.RawCode = raw

	copy(ie.Code[:32], newData[:32])
	newData = newData[32:]

	u, err := primitives.BytesToUint32(newData[:4])
	if err != nil {
		return data, err
	}
	newData = newData[4:]
	ie.Capacity = int(u)

	l, err := primitives.BytesToUint32(newData[:4])
	if err != nil {
		return data, err
	}
	newData = newData[4:]

	ie.Users = make([]UserEntry, l)
	for i := 0; i < int(l); i++ {
		newData, err = ie.Users[i].UnmarshalBinaryData(newData)
		if err != nil {
			return data, err
		}
	}

	ts := newData[:15]
	err = ie.Expires.UnmarshalBinary(ts)
	if err != nil {
		return data, err
	}

	newData = newData[15:]
	return
}

type UserEntry struct {
	Username  string
	ClaimTime time.Time
}

func NewUserEntry(username string) *UserEntry {
	u := new(UserEntry)
	u.Username = username
	u.ClaimTime = time.Now()

	return u
}

func (u *UserEntry) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	data, err := primitives.MarshalStringToBytes(u.Username, 100)
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = u.ClaimTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Next(buf.Len()), nil
}

func (u *UserEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = u.UnmarshalBinaryData(data)
	return
}

func (u *UserEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[UserEntry] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data

	var un string
	un, newData, err = primitives.UnmarshalStringFromBytesData(newData, 100)
	if err != nil {
		return data, err
	}
	u.Username = un

	td := newData[:15]
	err = u.ClaimTime.UnmarshalBinary(td)
	if err != nil {
		return data, err
	}
	newData = newData[15:]
	return
}
