package userdb

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DistributedSolutions/twofactor"
	"github.com/Emyrk/LendingBot/app/core/common/primitives"
)

type UserLevel uint32

const (
	// UserLevel
	SysAdmin   UserLevel = 1000
	Admin      UserLevel = 999
	Moderator  UserLevel = 998
	CommonUser UserLevel = 997
	Unassigned UserLevel = 0
)

func LevelToString(l UserLevel) string {
	switch l {
	case Unassigned:
		return "Unassigned"
	case SysAdmin:
		return "SysAdmin"
	case Admin:
		return "Admin"
	case Moderator:
		return "Moderator"
	case CommonUser:
		return "CommonUser"
	default:
		return "???"
	}
}

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

	// 2FA
	Has2FA     bool
	Enabled2FA bool
	User2FA    *twofactor.Totp
	Issuer     string

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
	u.Has2FA = false

	return u, nil
}

func (u *User) AuthenticatePassword(password string) bool {
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

func (u *User) String() string {
	return fmt.Sprintf("UserName: %s, Level: %s", u.Username, LevelToString(u.Level))
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

	if a.Has2FA != b.Has2FA {
		return false
	}

	if a.Enabled2FA != b.Enabled2FA {
		return false
	}

	if a.User2FA == nil && b.User2FA != nil {
		return false
	}

	if a.User2FA != nil && b.User2FA == nil {
		return false
	}

	return true
}

func (u *User) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// username
	b, err := primitives.MarshalStringToBytes(u.Username, UsernameMaxLength)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// password
	buf.Write(u.PasswordHash[:])
	//salt
	buf.Write(u.Salt[:])

	// starttime
	b, err = u.StartTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// jwttime
	b, err = u.JWTTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// level
	b = primitives.Uint32ToBytes(uint32(u.Level))
	buf.Write(b)

	// miniummlend
	str := strconv.FormatFloat(u.MiniumLend, 'f', 6, 64)
	b, err = primitives.MarshalStringToBytes(str, 100)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// has2fa
	b = primitives.BoolToBytes(u.Has2FA)
	buf.Write(b)

	if u.Has2FA {
		// 2fa enabled
		b = primitives.BoolToBytes(u.Enabled2FA)
		buf.Write(b)

		topBytes, err := u.User2FA.ToBytes()
		if err != nil {
			return nil, err
		}

		// issuer
		b, err = primitives.MarshalStringToBytes(u.Issuer, 100)
		if err != nil {
			return nil, err
		}
		buf.Write(b)

		// len 2fa
		l := len(topBytes)
		lb := primitives.Uint32ToBytes(uint32(l))
		buf.Write(lb)
		// 2fa
		buf.Write(topBytes)
	}

	// PoloniexKeys
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

	// username
	var username string
	username, newData, err = primitives.UnmarshalStringFromBytesData(newData, UsernameMaxLength)
	if err != nil {
		return data, nil
	}
	u.Username = username

	// password
	copy(u.PasswordHash[:], newData[:32])
	newData = newData[32:]

	//salt
	u.Salt = make([]byte, SaltLength)
	copy(u.Salt[:], newData[:SaltLength])
	newData = newData[SaltLength:]

	// starttime
	err = u.StartTime.UnmarshalBinary(newData[:15])
	if err != nil {
		return data, err
	}
	newData = newData[15:]

	// jwttime
	err = u.JWTTime.UnmarshalBinary(newData[:15])
	if err != nil {
		return data, err
	}
	newData = newData[15:]

	// level
	v, err := primitives.BytesToUint32(newData[:4])
	if err != nil {
		return data, err
	}
	u.Level = UserLevel(v)
	newData = newData[4:]

	// miniummlend
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

	// has2fa
	has2FA := primitives.ByteToBool(newData[0])
	newData = newData[1:]
	u.Has2FA = has2FA
	if has2FA {
		// 2fa enabled
		enabled := primitives.ByteToBool(newData[0])
		newData = newData[1:]
		u.Enabled2FA = enabled

		var issuer string
		issuer, newData, err = primitives.UnmarshalStringFromBytesData(newData, 100)
		if err != nil {
			return data, err
		}
		u.Issuer = issuer

		// len 2fa
		l, err := primitives.BytesToUint32(newData[:4])
		if err != nil {
			return data, err
		}
		newData = newData[4:]

		// 2fa
		totp, err := twofactor.TOTPFromBytes(newData[:l], u.Issuer)
		if err != nil {
			return data, err
		}
		u.User2FA = totp
		newData = newData[l:]
	}

	// PoloniexKeys
	newData, err = u.PoloniexKeys.UnmarshalBinaryData(newData)
	if err != nil {
		return data, err
	}

	return newData, nil
}
