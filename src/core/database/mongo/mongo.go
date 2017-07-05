package mongo

import (
	"gopkg.in/mgo.v2"
)

const (
	//USER BEGIN
	USER_DB      = "userdb"
	USER_DB_TEST = "userdb_test"

	C_USER = "user"
	//AUDIT END

	//STAT BEGIN
	STAT_DB      = "statdb"
	STAT_DB_TEST = "statdb_test"

	C_UserStat_POL = "poloniexUserStat"
	C_LendHist_POL = "poloniexLendingHist"
	C_Exchange_POL = "poloniexExchange"

	C_UserStat_BIT = "bitfinexUserStat"
	C_LendHist_BIT = "bitfinexLendingHist"
	C_Exchange_BIT = "bitfinexExchange"
	//STAT END

	//AUDIT BEGIN
	AUDIT_DB      = "auditdb"
	AUDIT_DB_TEST = "auditdb_test"

	C_Audit = "audit"
	//AUDIT END
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
	if c.baseSession == nil {
		session, err := mgo.Dial(c.uri)
		if err != nil {
			return nil, err
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
	db := createMongoDB(uri, USER_DB, dbu, dbp)

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

	c := session.DB(USER_DB).C(C_USER)

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

func CreateStatDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, STAT_DB, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(STAT_DB).C(C_UserStat_POL)

	var index mgo.Index
	// index := mgo.Index{
	// 	Key:        []string{"_id.time"},
	// 	Unique:     true,
	// 	DropDups:   true,
	// 	Background: true,
	// 	Sparse:     true,
	// }

	// err = c.EnsureIndex(index)
	// if err != nil {
	// 	return nil, err
	// }

	c = session.DB(STAT_DB).C(C_LendHist_POL)
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

	// c = session.DB(STAT_DB).C(C_Exchange_POL)
	// index = mgo.Index{
	// 	Key:        []string{},
	// 	Unique:     true,
	// 	DropDups:   true,
	// 	Background: true,
	// 	Sparse:     true,
	// }

	// err = c.EnsureIndex(index)
	// if err != nil {
	// 	return nil, err
	// }
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

func CreateTestStatDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, STAT_DB_TEST, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(STAT_DB_TEST).C(C_UserStat_POL)

	var index mgo.Index
	// index := mgo.Index{
	// 	Key:        []string{"_id.time"},
	// 	Unique:     true,
	// 	DropDups:   true,
	// 	Background: true,
	// 	Sparse:     true,
	// }

	// err = c.EnsureIndex(index)
	// if err != nil {
	// 	return nil, err
	// }

	c = session.DB(STAT_DB_TEST).C(C_LendHist_POL)
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

	// c = session.DB(STAT_DB_TEST).C(C_Exchange_POL)
	// index = mgo.Index{
	// 	Key:        []string{},
	// 	Unique:     true,
	// 	DropDups:   true,
	// 	Background: true,
	// 	Sparse:     true,
	// }

	// err = c.EnsureIndex(index)
	// if err != nil {
	// 	return nil, err
	// }
	return db, nil
}
