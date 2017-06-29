package balancer

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/slack"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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

	auditDB *mongo.MongoDB
	userDB  *mongo.MongoDB
}

type AuditUser struct {
	Username string
	Exchange int

	// Used in one routine
	hits int
}

func NewAuditor(h *Hive, uri string, dbu string, dbp string) *Auditor {
	a := new(Auditor)
	a.ConnectionPool = h

	var err error
	a.auditDB, err = mongo.CreateAuditDB(uri, dbu, dbp)
	if err != nil {
		slack.SendMessage(":rage:", "hive", "alerts", fmt.Sprintf("@channel Auditor for hive: Oy!.. failed to connect to the mongodb, I am panicing! Error: %s", err.Error()))
		panic(fmt.Sprintf("Failed to connect to db: %s", err.Error()))
	}
	a.userDB, err = mongo.CreateUserDB(uri, dbu, dbp)
	if err != nil {
		slack.SendMessage(":rage:", "hive", "alerts", fmt.Sprintf("@channel Auditor for hive: Oy!.. failed to connect to the mongodb, I am panicing! Error: %s", err.Error()))
		panic(fmt.Sprintf("Failed to connect to db: %s", err.Error()))
	}
	return a
}

type AuditReport struct {
	UserLogsReport map[string]*UserLogs
	CorrectionList []*AuditUser
	NoExtensive    bool
	Time           time.Time `bson:"_id"`
}

type UserLogs struct {
	Healthy   bool
	LastTouch time.Time
	SlaveID   string
	Logs      string
}

func (a *Auditor) PerformAudit() *AuditReport {
	var correct []userdb.User
	all, err := a.GetAllFullUsers()
	if err != nil {
		//TODO
		fmt.Println("Error retrieving full uesrs: %s\n", err.Error())
		return nil
	}
	if all == nil {
		return nil
	}
	logs := make(map[string]*UserLogs)
	for _, u := range *all {
		var exchs []int
		if len(u.PoloniexEnabled.Keys()) > 0 {
			exchs = append(exchs, PoloniexExchange)
		}
		// if len(u.BinfinexEnabled.Keys()) > 0 {
		// 	exchs = append(exchs, PoloniexExchange)
		// }
		for _, e := range exchs {
			logs[u.Username] = new(UserLogs)
			logs[u.Username].SlaveID = "Unknown"
			id, ok := a.ConnectionPool.Slaves.GetUser(u.Username, e)
			if !ok {
				// User was not found in a slave. Allocate this user
				err := a.ConnectionPool.AddUser(nil) //GetFullUser(u.Username, e)) TODO
				if err != nil {
					logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was not found to be working. Was unable to allocate to bee: %s\n",
						time.Now(), u.Username, GetExchangeString(e), err)
					correct = append(correct, u)
				} else {
					logs[u.Username].Logs += fmt.Sprintf("%s [Warning] %s on %s was not found to be working. Was allocated to a bee\n",
						time.Now(), u.Username, GetExchangeString(e), err)
				}
			} else {
				// User was found
				bee, ok := a.ConnectionPool.Slaves.GetAndLockBee(id, true)
				if !ok {
					// User was found, but the bee it was allocated to is not.
					logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was found, but the bee it was allocated too was not.\n",
						time.Now(), u.Username, GetExchangeString(e))
					correct = append(correct, u)
				} else {
					logs[u.Username].SlaveID = bee.ID
					found := false
					for _, bu := range bee.Users {
						// Everything looks good
						if u.Username == bu.Username && e == bu.Exchange {
							logs[u.Username].Logs += fmt.Sprintf("%s [INFO] %s on %s  was last touched %s\n",
								time.Now(), u.Username, GetExchangeString(e), bu.LastTouch)
							logs[u.Username].Healthy = true
							logs[u.Username].LastTouch = bu.LastTouch
							found = true
						}
						break
					}
					if !found {
						logs[u.Username].Logs += fmt.Sprintf("%s [ERROR] %s on %s was found, but the bee [%s] it was allocated does not seem to have it.\n",
							time.Now(), u.Username, GetExchangeString(e), bee.ID)
						correct = append(correct, u)
					}
				}
			}
		}
	}

	// ExtensiveCorrect
	nochanges := a.ExtensiveSearchAndCorrect(nil, logs) //correct, logs) TODO

	ar := new(AuditReport)
	ar.UserLogsReport = logs
	ar.CorrectionList = nil //correct TODO
	ar.NoExtensive = nochanges
	ar.Time = time.Now().UTC()

	return ar
}

// func (a *Auditor) GetFullUser(username string, exchange int) *User {
// 	s, c, err := a.auditDB.GetCollection(mongo.C_USER)
// 	if err != nil {
// 		return nil, fmt.Errorf("GetAllUsers: getCol: %s", err.Error())
// 	}
// 	defer s.Close()

// 	var result User
// 	err = c.FindId(username).One(&result)
// 	if err == mgo.ErrNotFound {
// 		return nil, nil
// 	}
// 	if err != nil {
// 		return nil, fmt.Errorf("GetAllUsers: find: %s", err.Error())
// 	}

// 	result.PoloniexKeys.SetEmptyIfBlank()
// 	return &result, nil
// }

func (a *Auditor) GetAllFullUsers() (*[]userdb.User, error) {
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
		users[i] = u
	}
	return &users, nil
}

// ExtensiveSearchAndCorrect will go through every bee and correct any users given
func (a *Auditor) ExtensiveSearchAndCorrect(correct []*AuditUser, userlogs map[string]*UserLogs) bool {
	// These users have an issue and will be corrected
	if len(correct) == 0 {
		return true
	}

	fix := make(map[string]*AuditUser)
	for _, u := range correct {
		fix[u.Username] = u
	}

	// Look for 2 bees having the same user
	allusers := make(map[string]string)

	for _, b := range a.ConnectionPool.Slaves.GetAndLockAllBees(true) {
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
