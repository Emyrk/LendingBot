package balancer

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/slack"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/sirupsen/logrus"
)

var auditLogger = instanceLogger.WithFields(log.Fields{"package": "auditor"})

// Auditor performs an audit
//		An Audit:
//			- Grabs all users from DB
//			- Ensures all users are allocated to a bee
//			- Ensures all users have been touched within 2min
//			- Builds a Report for front end
//				- Each slave and their notes
type Auditor struct {
	ConnectionPool *Hive

	Report string

	CipherKey [32]byte

	auditDB *mongo.MongoDB
	userDB  *mongo.MongoDB

	performing bool
}

type AuditUser struct {
	User     userdb.User
	Exchange int

	// Used in one routine
	hits int
}

func NewAuditor(h *Hive, uri string, dbu string, dbp string, cipherkey [32]byte) *Auditor {
	a := new(Auditor)
	a.CipherKey = cipherkey
	a.ConnectionPool = h

	var err error
	//TODO JESSE
	a.auditDB, err = mongo.CreateAuditDB(uri, dbu, dbp)
	if err != nil {
		slack.SendMessage(":rage:", "hive", "alerts", fmt.Sprintf("@channel Auditor for hive: Oy!.. failed to connect to the mongodb, I am panicing! Error: %s", err.Error()))
		panic(fmt.Sprintf("Failed to connect to db: %s", err.Error()))
	}
	//TODO JESSE
	a.userDB, err = mongo.CreateUserDB(uri, dbu, dbp)
	if err != nil {
		slack.SendMessage(":rage:", "hive", "alerts", fmt.Sprintf("@channel Auditor for hive: Oy!.. failed to connect to the mongodb, I am panicing! Error: %s", err.Error()))
		panic(fmt.Sprintf("Failed to connect to db: %s", err.Error()))
	}
	return a
}

type AuditReport struct {
	UsersInDB      []string
	Bees           []string
	UserNotes      []string
	UserLogsReport map[string]*UserLogs
	CorrectionList []AuditUser
	NoExtensive    bool
	Time           time.Time `bson:"_id"`
}

func (a *AuditReport) String() string {
	str := fmt.Sprintf("======= Audit Report =======\n")
	str += fmt.Sprintf("   %-20s\n   %-20s : %d\n   %-20s : %d\n   %-20s : %s\n   %-20s : %d\n   %-20s : %d\n",
		"Summary",
		"Bees", len(a.Bees),
		"Corrections", len(a.CorrectionList),
		"Time", a.Time,
		"Users In DB", len(a.UsersInDB),
		"Users+Exch Active", len(a.UserLogsReport))

	str += "  ===== Bees =====  \n"
	for _, b := range a.Bees {
		str += fmt.Sprintf(" - %s\n", b)
	}

	str += "  ===== Notes =====  \n"
	for _, n := range a.UserNotes {
		str += fmt.Sprintf(" - %s\n", n)
	}

	str += "  ===== Users From DB =====  \n"
	for _, u := range a.UsersInDB {
		str += fmt.Sprintf(" - %s\n", u)
	}

	str += "  ===== User Logs ===== \n"
	t := len(a.UserLogsReport)
	c := 1
	for k, l := range a.UserLogsReport {
		str += fmt.Sprintf("---------- User: %s, %d/%d ----------\n", k, c, t)
		str += l.String()
		c++
	}

	return str
}

type UserLogs struct {
	Healthy   bool
	LastTouch time.Time
	SlaveID   string
	Logs      string
}

func (l *UserLogs) String() string {
	str := fmt.Sprintf("%-15s : %t\n", "Healthy", l.Healthy)
	str += fmt.Sprintf("%-15s : %s\n", "LastTouch", l.LastTouch)
	str += fmt.Sprintf("%-15s : %s\n", "SlaveID", l.SlaveID)
	str += fmt.Sprintf("%-15s \n%s\n", "Logs", l.Logs)
	return str
}

