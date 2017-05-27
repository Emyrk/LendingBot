package userdb

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core/common/primitives"
)

const UsernameMaxLength int = 100
const SaltLength int = 5

type User struct {
	Username     string // Not case sensitive
	PasswordHash [32]byte
	Salt         []byte

	StartTime  time.Time
	MiniumLend float64

	PoloniexKeys *PoloniexKeys
}

// filterUsername returns false if illegal characters
func filterUsername(username string) error {
	if len(username) > 100 {
		return fmt.Errorf("Username length is too long. Must be under %d, inputed length is %d", UsernameMaxLength, len(username))
	}
	return nil
}

func GetUsernameHash(username string) [32]byte {
	return sha256.Sum256([]byte(strings.ToLower(username)))
}

func NewUser(username string, password string) (*User, error) {
	u := new(User)

	if err := filterUsername(username); err != nil {
		return nil, err
	}

	u.Username = username
	u.Salt = make([]byte, SaltLength)
	_, err := rand.Read(u.Salt)
	if err != nil {
		return nil, err
	}

	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	u.PasswordHash = hash

	u.PoloniexKeys = new(PoloniexKeys)
	u.StartTime = time.Now()
	return u, nil
}

func (u *User) Authenticate(password string) bool {
	hash := u.getPasswordHashFromPassword(password)
	if bytes.Compare(u.PasswordHash[:], hash[:]) == 0 {
		return true
	}
	return false
}

func (u *User) getPasswordHashFromPassword(password string) [32]byte {
	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	return hash
}

func (u *User) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	b, err := primitives.MarshalStringToBytes(u.Username, UsernameMaxLength)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	buf.Write(u.PasswordHash[:])
	buf.Write(u.Salt[:])

	b, err = u.PoloniexKeys.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = u.StartTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	err = binary.Write(buf, binary.BigEndian, u.MiniumLend)
	if err != nil {
		return nil, err
	}

	return buf.Next(buf.Len()), nil
}

func (u *User) UnmarshalBinary(data []byte) (err error) {
	_, err = u.UnmarshalBinaryData(data)
	return
}

func (u *User) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[User] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data

	var username string
	username, newData, err = primitives.UnmarshalStringFromBytesData(newData, UsernameMaxLength)
	if err != nil {
		return data, nil
	}
	u.Username = username

	copy(u.PasswordHash[:], newData[:32])
	newData = newData[32:]

	copy(u.Salt[:], newData[:SaltLength])
	newData = newData[SaltLength:]

	newData, err = u.PoloniexKeys.UnmarshalBinaryData(newData)
	if err != nil {
		return data, nil
	}

	err = u.StartTime.UnmarshalBinary(newData)
	if err != nil {
		return data, err
	}
	newData = newData[15:]

	buf := bytes.NewBuffer(newData[:8])
	var f *float64
	err = binary.Read(buf, binary.BigEndian, f)
	if err != nil {
		return data, err
	}
	u.MiniumLend = *f
	newData = newData[8:]

	return newData, nil
}
