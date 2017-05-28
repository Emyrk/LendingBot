package userdb

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core/common/primitives"
)

type UserLevel uint32

const (
	// UserLevel
	Unassigned UserLevel = 0
	SysAdmin   UserLevel = 1
	Admin      UserLevel = 2
	Moderator  UserLevel = 3
	CommonUser UserLevel = 4
)

const UsernameMaxLength int = 100
const SaltLength int = 5

type User struct {
	Username     string // Not case sensitive
	PasswordHash [32]byte
	Salt         []byte

	StartTime  time.Time
	JWTTime    time.Time
	Level      UserLevel
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

func NewBlankUser() *User {
	u := new(User)
	u.PoloniexKeys = NewBlankPoloniexKeys()
	return u
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

	u.Level = CommonUser

	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	u.PasswordHash = hash

	u.PoloniexKeys = NewBlankPoloniexKeys()

	u.StartTime = time.Now()
	u.JWTTime = time.Now()
	u.Level = CommonUser
	return u, nil
}

func (u *User) Authenticate(password string) bool {
	hash := u.getPasswordHashFromPassword(password)
	if bytes.Compare(u.PasswordHash[:], hash[:]) == 0 {
		return true
	}
	return false
}

func (u *User) GetCipherKey(cipherKey [32]byte) [32]byte {
	uhash := GetUsernameHash(u.Username)
	return sha256.Sum256(append(cipherKey[:], uhash[:]...))
}

func (u *User) getPasswordHashFromPassword(password string) [32]byte {
	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	return hash
}

func (a *User) IsSameAs(b *User) bool {
	if a.Username != b.Username {
		return false
	}

	if bytes.Compare(a.PasswordHash[:], b.PasswordHash[:]) != 0 {
		return false
	}

	if bytes.Compare(a.Salt, b.Salt) != 0 {
		return false
	}

	if a.MiniumLend != b.MiniumLend {
		return false
	}

	if !a.PoloniexKeys.IsSameAs(b.PoloniexKeys) {
		return false
	}

	if a.Level != b.Level {
		return false
	}

	return true
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

	b, err = u.JWTTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b = primitives.Uint32ToBytes(uint32(u.Level))
	buf.Write(b)

	str := strconv.FormatFloat(u.MiniumLend, 'f', 6, 64)
	b, err = primitives.MarshalStringToBytes(str, 100)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = u.PoloniexKeys.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

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

	u.Salt = make([]byte, SaltLength)
	copy(u.Salt[:], newData[:SaltLength])
	newData = newData[SaltLength:]

	newData, err = u.PoloniexKeys.UnmarshalBinaryData(newData)
	if err != nil {
		return data, nil
	}

	err = u.StartTime.UnmarshalBinary(newData[:15])
	if err != nil {
		return data, err
	}
	newData = newData[15:]

	err = u.JWTTime.UnmarshalBinary(newData[:15])
	if err != nil {
		return data, err
	}
	newData = newData[15:]

	v, err := primitives.BytesToUint32(newData[:4])
	if err != nil {
		return data, err
	}
	u.Level = UserLevel(v)
	newData = newData[4:]

	// Float64
	var resp string
	resp, newData, err = primitives.UnmarshalStringFromBytesData(newData, 100)
	if err != nil {
		return data, err
	}
	f, err := strconv.ParseFloat(resp, 64)
	if err != nil {
		return data, err
	}
	u.MiniumLend = f
	//

	newData, err = u.PoloniexKeys.UnmarshalBinaryData(newData)
	if err != nil {
		return data, err
	}

	return newData, nil
}
