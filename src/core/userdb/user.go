package userdb

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DistributedSolutions/twofactor"
	"github.com/Emyrk/LendingBot/src/core/common/primitives"
)

type UserV1 struct {
	Username     string // Not case sensitive
	PasswordHash [32]byte
	Salt         []byte

	StartTime       time.Time
	JWTTime         time.Time
	Level           UserLevel
	MiniumLend      PoloniexMiniumumLendStruct
	LendingStrategy uint32

	// 2FA
	Has2FA     bool
	Enabled2FA bool
	User2FA    *twofactor.Totp
	Issuer     string

	// JWT Change Pass
	JWTOTP [43]byte

	// Email Verify
	Verified     bool
	VerifyString string

	PoloniexEnabled PoloniexEnabledStruct
	PoloniexKeys    *ExchangeKeys
}

func NewV1BlankUser() *UserV1 {
	u := new(UserV1)
	u.PoloniexKeys = NewBlankExchangeKeys()
	return u
}

func NewV1User(username string, password string) (*UserV1, error) {
	u := new(UserV1)

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

	u.PoloniexKeys = NewBlankExchangeKeys()

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

	return u, nil
}

func (u *UserV1) MakePasswordHash(password string) [32]byte {
	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	return hash
}

func (u *UserV1) ClearJWTOTP() {
	var tmp [43]byte
	u.JWTOTP = tmp
}

func (u *UserV1) SetJWTOTP(jwtOTP [43]byte) error {
	if _, found := u.GetJWTOTP(); found {
		return fmt.Errorf("User already has a JWTOTP")
	}
	u.JWTOTP = jwtOTP
	return nil
}

func (u *UserV1) GetJWTOTP() (jwtOTP [43]byte, found bool) {
	if bytes.Compare(u.JWTOTP[:], make([]byte, 43)) == 0 {
		return jwtOTP, false
	}
	return u.JWTOTP, true
}

func (u *UserV1) AuthenticatePassword(password string) bool {
	hash := u.getPasswordHashFromPassword(password)
	if bytes.Compare(u.PasswordHash[:], hash[:]) == 0 {
		return true
	}
	return false
}

func (u *UserV1) GetCipherKey(cipherKey [32]byte) [32]byte {
	uhash := GetUsernameHash(u.Username)
	return sha256.Sum256(append(cipherKey[:], uhash[:]...))
}

func (u *UserV1) getPasswordHashFromPassword(password string) [32]byte {
	passAndSalt := append([]byte(password), u.Salt...)
	hash := sha256.Sum256(passAndSalt)
	return hash
}

func (u *UserV1) GetLevelString() string {
	return LevelToString(u.Level)
}

func (u *UserV1) String() string {
	return fmt.Sprintf("UserName: %s, Level: %s", u.Username, LevelToString(u.Level))
}

