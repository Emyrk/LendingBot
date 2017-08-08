package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/cryption"
	"github.com/Emyrk/LendingBot/src/core/payment"
	"github.com/Emyrk/LendingBot/src/core/poloniex"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/badoux/checkmail"
	"github.com/revel/revel"
	log "github.com/sirupsen/logrus"
)

const (
	DB_MAP   = iota
	DB_BOLT  = iota
	DB_MONGO = iota
)

func init() {
	RegisterPrometheus()
}

type State struct {
	userDB          *userdb.UserDatabase
	userStatistic   *userdb.UserStatisticsDB
	userInviteCodes *userdb.InviteDB
	paymentDB       *payment.PaymentDatabase
	PoloniexAPI     poloniex.IPoloniex
	CipherKey       [32]byte
	JWTSecret       [32]byte

	Master *Master

	// Poloniex Cache
	poloniexCache *PoloniexAccessCache
}

func NewFakePoloniexState() *State {
	return newState(DB_MAP, true)
}

func NewState() *State {
	return newState(DB_BOLT, false)
}

func NewStateWithMap() *State {
	return newState(DB_MAP, false)
}

func NewStateWithMongo() *State {
	return newState(DB_MONGO, false)
}

func (s *State) GetUserStatsDB() *userdb.UserStatisticsDB {
	return s.userStatistic
}

func (s *State) VerifyState() error {
	return s.userDB.VerifyDatabase(s.CipherKey)
}

func newState(dbType int, fakePolo bool) *State {
	uri := revel.Config.StringDefault("database.uri", "mongodb://localhost:27017")
	mongoRevelPass := os.Getenv("MONGO_REVEL_PASS")
	if mongoRevelPass == "" && revel.RunMode == "prod" {
		panic("Running in prod, but no revel pass given in env var 'MONGO_REVEL_PASS'")
	}

	var err error
	s := new(State)
	switch dbType {
	case DB_MAP:
		s.userDB = userdb.NewMapUserDatabase()
	case DB_BOLT:
		v := os.Getenv("USER_DB")
		if len(v) == 0 {
			v = "UserDatabase.db"
		}
		s.userDB = userdb.NewBoltUserDatabase(v)
	case DB_MONGO:
		s.userDB, err = userdb.NewMongoUserDatabase(uri, "revel", mongoRevelPass)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to user mongodb: %s\n", err.Error()))
		}
	default:
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

	if !revel.DevMode {
		s.CipherKey = getCipherKey()
	}

	jck := make([]byte, 32)
	_, err = rand.Read(jck)
	if err != nil {
		panic(fmt.Sprintf("Could not generate JWT Siging Key %s", err.Error()))
	}
	copy(s.JWTSecret[:], jck[:])

	switch dbType {
	case DB_MAP:
		s.userStatistic, err = userdb.NewUserStatisticsMapDB()
	case DB_BOLT:
		s.userStatistic, err = userdb.NewUserStatisticsDB()
	case DB_MONGO:
		s.userStatistic, err = userdb.NewUserStatisticsMongoDB(uri, "revel", mongoRevelPass)
	}
	if err != nil {
		panic(fmt.Sprintf("Could not create user statistic database %s", err.Error()))
	}

	s.poloniexCache = NewPoloniexAccessCache()

	switch dbType {
	case DB_MAP:
		s.userInviteCodes = userdb.NewInviteMapDB()
	case DB_BOLT:
		s.userInviteCodes = userdb.NewInviteDB()
	case DB_MONGO:
		//todo
		fallthrough
	default:
		s.userInviteCodes = userdb.NewInviteDB()
	}

	switch dbType {
	case DB_MAP:
		s.paymentDB, err = payment.NewPaymentDatabaseMap(uri, "revel", mongoRevelPass)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to user mongodb: %s\n", err.Error()))
		}
	case DB_BOLT:
		fallthrough
	case DB_MONGO:
		s.paymentDB, err = payment.NewPaymentDatabase(uri, "revel", mongoRevelPass)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to user mongodb: %s\n", err.Error()))
		}
	default:
		s.paymentDB, err = payment.NewPaymentDatabase(uri, "revel", mongoRevelPass)
		if err != nil {
			panic(fmt.Sprintf("Error connecting to user mongodb: %s\n", err.Error()))
		}
	}

	// SWITCHED TO BEES
	// s.Master = NewMaster()
	// s.Master.Run(6667)

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
	s.userDB.Close()
	s.userStatistic.Close()
	s.userInviteCodes.Close()
	return nil
}

