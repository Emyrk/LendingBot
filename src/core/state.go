package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/cryption"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/badoux/checkmail"
	"github.com/revel/revel"
)

type State struct {
	userDB          *userdb.UserDatabase
	userStatistic   *userdb.UserStatisticsDB
	userInviteCodes *userdb.InviteDB
	PoloniexAPI     poloniex.IPoloniex
	CipherKey       [32]byte
	JWTSecret       [32]byte

	// Poloniex Cache
	poloniexCache *PoloniexAccessCache
}

func NewFakePoloniexState() *State {
	return newState(true, true)
}

func NewState() *State {
	return newState(false, false)
}

func NewStateWithMap() *State {
	return newState(true, false)
}

func (s *State) VerifyState() error {
	return s.userDB.VerifyDatabase(s.CipherKey)
}

func newState(withMap bool, fakePolo bool) *State {
	s := new(State)
	if withMap {
		s.userDB = userdb.NewMapUserDatabase()
	} else {
		v := os.Getenv("USER_DB")
		if len(v) == 0 {
			v = "UserDatabase.db"
		}
		s.userDB = userdb.NewBoltUserDatabase(v)
	}

	if fakePolo {
		s.PoloniexAPI = poloniex.StartFakePoloniex()
	} else {
		s.PoloniexAPI = poloniex.StartPoloniex()
	}

	if !withMap {
		s.CipherKey = getCipherKey()
	}

	jck := make([]byte, 32)
	_, err := rand.Read(jck)
	if err != nil {
		panic(fmt.Sprintf("Could not generate JWT Siging Key %s", err.Error()))
	}
	copy(s.JWTSecret[:], jck[:])

	if withMap {
		s.userStatistic, err = userdb.NewUserStatisticsMapDB()
	} else {
		s.userStatistic, err = userdb.NewUserStatisticsDB()
	}
	if err != nil {
		panic(fmt.Sprintf("Could create user statistic database %s", err.Error()))
	}

	s.poloniexCache = NewPoloniexAccessCache()

	return s
}

func getCipherKey() [32]byte {
	var sec [32]byte

	str := os.Getenv("HODLZONE_KEY")
	if str == "" {
		if !revel.DevMode {
			panic("No private key given when running! I won't allow that!")
		}
		fmt.Println("WARNING! NO PRIVATE KEY IS GIVEN. I'll let it go as we are in devmode")
		ck := make([]byte, 32)
		copy(sec[:32], ck[:32])
		return sec
	}
	ck, err := hex.DecodeString(str)
	if err != nil {
		panic(fmt.Sprintf("Error when parsing key: %s", err.Error()))
	}
	if len(ck) != 32 {
		panic(fmt.Sprintf("Key length must be 32 bytes, found %d", len(ck)))
	}
	copy(sec[:32], ck[:32])
	return sec
}

func (s *State) Close() error {
	return s.userDB.Close()
}

func (s *State) SetUserMinimumLoan(username string, minimumAmt float64, currency string) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.MiniumLend.Set(currency, minimumAmt)
	return s.userDB.PutUser(u)
}

func (s *State) NewUser(username string, password string) *primitives.ApiError {
	ou, err := s.userDB.FetchUser(username)
	if err != nil {
		return &primitives.ApiError{
			fmt.Errorf("could not check if username exists: %s", err.Error()),
			fmt.Errorf("Internal error. Please try again."),
		}
	}

	if ou != nil {
		return &primitives.ApiError{
			fmt.Errorf("Attempted to create duplicate user: %s", ou.Username),
			fmt.Errorf("Username already exists."),
		}
	}

	if err := ValidateEmail(username); err != nil {
		return &primitives.ApiError{
			err,
			fmt.Errorf("Email failed to validate."),
		}
	}

	u, err := userdb.NewUser(username, password)
	if err != nil {
		return &primitives.ApiError{
			err,
			fmt.Errorf("Failed to create user. Please try again."),
		}
	}

	err = s.userDB.PutVerifystring(userdb.GetUsernameHash(username), u.VerifyString)
	if err != nil {
		return &primitives.ApiError{
			err,
			fmt.Errorf("Failed to create user. Please try again."),
		}
	}

	err = s.userDB.PutUser(u)
	if err != nil {
		return &primitives.ApiError{
			err,
			fmt.Errorf("Internal error. Please try again."),
		}
	}

	return nil
}

func ValidateEmail(email string) error {
	return checkmail.ValidateFormat(email)
}

func (s *State) ListInviteCodes() ([]userdb.InviteEntry, error) {
	return s.userInviteCodes.ListAll()
}

func (s *State) ClaimInviteCode(username string, code string) (bool, error) {
	return s.userInviteCodes.ClaimInviteCode(username, code)
}

func (s *State) AddInviteCode(code string, capacity int, expires time.Time) error {
	return s.userInviteCodes.CreateInviteCode(code, capacity, expires)
}

func (s *State) SetUserKeys(username string, acessKey string, secretKey string) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	pk, err := userdb.NewPoloniexKeys(acessKey, secretKey, u.GetCipherKey(s.CipherKey))
	if err != nil {
		return err
	}

	u.PoloniexKeys = pk

	return s.userDB.PutUser(u)
}

