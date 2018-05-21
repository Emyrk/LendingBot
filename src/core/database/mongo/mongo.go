package mongo

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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

	//PAYMENT BEGIN DB
	PAYMENT_DB      = "paymentdb"
	PAYMENT_DB_TEST = "paymentdb_test"

	C_Status       = "status"
	C_Debt         = "debt"
	C_Paid         = "paid"
	C_CoinbaseCode = "coinbasecode"
	C_HODLZONECode = "hodlzonecode"
	C_PendingPaid  = "pendingPaid"
	//PAYMENT END

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

var letsencrypt_root_crt = `-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIRAIIQz7DSQONZRGPgu2OCiwAwDQYJKoZIhvcNAQELBQAw
TzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh
cmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMTUwNjA0MTEwNDM4
WhcNMzUwNjA0MTEwNDM4WjBPMQswCQYDVQQGEwJVUzEpMCcGA1UEChMgSW50ZXJu
ZXQgU2VjdXJpdHkgUmVzZWFyY2ggR3JvdXAxFTATBgNVBAMTDElTUkcgUm9vdCBY
MTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAK3oJHP0FDfzm54rVygc
h77ct984kIxuPOZXoHj3dcKi/vVqbvYATyjb3miGbESTtrFj/RQSa78f0uoxmyF+
0TM8ukj13Xnfs7j/EvEhmkvBioZxaUpmZmyPfjxwv60pIgbz5MDmgK7iS4+3mX6U
A5/TR5d8mUgjU+g4rk8Kb4Mu0UlXjIB0ttov0DiNewNwIRt18jA8+o+u3dpjq+sW
T8KOEUt+zwvo/7V3LvSye0rgTBIlDHCNAymg4VMk7BPZ7hm/ELNKjD+Jo2FR3qyH
B5T0Y3HsLuJvW5iB4YlcNHlsdu87kGJ55tukmi8mxdAQ4Q7e2RCOFvu396j3x+UC
B5iPNgiV5+I3lg02dZ77DnKxHZu8A/lJBdiB3QW0KtZB6awBdpUKD9jf1b0SHzUv
KBds0pjBqAlkd25HN7rOrFleaJ1/ctaJxQZBKT5ZPt0m9STJEadao0xAH0ahmbWn
OlFuhjuefXKnEgV4We0+UXgVCwOPjdAvBbI+e0ocS3MFEvzG6uBQE3xDk3SzynTn
jh8BCNAw1FtxNrQHusEwMFxIt4I7mKZ9YIqioymCzLq9gwQbooMDQaHWBfEbwrbw
qHyGO0aoSCqI3Haadr8faqU9GY/rOPNk3sgrDQoo//fb4hVC1CLQJ13hef4Y53CI
rU7m2Ys6xt0nUW7/vGT1M0NPAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNV
HRMBAf8EBTADAQH/MB0GA1UdDgQWBBR5tFnme7bl5AFzgAiIyBpY9umbbjANBgkq
hkiG9w0BAQsFAAOCAgEAVR9YqbyyqFDQDLHYGmkgJykIrGF1XIpu+ILlaS/V9lZL
ubhzEFnTIZd+50xx+7LSYK05qAvqFyFWhfFQDlnrzuBZ6brJFe+GnY+EgPbk6ZGQ
3BebYhtF8GaV0nxvwuo77x/Py9auJ/GpsMiu/X1+mvoiBOv/2X/qkSsisRcOj/KK
NFtY2PwByVS5uCbMiogziUwthDyC3+6WVwW6LLv3xLfHTjuCvjHIInNzktHCgKQ5
ORAzI4JMPJ+GslWYHb4phowim57iaztXOoJwTdwJx4nLCgdNbOhdjsnvzqvHu7Ur
TkXWStAmzOVyyghqpZXjFaH3pO3JLF+l+/+sKAIuvtd7u+Nxe5AW0wdeRlN8NwdC
jNPElpzVmbUq4JUagEiuTDkHzsxHpFKVK7q4+63SM1N95R1NbdWhscdCb+ZAJzVc
oyi3B43njTOQ5yOf+1CceWxG1bQVs5ZufpsMljq4Ui0/1lvh+wjChP4kqKOJ2qxq
4RgqsahDYVvTH9w7jXbyLeiNdd8XM2w9U/t7y0Ff/9yi0GE44Za4rF2LN9d11TPA
mRGunUHBcnWEvgJBQl9nJEiU0Zsnvgc/ubhPgXRR4Xq37Z0j4r7g1SgEEzwxA57d
emyPxgcYxn/eR44/KJ4EBs+lVDR3veyJm+kXQ99b21/+jh5Xos1AnX5iItreGCc=
-----END CERTIFICATE-----
`

func (c *MongoDB) CreateSession() (*mgo.Session, error) {
	var err error
	if c.baseSession == nil {
		var session *mgo.Session
		if len(c.dbusername) > 0 && len(c.dbpass) > 0 && c.dbpass != "MadeUpPass" {
			certs := x509.NewCertPool()
			certs.AppendCertsFromPEM([]byte(letsencrypt_root_crt))
			_ = certs
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
				return nil, fmt.Errorf("Error making TLS connection to server[%s]: %s", c.uri, err.Error())
			}
		} else {
			session, err = mgo.Dial(c.uri)
			if err != nil {
				return nil, fmt.Errorf("Error making UNSECURE connection to server[%s]: %s", c.uri, err.Error())
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

func CreatePaymentDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, PAYMENT_DB, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return db, nil
}

func CreateTestPaymentDB(uri, dbu, dbp string) (*MongoDB, error) {
	db := createMongoDB(uri, PAYMENT_DB_TEST, dbu, dbp)

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// c := session.DB(USER_DB_TEST).C(C_USER)

	// index := mgo.Index{
	// 	Key:        []string{"username"},
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