func (s *State) SetAllUserMinimumLoan(username, mins string, exchange userdb.UserExchange) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	switch exchange {
	case userdb.PoloniexExchange:
		var coinsMin userdb.PoloniexMiniumumLendStruct
		err := json.Unmarshal([]byte(mins), &coinsMin)
		if err != nil {
			return fmt.Errorf("Poloniex: %s", err.Error())
		}
		u.PoloniexMiniumLend.SetAll(coinsMin)
		break
	case userdb.BitfinexExchange:
		var coinsMin userdb.BitfinexMiniumumLendStruct
		err := json.Unmarshal([]byte(mins), &coinsMin)
		if err != nil {
			return fmt.Errorf("Bitfinex: %s", err.Error())
		}
		u.BitfinexMiniumumLend.SetAll(coinsMin)
		break
	default:
		return fmt.Errorf("Exchange not recognized: %s", exchange)
	}

	return s.userDB.PutUser(u)
}

func (s *State) SetUserMinimumLoan(username string, minimumAmt float64, currency string) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.PoloniexMiniumLend.Set(currency, minimumAmt)
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
		log.Errorf("[NewUser - 5] Failed: %s :: %v", err.Error(), u)
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
	if code == "" {
		return false, fmt.Errorf("Code cannot be length 0")
	}
	return s.userInviteCodes.ClaimInviteCode(username, code)
}

func (s *State) AddInviteCode(code string, capacity int, expires time.Time) error {
	if code == "" {
		return fmt.Errorf("Code cannot be length 0")
	}
	return s.userInviteCodes.CreateInviteCode(code, capacity, expires)
}

func (s *State) SetUserKeys(username, acessKey, secretKey string, exchange userdb.UserExchange) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return fmt.Errorf("There was an internal problem. Please log out and try again. If problems continue, contact Support@hodl.zone")
	}

	pk, err := userdb.NewExchangeKeys(acessKey, secretKey, u.GetCipherKey(s.CipherKey))
	if err != nil {
		return fmt.Errorf("There was a problem setting your keys. Please double check the keys and try again. Contact Support@hodl.zone if the problem persists")
	}

	switch exchange {
	case userdb.PoloniexExchange:
		if len(secretKey) != 128 {
			return fmt.Errorf("Your secret key must be 128 characters long, found %d characters", len(secretKey))
		}
		u.PoloniexKeys = pk

		// _, err = s.PoloniexGetBalances(username)
		// if err != nil {
		// 	return fmt.Errorf("There was an error using your Poloniex keys. There is a chance they are not valid. Try setting them again, and if this continues contact Support@hodl.zone")
		// }
	case userdb.BitfinexExchange:
		u.BitfinexKeys = pk
	default:
		return fmt.Errorf("Exchange not recognized: %s", exchange)
	}

	err = s.userDB.PutUser(u)
	if err != nil {
		return fmt.Errorf("There was a database error. Please try again in a few minutes, then contact Support@hodl.zon")
	}

	return nil
}

func (s *State) GetUserStatistics(username string, dayRange int, u string) ([][]userdb.AllUserStatistic, error) {
	ptr := new(userdb.UserExchange)
	if u == "" {
		ptr = nil
	} else if u == "bit" {
		*ptr = userdb.BitfinexExchange
	} else {
		*ptr = userdb.PoloniexExchange
	}
	return s.userStatistic.GetStatistics(username, dayRange, ptr)
}

func (s *State) EnableUserLending(username string, c string, exchange userdb.UserExchange) error {
	u, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return err
	}
	switch exchange {
	case userdb.PoloniexExchange:
		var coins userdb.PoloniexEnabledStruct
		err := json.Unmarshal([]byte(c), &coins)
		if err != nil {
			return err
		}
		u.PoloniexEnabled.Enable(coins)
		break
	case userdb.BitfinexExchange:
		var coins userdb.BitfinexEnabledStruct
		err := json.Unmarshal([]byte(c), &coins)
		if err != nil {
			return err
		}
		u.BitfinexEnabled.Enable(coins)
		break
	default:
		return fmt.Errorf("Exchange not recognized: %s", exchange)
	}
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

func (s *State) GetQuickPoloniexStatistics(currency string) *userdb.PoloniexStats {
	return s.userStatistic.GetQuickPoloniexStatistics(currency)
}