func (s *State) GetUserStatistics(username string, dayRange int) ([][]userdb.UserStatistic, error) {
	return s.userStatistic.GetStatistics(username, dayRange)
}

func (s *State) EnableUserLending(username string, coins userdb.PoloniexEnabledStruct) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.PoloniexEnabled.Enable(coins)
	// if !enabled {
	// 	s.removeFromPoloniexCache(username)
	// }

	return s.userDB.PutUser(u)
}

func (s *State) FetchUser(username string) (*userdb.User, error) {
	return s.userDB.FetchUser(username)
}

func (s *State) FetchAllUsers() ([]userdb.User, error) {
	return s.userDB.FetchAllUsers()
}

func (s *State) GetPoloniexStatistics(currency string) *userdb.PoloniexStats {
	return s.userStatistic.GetPoloniexStatistics(currency)
}

// RecordPoloniexStatistics is for recording the current lending rate on poloniex
func (s *State) RecordPoloniexStatistics(currency string, rate float64) error {
	return s.userStatistic.RecordPoloniexStatistic(currency, rate)
}

func (s *State) GetPoloniexStatsPastXDays(dayRange int, currency string) [][]userdb.PoloniexRateSample {
	return s.userStatistic.GetPoloniexDataLastXDays(dayRange, currency)
}

// RecordStatistics is for recording an individual user's statistics at a given time
func (s *State) RecordStatistics(stats *userdb.UserStatistic) error {
	if !s.poloniexCache.shouldRecordStats(stats.Username) {
		return nil
	}

	err := s.userStatistic.RecordData(stats)
	if err != nil {
		return err
	}

	return nil
}

func (s *State) GetNewJWTOTP(username string) (string, error) {
	return s.setupNewJWTOTP(username, cryption.JWT_EXPIRY_TIME_NEW_PASS)
}

func (s *State) setupNewJWTOTP(username string, t time.Duration) (string, error) {
	tokenString, err := cryption.NewJWTString(username, s.JWTSecret, t)
	if err != nil {
		return "", err
	}
	sig, err := cryption.GetJWTSignature(tokenString)
	if err != nil {
		return "", err
	}

	var b [43]byte
	copy(b[:], sig)
	if err = s.userDB.UpdateJWTOTP(username, b); err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *State) SetNewPasswordJWTOTP(tokenString string, password string) bool {
	token, err := cryption.VerifyJWT(tokenString, s.JWTSecret)
	if err != nil {
		fmt.Printf("Error comparing JWT for pass reset: %s\n", err.Error())
		return false
	}

	email, ok := token.Claims().Get("email").(string)
	if !ok {
		fmt.Printf("Error Retrieving email for pass reset: %s\n", err.Error())
		return false
	}

	userSig, ok := s.userDB.GetJWTOTP(email)
	if !ok {
		fmt.Printf("Error with getting Token for user for pass reset: %s\n", err.Error())
		return false
	}

	tokenSig, err := cryption.GetJWTSignature(tokenString)
	if err != nil {
		fmt.Printf("Error retrieving sig for JWT for pass reset: %s\n", err)
		return false
	}

	s.setUserPass(email, password, nil)

	return string(userSig[:]) == tokenSig
}

func (s *State) SetUserNewPass(username string, oldPassword string, newPassword string) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	if !u.AuthenticatePassword(oldPassword) {
		return fmt.Errorf("Error resting user pass, old pass hash does not match up.")
	}

	return s.setUserPass(username, newPassword, u)
}

func (s *State) setUserPass(username string, password string, u *userdb.User) error {
	if u == nil {
		newU, err := s.userDB.FetchUserIfFound(username)
		if err != nil {
			return err
		}
		u = newU
	}

	hash := u.MakePasswordHash(password)
	u.PasswordHash = hash

	return s.userDB.PutUser(u)
}

type SafeUser struct {
	Username  string `json:"email"`
	Privilege string `json:"priv"`
}

func (s *State) GetAllUsers() (*[]SafeUser, error) {
	users, err := s.userDB.FetchAllUsers()
	if err != nil {
		return nil, fmt.Errorf("ERROR: Error getting all users: %s\n", err.Error())
	}
	var safeUsers []SafeUser
	for i, u := range users {
		safeUsers[i] = SafeUser{
			u.Username,
			u.GetLevelString(),
		}
	}
	return &safeUsers, nil
}

func (s *State) DeleteUser() error {
	//TODO DELETE USER
	return nil
}

func (s *State) UpdateUserPrivilege(priv string, email string) (*string, error) {
	u, err := s.userDB.FetchUserIfFound(email)
	if err != nil {
		return nil, err
	}
	u.Level = userdb.StringToLevel(priv)

	userLevelString := userdb.LevelToString(u.Level)
	return &userLevelString, s.userDB.PutUser(u)
}

func (s *State) HasUserPrivilege(email string, priv userdb.UserLevel) bool {
	u, err := s.userDB.FetchUserIfFound(email)
	if err != nil {
		fmt.Printf("WARNING: IMPORTANT: User does not have priv, but attempting to get in [%s]: %s\n", email, err.Error())
		return false
	}
	if u.Level < priv {
		fmt.Printf("WARNING: IMPORTANT: User does not have priv, but attempting to get in [%s]: %s\n", email, err.Error())
		return false
	}
	return true
}