func (a *UserV1) IsSameAs(b *UserV1) bool {
	if a.Username != b.Username {
		return false
	}

	if bytes.Compare(a.PasswordHash[:], b.PasswordHash[:]) != 0 {
		return false
	}

	if bytes.Compare(a.Salt, b.Salt) != 0 {
		return false
	}

	am, _ := a.MiniumLend.MarshalBinary()
	bm, _ := b.MiniumLend.MarshalBinary()
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

func (u *UserV1) MarshalBinary() ([]byte, error) {
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

	b, err = u.MiniumLend.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// Lending Strat
	b = primitives.Uint32ToBytes(u.LendingStrategy)
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

	// JWTOTP
	buf.Write(u.JWTOTP[:43])

	// Email Verified
	b = primitives.BoolToBytes(u.Verified)
	buf.Write(b)

	// Verify String
	b, err = primitives.MarshalStringToBytes(u.VerifyString, 64)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// PoloniexKeys
	b = u.PoloniexEnabled.Bytes()
	buf.Write(b)

	// PoloniexKeys
	b, err = u.PoloniexKeys.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	return buf.Next(buf.Len()), nil
}

func (u *UserV1) UnmarshalBinary(data []byte) (err error) {
	_, err = u.UnmarshalBinaryData(data)
	return
}

func (u *UserV1) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
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
	newData, err = u.MiniumLend.UnmarshalBinaryData(newData)
	if err != nil {
		return data, err
	}

	// Lending Strat
	v, err = primitives.BytesToUint32(newData[:4])
	if err != nil {
		return data, err
	}
	u.LendingStrategy = v
	newData = newData[4:]

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

	copy(u.JWTOTP[:43], newData[:43])
	newData = newData[43:]

	// Verified
	verified := primitives.ByteToBool(newData[0])
	newData = newData[1:]
	u.Verified = verified

	// VerifyString
	var vrystr string
	vrystr, newData, err = primitives.UnmarshalStringFromBytesData(newData, 64)
	if err != nil {
		return data, nil
	}
	u.VerifyString = vrystr

	// PoloniexEnabled
	newData, err = u.PoloniexEnabled.UnmarshalBinaryData(newData)
	if err != nil {
		return data, nil
	}

	// PoloniexKeys
	newData, err = u.PoloniexKeys.UnmarshalBinaryData(newData)
	if err != nil {
		return data, err
	}

	return newData, nil
}

type PoloniexMiniumumLendStruct struct {
	BTC  float64 `json:"BTC"`
	BTS  float64 `json:"BTS"`
	CLAM float64 `json:"CLAM"`
	DOGE float64 `json:"DOGE"`
	DASH float64 `json:"DASH"`
	LTC  float64 `json:"LTC"`
	MAID float64 `json:"MAID"`
	STR  float64 `json:"STR"`
	XMR  float64 `json:"XMR"`
	XRP  float64 `json:"XRP"`
	ETH  float64 `json:"ETH"`
	FCT  float64 `json:"FCT"`
}

func (m *PoloniexMiniumumLendStruct) GetAll() []float64 {
	var mins []float64

	mins = append(mins, m.BTC)
	mins = append(mins, m.BTS)
	mins = append(mins, m.CLAM)
	mins = append(mins, m.DOGE)
	mins = append(mins, m.DASH)
	mins = append(mins, m.LTC)
	mins = append(mins, m.MAID)
	mins = append(mins, m.STR)
	mins = append(mins, m.XMR)
	mins = append(mins, m.XRP)
	mins = append(mins, m.ETH)
	mins = append(mins, m.FCT)

	return mins
}

func (m *PoloniexMiniumumLendStruct) Set(currency string, min float64) bool {
	switch currency {
	case "BTC":
		m.BTC = min
	case "BTS ":
		m.BTS = min
	case "CLAM":
		m.CLAM = min
	case "DOGE":
		m.DOGE = min
	case "DASH":
		m.DASH = min
	case "LTC ":
		m.LTC = min
	case "MAID":
		m.MAID = min
	case "STR ":
		m.STR = min
	case "XMR ":
		m.XMR = min
	case "XRP ":
		m.XRP = min
	case "ETH ":
		m.ETH = min
	case "FCT ":
		m.FCT = min
	default:
		return false
	}
	return true
}

func (m *PoloniexMiniumumLendStruct) Get(currency string) float64 {
	switch currency {
	case "BTC":
		return m.BTC
	case "BTS":
		return m.BTS
	case "CLAM":
		return m.CLAM
	case "DOGE":
		return m.DOGE
	case "DASH":
		return m.DASH
	case "LTC":
		return m.LTC
	case "MAID":
		return m.MAID
	case "STR":
		return m.STR
	case "XMR":
		return m.XMR
	case "XRP":
		return m.XRP
	case "ETH":
		return m.ETH
	case "FCT":
		return m.FCT
	}
	return 0
}

func (m *PoloniexMiniumumLendStruct) SetAll(coins PoloniexMiniumumLendStruct) {
	m.BTC = coins.BTC
	m.BTS = coins.BTS
	m.CLAM = coins.CLAM
	m.DOGE = coins.DOGE
	m.DASH = coins.DASH
	m.LTC = coins.LTC
	m.MAID = coins.MAID
	m.STR = coins.STR
	m.XMR = coins.XMR
	m.XRP = coins.XRP
	m.ETH = coins.ETH
	m.FCT = coins.FCT
}