func (a *Auditor) PerformAudit() *AuditReport {
	start := time.Now()
	if a.performing {
		return nil
	}
	a.performing = true
	defer func() {
		a.performing = false
	}()

	flog := auditLogger.WithField("func", "PerformAudit")
	flog.Infof("Starting audit")

	ar := new(AuditReport)
	var correct []AuditUser

	bees := a.ConnectionPool.Slaves.GetAndLockAllBees(true)
	for _, b := range bees {
		ustr := ""
		for _, u := range b.Users {
			ustr += fmt.Sprintf("[%s|%s],", u.Username, GetExchangeString(u.Exchange))
			ar.UserNotes = append(ar.UserNotes, fmt.Sprintf("[%s|%s] LastTouch: %s, LastSave: %s\n    %s", u.Username, GetExchangeString(u.Exchange), u.LastTouch, u.LastHistorySaved, u.Notes))
		}
		ar.Bees = append(ar.Bees, fmt.Sprintf("[%s] Users: %d (%s)", b.ID, len(b.Users), ustr))
	}
	a.ConnectionPool.Slaves.RUnlock()

	all, err := a.GetAllFullUsers()
	if err != nil {
		auditLogger.WithFields(log.Fields{"func": "PerformAudit"}).Errorf("Error retreiving full users: %s", err.Error())
		return nil
	}
	if all == nil {
		return nil
	}

	logs := make(map[string]*UserLogs)
	// Cycle through all users in the database
	for _, u := range all {
		var exchs []int
		// Currency pairs are enabled
		if len(u.PoloniexEnabled.Keys()) > 0 {
			exchs = append(exchs, PoloniexExchange)
		}
		if len(u.BitfinexEnabled.Keys()) > 0 {
			exchs = append(exchs, BitfinexExchange)
		}

		pkeyStr := ""
		for _, k := range u.PoloniexEnabled.Keys() {
			pkeyStr += k + ", "
		}
		bkeyStr := ""
		for _, k := range u.BitfinexEnabled.Keys() {
			bkeyStr += k + ", "
		}
		ar.UsersInDB = append(ar.UsersInDB, fmt.Sprintf("%s [Poloniex: %s] [Bitfinex: %s]", u.Username, pkeyStr))
		for _, e := range exchs {
			// We keep logs on every user, even if successful
			logs[u.Username] = new(UserLogs)
			logs[u.Username].SlaveID = "Unknown"
			id, ok := a.ConnectionPool.Slaves.GetUser(u.Username, e)
			if !ok {
				// User was not found in the slavepool

				// Get the user with keys
				balus, err := a.UserDBUserToBalancerUser(&u, e)
				if err != nil {
					logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was not found to be working. Was unable to get api keys: %s\n",
						time.Now(), u.Username, GetExchangeString(e), err)
					continue
				}
				// User was not found in a slave. Allocate this user
				err = a.ConnectionPool.AddUser(balus)
				if err != nil {
					logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was not found to be working. Was unable to allocate to bee: %s\n",
						time.Now(), u.Username, GetExchangeString(e), err)
					correct = append(correct, AuditUser{User: u, Exchange: e})
				} else {
					logs[u.Username].Logs += fmt.Sprintf("%s [Warning] %s on %s was not found to be working. Was allocated to a bee, and maybe resolved\n",
						time.Now(), u.Username, GetExchangeString(e))
				}
			} else {
				// User was found
				bee, ok := a.ConnectionPool.Slaves.GetAndLockBee(id, true)
				if !ok {
					// User was found, but the bee it was allocated to is not.
					logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was found, but the bee it was allocated too was not.\n",
						time.Now(), u.Username, GetExchangeString(e))
					correct = append(correct, AuditUser{User: u, Exchange: e})
				} else {
					// Bee and user found
					logs[u.Username].SlaveID = bee.ID
					found := false
					// Verify user
					for _, bu := range bee.Users {
						// Everything looks good
						if u.Username == bu.Username && e == bu.Exchange {
							logs[u.Username].Logs += fmt.Sprintf("%s [INFO] %s on %s  was last touched %s\n",
								time.Now(), u.Username, GetExchangeString(e), bu.LastTouch)
							logs[u.Username].Healthy = true
							logs[u.Username].LastTouch = bu.LastTouch
							found = true
							break
						}
					}
					if !found {
						logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was found, but the bee [%s] it was allocated does not seem to have it.\n",
							time.Now(), u.Username, GetExchangeString(e), bee.ID)
						correct = append(correct, AuditUser{User: u, Exchange: e})
					}
				}
				a.ConnectionPool.Slaves.RUnlock()
			}
		}
	}

	// ExtensiveCorrect
	nochanges := a.ExtensiveSearchAndCorrect(correct, logs)

	ar.UserLogsReport = logs
	ar.CorrectionList = correct
	ar.NoExtensive = nochanges
	ar.Time = time.Now().UTC()

	flog.WithFields(log.Fields{"corrections": len(correct)}).Infof("Audit performed in %fs.", time.Since(start).Seconds())
	return ar
}

func (a *Auditor) UserDBUserToBalancerUser(u *userdb.User, exch int) (*User, error) {
	balUser := new(User)
	balUser.Username = u.Username
	var err error
	switch exch {
	case PoloniexExchange:
		if u.PoloniexKeys.APIKeyEmpty() {
			return nil, fmt.Errorf("no API key for this exchange")
		}
		balUser.AccessKey, err = u.PoloniexKeys.DecryptAPIKeyString(u.GetCipherKey(a.CipherKey))
		if err != nil {
			return nil, err
		}

		balUser.SecretKey, err = u.PoloniexKeys.DecryptAPISecretString(u.GetCipherKey(a.CipherKey))
		if err != nil {
			return nil, err
		}
	case BitfinexExchange:
		if u.BitfinexKeys.APIKeyEmpty() {
			return nil, fmt.Errorf("no API key for this exchange")
		}
		balUser.AccessKey, err = u.BitfinexKeys.DecryptAPIKeyString(u.GetCipherKey(a.CipherKey))
		if err != nil {
			return nil, err
		}

		balUser.SecretKey, err = u.BitfinexKeys.DecryptAPISecretString(u.GetCipherKey(a.CipherKey))
		if err != nil {
			return nil, err
		}
	}
	return balUser, nil
}

