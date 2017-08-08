package mongo

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var _ = fmt.Println

const (
	//USER BEGIN
	USER_DB      = "userdb"
	USER_DB_TEST = "userdb_test"

	C_USER    = "user"
	C_Session = "session"
	//AUDIT END

	//STAT BEGIN
	STAT_DB      = "statdb"
	STAT_DB_TEST = "statdb_test"

	C_UserStat = "userStat"
	C_LendHist = "lendingHist"

	C_Exchange_POL = "poloniexExchange"
	C_Exchange_BIT = "bitfinexExchange"

	C_BotActivity = "botActivity"
	//STAT END

	//AUDIT BEGIN
	AUDIT_DB      = "auditdb"
	AUDIT_DB_TEST = "auditdb_test"

	C_Audit = "audit"
	//AUDIT END

	ADMIN_DB = "admin"
)

type MongoDB struct {
	uri         string
	DbName      string
	dbusername  string
	dbpass      string
	baseSession *mgo.Session
}

func (c *MongoDB) GetURI() string {
	return c.uri
}

func createMongoDB(uri string, dbname, dbu, dbp string) *MongoDB {
	mongoDB := &MongoDB{
		uri,
		dbname,
		dbu,
		dbp,
		nil,
	}
	return mongoDB
}

func (c *MongoDB) CreateSession() (*mgo.Session, error) {
	var err error
	if c.baseSession == nil {
		var session *mgo.Session
		if len(c.dbusername) > 0 && len(c.dbpass) > 0 {
			dialInfo := &mgo.DialInfo{
				Addrs:    []string{c.uri},
				Database: ADMIN_DB,
				Username: c.dbusername,
				Password: c.dbpass,
				DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
					return tls.Dial("tcp", addr.String(), &tls.Config{})
				},
				Timeout: time.Second * 10,
			}
			session, err = mgo.DialWithInfo(dialInfo)
			if err != nil {
				return nil, err
			}
		} else {
			session, err = mgo.Dial(c.uri)
			if err != nil {
				return nil, err
			}
		}
		c.baseSession = session

		// See https://godoc.org/labix.org/v2/mgo#Session.SetMode
		c.baseSession.SetMode(mgo.Monotonic, true)
	}

	return c.baseSession.Clone(), nil
}

func (c *MongoDB) GetCollection(collectionName string) (*mgo.Session, *mgo.Collection, error) {
	session, err := c.CreateSession()
	if err != nil {
		return nil, nil, err
	}

	return session, session.DB(c.DbName).C(collectionName), nil
}

func CreateAuditDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, AUDIT_DB, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// c := session.DB(AUDIT_DB).C(C_Audit)

	return db, nil
}

func CreateUserDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, USER_DB, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// c := session.DB(USER_DB).C(C_USER)

	return db, nil
}

func CreateStatDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, STAT_DB, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return db, nil
}

func CreateTestUserDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, USER_DB_TEST, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(USER_DB_TEST).C(C_USER)

	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateBlankTestUserDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, USER_DB_TEST, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(USER_DB_TEST).C(C_USER)

	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		return nil, err
	}

	//remove all but admin user
	_, err = c.RemoveAll(bson.M{"_id": bson.M{"$ne": "admin@admin.com"}})
	if err != nil {
		return nil, err
	}

	c = session.DB(USER_DB_TEST).C(C_Session)

	_, err = c.RemoveAll(bson.M{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTestStatDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, STAT_DB_TEST, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(STAT_DB_TEST).C(C_Session)

	var index mgo.Index
	index = mgo.Index{
		Key:        []string{"email"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateBlankTestStatDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, STAT_DB_TEST, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	err = session.DB(STAT_DB_TEST).DropDatabase()
	if err != nil {
		return nil, err
	}

	c := session.DB(STAT_DB_TEST).C(C_Session)

	var index mgo.Index
	index = mgo.Index{
		Key:        []string{"email"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		return nil, err
	}

	return db, nil
}