func (m *PoloniexMiniumumLendStruct) UnmarshalBinary(data []byte) (err error) {
	_, err = m.UnmarshalBinaryData(data)
	return
}

func (m *PoloniexMiniumumLendStruct) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[PoloniexEnabledStruct] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data
	var v float64

	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.BTC = v

	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.BTS = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.CLAM = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.DOGE = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.DASH = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.LTC = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.MAID = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.STR = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.XMR = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.XRP = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.ETH = v
	v, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return data, err
	}
	m.FCT = v

	return
}

func (m *PoloniexMiniumumLendStruct) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	data, err := primitives.Float64ToBytes(m.BTC)
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = primitives.Float64ToBytes(m.BTS)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.CLAM)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.DOGE)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.DASH)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.LTC)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.MAID)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.STR)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.XMR)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.XRP)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.ETH)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	data, err = primitives.Float64ToBytes(m.FCT)
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Next(buf.Len()), nil
}

type PoloniexEnabledStruct struct {
	BTC  bool `json:"BTC"`
	BTS  bool `json:"BTS"`
	CLAM bool `json:"CLAM"`
	DOGE bool `json:"DOGE"`
	DASH bool `json:"DASH"`
	LTC  bool `json:"LTC"`
	MAID bool `json:"MAID"`
	STR  bool `json:"STR"`
	XMR  bool `json:"XMR"`
	XRP  bool `json:"XRP"`
	ETH  bool `json:"ETH"`
	FCT  bool `json:"FCT"`
}

type EnabledStruct struct {
	Currency string `json:"currency"`
	Enabled  bool   `json:"enabled"`
}

func (pe *PoloniexEnabledStruct) GetAll() []EnabledStruct {
	var arr []EnabledStruct
	arr = append(arr, EnabledStruct{"BTC", pe.BTC})
	arr = append(arr, EnabledStruct{"BTS", pe.BTS})
	arr = append(arr, EnabledStruct{"CLAM", pe.CLAM})
	arr = append(arr, EnabledStruct{"DOGE", pe.DOGE})
	arr = append(arr, EnabledStruct{"DASH", pe.DASH})
	arr = append(arr, EnabledStruct{"LTC", pe.LTC})
	arr = append(arr, EnabledStruct{"MAID", pe.MAID})
	arr = append(arr, EnabledStruct{"STR", pe.STR})
	arr = append(arr, EnabledStruct{"XMR", pe.XMR})
	arr = append(arr, EnabledStruct{"XRP", pe.XRP})
	arr = append(arr, EnabledStruct{"ETH", pe.ETH})
	arr = append(arr, EnabledStruct{"FCT", pe.FCT})
	return arr
}

func (pe *PoloniexEnabledStruct) Keys() []string {
	var arr []string
	if pe.BTC {
		arr = append(arr, "BTC")
	}
	if pe.BTS {
		arr = append(arr, "BTS")
	}
	if pe.CLAM {
		arr = append(arr, "CLAM")
	}
	if pe.DOGE {
		arr = append(arr, "DOGE")
	}
	if pe.DASH {
		arr = append(arr, "DASH")
	}
	if pe.LTC {
		arr = append(arr, "LTC")
	}
	if pe.MAID {
		arr = append(arr, "MAID")
	}
	if pe.STR {
		arr = append(arr, "STR")
	}
	if pe.XMR {
		arr = append(arr, "XMR")
	}
	if pe.XRP {
		arr = append(arr, "XRP")
	}
	if pe.ETH {
		arr = append(arr, "ETH")
	}
	if pe.FCT {
		arr = append(arr, "FCT")
	}
	return arr
}