func (a *Auditor) GetFullUser(username string, exchange int) (*User, error) {
	s, c, err := a.auditDB.GetCollection(mongo.C_USER)
	if err != nil {
		return nil, fmt.Errorf("GetUsers: getCol: %s", err.Error())
	}
	defer s.Close()

	var result userdb.User
	err = c.FindId(username).One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetUsers: find: %s", err.Error())
	}

	result.PoloniexKeys.SetEmptyIfBlank()
	result.BitfinexKeys.SetEmptyIfBlank()

	return a.UserDBUserToBalancerUser(&result, exchange)
}

func (a *Auditor) GetAllFullUsers() ([]userdb.User, error) {
	s, c, err := a.auditDB.GetCollection(mongo.C_USER)
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers: getCol: %s", err.Error())
	}
	defer s.Close()

	var results []userdb.User
	err = c.Find(nil).All(&results)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers: find: %s", err.Error())
	}

	//need to blank out the poloniex stuff to appease embedded database
	users := make([]userdb.User, len(results), len(results))
	for i, u := range results {
		u.PoloniexKeys.SetEmptyIfBlank()
		u.BitfinexKeys.SetEmptyIfBlank()
		users[i] = u
	}
	return users, nil
}

// ExtensiveSearchAndCorrect will go through every bee and correct any users given
func (a *Auditor) ExtensiveSearchAndCorrect(correct []AuditUser, userlogs map[string]*UserLogs) bool {
	// These users have an issue and will be corrected
	if len(correct) == 0 {
		return true
	}

	fix := make(map[string]*AuditUser)
	for _, u := range correct {
		fix[u.User.Username] = &u
	}

	// Look for 2 bees having the same user
	allusers := make(map[string]string)

	bees := a.ConnectionPool.Slaves.GetAllBees()
	for _, b := range bees {
		for _, bu := range b.Users {
			if e, ok := fix[bu.Username]; ok {
				if e.Exchange == bu.Exchange {
					// We found the user and their bee. Fix the usermap and report
					fix[bu.Username].hits++
					if fix[bu.Username].hits > 1 {
						// 2 bees have this user! Remove from second bee
						userlogs[bu.Username].Logs += fmt.Sprintf("%s [REPAIR-ERROR-COR] %s on %s was found on another bee [%s]. Will remove from this bee as it is on another\n",
							time.Now(), bu.Username, GetExchangeString(bu.Exchange), b.ID)
						err := a.ConnectionPool.RemoveUserFromBee(b.ID, bu.Username, bu.Exchange)
						if err != nil {
							userlogs[bu.Username].Logs += fmt.Sprintf("%s [REPAIR-ERROR-COR] %s on %s was had an error being removed from bee [%s]: %s\n",
								time.Now(), bu.Username, GetExchangeString(bu.Exchange), b.ID, err.Error())
						}
					} else {
						// Correct usermap
						userlogs[bu.Username].Logs += fmt.Sprintf("%s [REPAIR-COR] %s on %s was found at bee [%s]. It was not found in the usermap. We will add to the usermap\n",
							time.Now(), bu.Username, GetExchangeString(bu.Exchange), b.ID)
						a.ConnectionPool.Slaves.AddUser(bu.Username, bu.Exchange, b.ID)
					}
				}
			} else {
				_, ok := allusers[bu.Username]
				if ok {
					// Duplicate users. Delete one
					userlogs[bu.Username].Logs += fmt.Sprintf("%s [REPAIR-ERROR] %s on %s was found on another bee [%s]. It should be only on bee [%s]. Will remove from this bee as it is on another\n",
						time.Now(), bu.Username, GetExchangeString(bu.Exchange), b.ID, allusers[bu.Username])
					err := a.ConnectionPool.RemoveUserFromBee(b.ID, bu.Username, bu.Exchange)
					if err != nil {
						userlogs[bu.Username].Logs += fmt.Sprintf("%s [REPAIR-ERROR] %s on %s was unable to be removed from bee[%s]: %s\n",
							time.Now(), bu.Username, GetExchangeString(bu.Exchange), b.ID, err)
					}
				} else {
					allusers[bu.Username] = b.ID
				}
			}
		}
	}

	return false
}

func (a *Auditor) SaveAudit(auditReport *AuditReport) error {
	s, c, err := a.auditDB.GetCollection(mongo.AUDIT_DB)
	if err != nil {
		return fmt.Errorf("Mongo cannot save audit: %s", err.Error())
	}
	defer s.Close()

	upsertKey := bson.M{"_id": auditReport.Time}
	upsertAction := bson.M{"$set": auditReport}
	_, err = c.Upsert(upsertKey, upsertAction)
	if err != nil {
		return fmt.Errorf("Mongo failed to upsert: %s", err)
	}
	return nil
}