func (s *State) GetPoloniexStatistics(currency string) (*userdb.PoloniexStats, error) {
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
func (s *State) RecordStatistics(stats *userdb.AllUserStatistic) error {
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
	Username          string                 `json:"email"`
	Privilege         string                 `json:"priv"`
	PoloniexEnabled   []userdb.EnabledStruct `json:"enabled"`
	PoloniexMiniumums []float64              `json:"minimum"`
}

func (s *State) GetAllUsers() (*[]SafeUser, error) {
	users, err := s.userDB.FetchAllUsers()
	if err != nil {
		return nil, fmt.Errorf("ERROR: Error getting all users: %s\n", err.Error())
	}
	safeUsers := make([]SafeUser, len(users), len(users))
	for i, u := range users {
		safeUsers[i] = SafeUser{
			u.Username,
			u.GetLevelString(),
			u.PoloniexEnabled.GetAll(),
			u.PoloniexMiniumLend.GetAll(),
		}
	}
	return &safeUsers, nil
}

func (s *State) DeleteUser() error {
	//TODO DELETE USER
	return nil
}

func (s *State) DeleteInvite(code string) error {
	return s.userInviteCodes.ExpireInviteCode(code)
}

func (s *State) UpdateUserPrivilege(email string, priv string) (*string, error) {
	u, err := s.userDB.FetchUserIfFound(email)
	if err != nil {
		return nil, err
	}
	u.Level = userdb.StringToLevel(priv)

	userLevelString := userdb.LevelToString(u.Level)
	return &userLevelString, s.userDB.PutUser(u)
}

func (s *State) SaveLendingHistory(lendHist *userdb.AllLendingHistoryEntry) error {
	return s.userStatistic.SaveLendingHistory(lendHist)
}

func (s *State) LoadLendingSummary(username string, t time.Time) (*userdb.AllLendingHistoryEntry, error) {
	data, err := s.userStatistic.GetLendHistorySummary(username, t)
	if err != nil {
		return data, err
	}
	data.Pop()
	na := time.Time{}
	if data.Time == na {
		data.SetTime(t)
	}
	return data, err
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

func (s *State) GetActivityLog(email string, timeString string) (*[]userdb.BotActivityLogEntry, error) {
	if timeString == "" {
		botAct, err := s.userStatistic.GetBotActivity(email)
		if err != nil {
			return nil, err
		}
		if botAct == nil {
			b := make([]userdb.BotActivityLogEntry, 0, 0)
			return &b, nil
		}
		return botAct.ActivityLog, nil
	}

	botActLogs, err := s.userStatistic.GetBotActivityTimeGreater(email, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return botActLogs, nil
}

func (s *State) SetUserReferee(username, refereeCode string) *primitives.ApiError {
	//calls get payment status to set referral code automatically if status does not exist
	exists, err := s.paymentDB.ReferralCodeExists(refereeCode)
	if err != nil {
		errMes := fmt.Errorf("Error checking referral code exists: %s", err.Error())
		return primitives.NewAPIErrorInternalError(errMes)
	}
	if !exists {
		return &primitives.ApiError{
			fmt.Errorf("RefereeCode [%s] does not exist", refereeCode),
			fmt.Errorf("The referee code entered does not exist."),
		}
	}

	status, err := s.GetPaymentStatus(username)
	if err != nil {
		errMes := fmt.Errorf("Error getting payment status: %s", err.Error())
		return primitives.NewAPIErrorInternalError(errMes)
	}

	if status.RefereeCode != "" {
		return &primitives.ApiError{
			fmt.Errorf("RefereeCode for user[%s] already set", username),
			fmt.Errorf("Your referee code has already been set."),
		}
	}

	status.RefereeCode = refereeCode

	err = s.paymentDB.SetStatus(*status)
	if err != nil {
		errMes := fmt.Errorf("Error setting status: %s", err.Error())
		return primitives.NewAPIErrorInternalError(errMes)
	}
	return nil
}

func (s *State) GetUserReferrals(username string) ([]payment.Status, error) {
	return s.paymentDB.GetUserReferrals(username)
}

func (s *State) GetPaymentDebtHistory(username string, limit int) ([]payment.Debt, error) {
	return s.paymentDB.GetDebtsLimitSortIfFound(username, 2, limit)
}

func (s *State) GetPaymentPaidHistory(username, dateAfterStr string) ([]payment.Paid, error) {
	var p []payment.Paid
	if dateAfterStr == "" {
		//get all payments
		return s.paymentDB.GetAllPaid(username, nil)
	}
	//parse time
	layout := "2017-07-31T18:35:34.970Z"
	dateAfter, err := time.Parse(layout, dateAfterStr)
	if err != nil {
		return p, fmt.Errorf("Failed to parse dateAfter: %s", err.Error())
	}
	return s.paymentDB.GetAllPaid(username, &dateAfter)
}

// Gets payment status, if none is found than will create new status with predefined
// referral code and username.
func (s *State) GetPaymentStatus(username string) (*payment.Status, error) {
	status, err := s.paymentDB.GetStatusIfFound(username)
	if err != nil {
		return nil, fmt.Errorf("Failed to get status: %s", err.Error())
	} else if status == nil {
		status, err = s.paymentDB.GenerateReferralCode(username)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate referral code: %s", err.Error())
		}
	}
	return status, nil
}

//returns true if referee code has been set
func (s *State) HasSetReferee(username string) bool {
	status, err := s.paymentDB.GetStatusIfFound(username)
	if err != nil || status == nil {
		return false
	}
	return status.RefereeCode != ""
}

func (s *State) MakePayment(username string, paid payment.Paid) error {
	err := s.paymentDB.AddPaid(paid)
	if err != nil {
		return fmt.Errorf("Error adding pay debt: %s", err.Error())
	}
	err = s.paymentDB.PayDebts(username, paid)
	if err != nil {
		return fmt.Errorf("Error paying debts: %s", err.Error())
	}
	return nil
}