//added in coin for future to enable specific coin
func (pe *PoloniexEnabledStruct) Enable(coins PoloniexEnabledStruct) {
	pe.BTC = coins.BTC
	pe.BTS = coins.BTS
	pe.CLAM = coins.CLAM
	pe.DOGE = coins.DOGE
	pe.DASH = coins.DASH
	pe.LTC = coins.LTC
	pe.MAID = coins.MAID
	pe.STR = coins.STR
	pe.XMR = coins.XMR
	pe.XRP = coins.XRP
	pe.ETH = coins.ETH
	pe.FCT = coins.FCT
}

func (pe *PoloniexEnabledStruct) Bytes() []byte {
	b, _ := pe.MarshalBinary()
	return b
}

func (pe *PoloniexEnabledStruct) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(primitives.BoolToBytes(pe.BTC))
	buf.Write(primitives.BoolToBytes(pe.BTS))
	buf.Write(primitives.BoolToBytes(pe.CLAM))
	buf.Write(primitives.BoolToBytes(pe.DOGE))
	buf.Write(primitives.BoolToBytes(pe.DASH))
	buf.Write(primitives.BoolToBytes(pe.LTC))
	buf.Write(primitives.BoolToBytes(pe.MAID))
	buf.Write(primitives.BoolToBytes(pe.STR))
	buf.Write(primitives.BoolToBytes(pe.XMR))
	buf.Write(primitives.BoolToBytes(pe.XRP))
	buf.Write(primitives.BoolToBytes(pe.ETH))
	buf.Write(primitives.BoolToBytes(pe.FCT))
	return buf.Next(buf.Len()), nil
}

func (pe *PoloniexEnabledStruct) UnmarshalBinary(data []byte) error {
	_, err := pe.UnmarshalBinaryData(data)
	return err
}

func (pe *PoloniexEnabledStruct) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[PoloniexEnabledStruct] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data
	pe.BTC = primitives.ByteToBool(newData[0])
	pe.BTS = primitives.ByteToBool(newData[1])
	pe.CLAM = primitives.ByteToBool(newData[2])
	pe.DOGE = primitives.ByteToBool(newData[3])
	pe.DASH = primitives.ByteToBool(newData[4])
	pe.LTC = primitives.ByteToBool(newData[5])
	pe.MAID = primitives.ByteToBool(newData[6])
	pe.STR = primitives.ByteToBool(newData[7])
	pe.XMR = primitives.ByteToBool(newData[8])
	pe.XRP = primitives.ByteToBool(newData[9])
	pe.ETH = primitives.ByteToBool(newData[10])
	pe.FCT = primitives.ByteToBool(newData[11])
	newData = newData[12:]
	return
}

type BitfinexMiniumumLendStruct struct {
	USD  float64 `json:"USD"`
	BTC  float64 `json:"BTC"`
	ETH  float64 `json:"ETH"`
	ETC  float64 `json:"ETC"`
	DASH float64 `json:"DASH"`
	ZEC  float64 `json:"ZEC"`
	XMR  float64 `json:"XMR"`
	LTC  float64 `json:"LTC"`
	EOS  float64 `json:"EOS"`
	IOT  float64 `json:"IOT"`
}

func (m *BitfinexMiniumumLendStruct) GetAll() []float64 {
	var mins []float64

	mins = append(mins, m.USD)
	mins = append(mins, m.BTC)
	mins = append(mins, m.ETH)
	mins = append(mins, m.ETC)
	mins = append(mins, m.DASH)
	mins = append(mins, m.ZEC)
	mins = append(mins, m.XMR)
	mins = append(mins, m.LTC)
	mins = append(mins, m.EOS)
	mins = append(mins, m.IOT)

	return mins
}

func (m *BitfinexMiniumumLendStruct) Set(currency string, min float64) bool {
	switch currency {
	case "USD":
		m.USD = min
	case "BTC ":
		m.BTC = min
	case "ETH":
		m.ETH = min
	case "ETC":
		m.ETC = min
	case "DASH":
		m.DASH = min
	case "ZEC ":
		m.ZEC = min
	case "XMR":
		m.XMR = min
	case "LTC ":
		m.LTC = min
	case "EOS":
		m.XMR = min
	case "IOT ":
		m.LTC = min
	default:
		return false
	}
	return true
}

