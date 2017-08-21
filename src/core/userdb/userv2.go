package userdb

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
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

var AvaiableCoins = []string{
	"BTC",
	"BTS",
	"CLAM",
	"DASH",
	"DOGE",
	"EOS",
	"ETC",
	"ETH",
	"FCT",
	"IOT",
	"LTC",
	"MAID",
	"STR",
	"USD",
	"XMR",
	"XRP",
	"ZEC",
}

func CoinExists(coin string) bool {
	for _, e := range AvaiableCoins {
		if e == coin {
			return true
		}
	}
	return false
}

//please add to all levels
var AllLevels = []UserLevel{
	SysAdmin,
	Admin,
	Moderator,
	CommonUser,
	Unassigned,
}

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

func StringToLevel(levelString string) UserLevel {
	switch levelString {
	case "Unassigned":
		return Unassigned
	case "SysAdmin":
		return SysAdmin
	case "Admin":
		return Admin
	case "Moderator":
		return Moderator
	case "CommonUser":
		return CommonUser
	default:
		return CommonUser
	}
}

const VerifyLength int = 64
const UsernameMaxLength int = 100
const SaltLength int = 5

type LendingHalt struct {
	Halt      bool      `json:"halt" bson:"halt"` //true = halt payments
	Reason    string    `json:"reason" bson:"reason"`
	Time      time.Time `json:"time" bson:"time"`
	TimeEmail time.Time `json:"timeemail" bson:"timeemail"` //last time an notification email was sent
	//EmailStop bool      `json:"timeemail" bson:"timeemail"` used to throttle emails
	/////////////////JESSE ADD THROTTLING
}

type User struct {
	Username     string   `bson:"_id" json:"username"` // Not case sensitive
	PasswordHash [32]byte `json:"passhash"`
	Salt         []byte   `json:"salt"`

	StartTime       time.Time `json:"starttime"`
	JWTTime         time.Time `json:"jwtime"`
	Level           UserLevel `json:"level"`
	LendingStrategy uint32    `json:"lendstrat"`

	// 2FA
	Has2FA     bool             `json:"has2fa"`
	Enabled2FA bool             `json:"enabled2fa"`
	User2FA    *primitives.Totp `json:"user2fa"`
	Issuer     string           `json:"issuer"`

	// JWT Change Pass
	JWTOTP [43]byte `json:"jwtotp"`

	// Email Verify
	Verified     bool   `json:"verified"`
	VerifyString string `json:"verifystring"`

	PoloniexMiniumLend  PoloniexMiniumumLendStruct `json:"polominlend"`
	PoloniexEnabled     PoloniexEnabledStruct      `json:"poloenabled"`
	PoloniexEnabledTime map[string]time.Time       `json:"poloenabledtime"` // When activated
	PoloniexKeys        *ExchangeKeys              `json:"polokeys"`

	BitfinexMiniumumLend BitfinexMiniumumLendStruct
	BitfinexEnabled      BitfinexEnabledStruct
	BitfinexKeys         *ExchangeKeys

	SessionExpiryTime time.Duration `bson:"sesexpdur"`

	LendingHalted LendingHalt `json:"lendhalt" bson:"lendhalt"`
}

func (u *User) SafeUnmarshal(data []byte) error {
	u1 := NewV1BlankUser()
	n, err := u1.UnmarshalBinaryData(data)
	if err == nil && len(n) == 0 {
		u.UserToV2User(u1)
		return nil
	}
	return json.Unmarshal(data, u)
}

func (u2 *User) UserToV2User(u *UserV1) {
	u2.Username = u.Username
	u2.PasswordHash = u.PasswordHash
	u2.Salt = u.Salt
	u2.StartTime = u.StartTime
	u2.JWTTime = u.JWTTime
	u2.Level = u.Level
	u2.LendingStrategy = u.LendingStrategy
	u2.Has2FA = u.Has2FA
	u2.Enabled2FA = u.Enabled2FA
	u2.Issuer = u.Issuer
	u2.JWTOTP = u.JWTOTP
	u2.Verified = u.Verified
	u2.VerifyString = u.VerifyString
	u2.PoloniexMiniumLend = u.MiniumLend
	u2.PoloniexEnabled = u.PoloniexEnabled
	u2.PoloniexKeys = u.PoloniexKeys
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
	u.User2FA = new(primitives.Totp)
	u.PoloniexKeys = NewBlankExchangeKeys()
	u.BitfinexKeys = NewBlankExchangeKeys()
	u.PoloniexEnabledTime = make(map[string]time.Time)
	return u
}

func NewUser(username string, password string) (*User, error) {
	u := NewBlankUser()

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

	u.PasswordHash = u.MakePasswordHash(password)

	u.StartTime = time.Now()
	u.JWTTime = time.Now()
	u.Level = CommonUser
	u.Has2FA = false

	verifyBytes := make([]byte, 32)
	_, err = rand.Read(verifyBytes)
	if err != nil {
		return nil, err
	}

	u.Verified = false
	u.VerifyString = hex.EncodeToString(verifyBytes)

	u.SessionExpiryTime = DEFAULT_SESSION_DUR

	return u, nil
}

func (u *User) MakePasswordHash(password string) [32]byte {
	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	return hash
}

func (u *User) ClearJWTOTP() {
	var tmp [43]byte
	u.JWTOTP = tmp
}

func (u *User) SetJWTOTP(jwtOTP [43]byte) error {
	if _, found := u.GetJWTOTP(); found {
		return fmt.Errorf("User already has a JWTOTP")
	}
	u.JWTOTP = jwtOTP
	return nil
}

func (u *User) GetJWTOTP() (jwtOTP [43]byte, found bool) {
	if bytes.Compare(u.JWTOTP[:], make([]byte, 43)) == 0 {
		return jwtOTP, false
	}
	return u.JWTOTP, true
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

func (u *User) GetLevelString() string {
	return LevelToString(u.Level)
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

	am, _ := a.PoloniexMiniumLend.MarshalBinary()
	bm, _ := b.PoloniexMiniumLend.MarshalBinary()
	if bytes.Compare(am, bm) != 0 {
		return false
	}

	if a.LendingStrategy != b.LendingStrategy {
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

	if a.Verified != b.Verified {
		return false
	}

	if bytes.Compare(a.JWTOTP[:], b.JWTOTP[:]) != 0 {
		return false
	}

	if a.VerifyString != b.VerifyString {
		return false
	}

	if bytes.Compare(a.PoloniexEnabled.Bytes(), b.PoloniexEnabled.Bytes()) != 0 {
		return false
	}

	return true
}

func (u *User) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	l, err := primitives.BytesToUint32(data[:4])
	if err != nil {
		return nil, err
	}

	// Not a v2 marshal
	if int(l) > len(data) {
		err = u.SafeUnmarshal(data)
		if err != nil {
			return nil, err
		} else {
			return []byte{}, nil
		}
	}

	err = u.SafeUnmarshal(data[4 : l+4])
	if err != nil {
		return nil, err
	}
	return data[4+l:], nil
}

func (u *User) UnmarshalBinary(data []byte) error {
	_, e := u.UnmarshalBinaryData(data)
	return e
}

func (u *User) MarshalBinary() ([]byte, error) {
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	l := len(data)
	return append(primitives.Uint32ToBytes(uint32(l)), data...), nil
}