func (m *BitfinexMiniumumLendStruct) Get(currency string) float64 {
	switch currency {
	case "USD":
		return m.USD
	case "BTC":
		return m.BTC
	case "ETH":
		return m.ETH
	case "ETC":
		return m.ETC
	case "DASH":
		return m.DASH
	case "ZEC":
		return m.ZEC
	case "XMR":
		return m.XMR
	case "LTC":
		return m.LTC
	case "EOS":
		return m.XMR
	case "IOT ":
		return m.IOT
	}
	return 0
}

func (m *BitfinexMiniumumLendStruct) SetAll(coins BitfinexMiniumumLendStruct) {
	m.USD = coins.USD
	m.BTC = coins.BTC
	m.ETH = coins.ETH
	m.ETC = coins.ETC
	m.DASH = coins.DASH
	m.ZEC = coins.ZEC
	m.XMR = coins.XMR
	m.LTC = coins.LTC
}

type BitfinexEnabledStruct struct {
	USD  bool `json:"USD"`
	BTC  bool `json:"BTC"`
	ETH  bool `json:"ETH"`
	ETC  bool `json:"ETC"`
	DASH bool `json:"DASH"`
	ZEC  bool `json:"ZEC"`
	XMR  bool `json:"XMR"`
	LTC  bool `json:"LTC"`
	EOS  bool `json:"EOS"`
	IOT  bool `json:"IOT"`
}

func (pe *BitfinexEnabledStruct) GetAll() []EnabledStruct {
	var arr []EnabledStruct
	arr = append(arr, EnabledStruct{"USD", pe.USD})
	arr = append(arr, EnabledStruct{"BTC", pe.BTC})
	arr = append(arr, EnabledStruct{"ETH", pe.ETH})
	arr = append(arr, EnabledStruct{"ETC", pe.ETC})
	arr = append(arr, EnabledStruct{"DASH", pe.DASH})
	arr = append(arr, EnabledStruct{"ZEC", pe.ZEC})
	arr = append(arr, EnabledStruct{"XMR", pe.XMR})
	arr = append(arr, EnabledStruct{"LTC", pe.LTC})
	arr = append(arr, EnabledStruct{"EOS", pe.EOS})
	arr = append(arr, EnabledStruct{"IOT", pe.IOT})
	return arr
}

func (pe *BitfinexEnabledStruct) Keys() []string {
	var arr []string
	if pe.USD {
		arr = append(arr, "USD")
	}
	if pe.BTC {
		arr = append(arr, "BTC")
	}
	if pe.ETH {
		arr = append(arr, "ETH")
	}
	if pe.ETC {
		arr = append(arr, "ETC")
	}
	if pe.DASH {
		arr = append(arr, "DASH")
	}
	if pe.ZEC {
		arr = append(arr, "ZEC")
	}
	if pe.XMR {
		arr = append(arr, "XMR")
	}
	if pe.LTC {
		arr = append(arr, "LTC")
	}
	if pe.EOS {
		arr = append(arr, "EOS")
	}
	if pe.IOT {
		arr = append(arr, "IOT")
	}
	return arr
}

//added in coin for future to enable specific coin
func (pe *BitfinexEnabledStruct) Enable(coins BitfinexEnabledStruct) {
	pe.USD = coins.USD
	pe.BTC = coins.BTC
	pe.ETH = coins.ETH
	pe.ETC = coins.ETC
	pe.DASH = coins.DASH
	pe.ZEC = coins.ZEC
	pe.XMR = coins.XMR
	pe.LTC = coins.LTC
	pe.EOS = coins.EOS
	pe.IOT = coins.IOT
}

type IExchangeEnabled interface {
	GetAll() []EnabledStruct
}

type UserExchange string

const (
	PoloniexExchange UserExchange = "pol"
	BitfinexExchange UserExchange = "bit"
)
