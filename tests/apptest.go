package tests

import (
	"fmt"
	// "net/http"
	"net/url"
	"time"

	"github.com/Emyrk/LendingBot/app/controllers"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/revel/revel/cache"
	// "github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel/testing"
)

var (
	expTime = 50 * time.Millisecond
	format  = "2006-01-02 15:04:05.000"
)

type AppTest struct {
	testing.TestSuite
}

func (t *AppTest) Before() {
	println("Set up")
}

func (t *AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

// func (t *AppTest) TestEmail() {

// r := email.NewHTMLRequest(email.SMTP_EMAIL_NO_REPLY, []string{
// 	"stevenmasley@gmail.com",
// 	"masley.dean@gmail.com",
// }, "This is a test email")

// 	err := r.ParseTemplate("test.html", struct {
// 		NameOne string
// 		NameTwo string
// 	}{
// 		"steve",
// 		"dean",
// 	})
// 	t.AssertEqual(false, err != nil)

// 	err = r.SendEmail()
// 	t.AssertEqual(false, err != nil)
// }

func addCode(t *AppTest) {
	v := url.Values{}
	v.Set("email", "admin@admin.com")
	v.Set("pass", "admin")
	login(t, v)

	v = url.Values{}
	v.Set("rawc", "testcode")
	v.Set("cap", fmt.Sprintf("%d", 10000))
	v.Set("hr", fmt.Sprintf("%d", 20))
	t.PostForm("/dashboard/sysadmin/makeinvite", v)

	logout(t)
}

func (t *AppTest) TestRegister() {
	t.AssertEqual(resetUserDB(), nil)
	t.AssertEqual(resetStatDB(), nil)
	cache.Flush()

	//just add it on first run dont care about error
	addCode(t)

	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")

	v := GetDefaultLoginValues()
	v.Set("ic", "testcode")
	t.PostForm("/register", v)
	t.AssertOk()
	t.AssertContentType("application/json; charset=utf-8")

	//check it was added to db
	users, err := getAllUsers()
	t.AssertEqual(nil, err)
	t.AssertEqual(2, len(users))

	//check that session is at count 1
	ses, err := getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 1)
	us := getUserSessionWithEmail(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 1)
	t.AssertEqual(us.ChangeState[0].SessionAction, userdb.OPENED)
	t.AssertEqual(us.Open, true)

	//check session cache
	var cacheSes controllers.CacheSession
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertEqual(nil, err)
	t.AssertEqual(1, len(cacheSes.Sessions))
	_, ok := cacheSes.Sessions[t.Session.ID()]
	t.AssertEqual(ok, true)

	logout(t)

	//check that session is at count 2
	ses, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 1)
	us = getUserSessionWithEmail(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 2)
	t.AssertEqual(us.ChangeState[0].SessionAction, userdb.OPENED)
	t.AssertEqual(us.ChangeState[1].SessionAction, userdb.CLOSED)
	t.AssertEqual(us.Open, false)

	//check session cache
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertNotEqual(nil, err)
}

func (t *AppTest) TestLoginLogout() {
	t.TestRegister()

	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")

	t.PostForm("/login", GetDefaultLoginValues())
	t.AssertOk()
	t.AssertContentType("application/json; charset=utf-8")

	//check that session is at count 3
	//because after register adds default 2
	ses, err := getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 1)
	us := getUserSessionWithEmail(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 3)
	t.AssertEqual(us.ChangeState[2].SessionAction, userdb.REOPENED)
	t.AssertEqual(us.Open, true)

	//check session cache
	var cacheSes controllers.CacheSession
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertEqual(nil, err)
	t.AssertEqual(1, len(cacheSes.Sessions))
	_, ok := cacheSes.Sessions[t.Session.ID()]
	t.AssertEqual(ok, true)

	t.Get("/logout")
	t.AssertOk()

	//check that session is at count 4
	ses, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 1)
	us = getUserSessionWithEmail(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 4)
	t.AssertEqual(us.ChangeState[3].SessionAction, userdb.CLOSED)
	t.AssertEqual(us.Open, false)

	//check session cache
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertNotEqual(nil, err)
}

func (t *AppTest) TestSetAndTimeoutExpiry() {
	t.TestRegister() //+2 session count, total = 2

	login(t, GetDefaultLoginValues()) //+1 session count, total = 3

	v := url.Values{}
	v.Set("sesexp", fmt.Sprintf("%d", expTime))
	t.PostForm("/dashboard/settings/changeexpiry", v)
	t.AssertOk()

	//validate timeout was changed
	u, err := getUser(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(nil, err)
	t.AssertEqual(u.SessionExpiryTime, expTime)
	////check session cache
	var cacheSes controllers.CacheSession
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertEqual(1, len(cacheSes.Sessions))
	t.AssertEqual(cacheSes.Expiry, expTime)

	//test that under expire time will result in success
	time.Sleep(expTime / 2)
	t.Get("/dashboard")
	t.AssertOk()

	//wait for timeout
	time.Sleep(expTime)

	//should error out because of invalid session
	t.Get("/dashboard") //+1 session count, total = 4
	t.AssertStatus(403)

	//session should be closed
	ses, err := getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 1)
	us := getUserSessionWithEmail(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 4)
	t.AssertEqual(us.ChangeState[3].SessionAction, userdb.CLOSED)
	t.AssertEqual(us.Open, false)

	//cache should be deleted
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.Assertf(nil != err, "Error should be empty", cacheSes)
}

func (t *AppTest) TestMultiSessionExpiry() {
	t.TestSetAndTimeoutExpiry() //+4 session count, OrigTestSuite total = 4

	// expTime := 50 * time.Millisecond

	login(t, GetDefaultLoginValues()) //+1 session count, OrigTestSuite total = 5
	t.Get("/dashboard")
	t.AssertOk()

	//create separate request for separate session
	otherTestSuite := AppTest{testing.NewTestSuite()}
	login(&otherTestSuite, GetDefaultLoginValues()) //+1 session count, OtherTestSuite total 1

	//session ids should be different
	t.AssertNotEqual(otherTestSuite.Session.ID(), t.Session.ID())

	///////////////
	//validate two separate sessions were created both in cache and in mongo
	///////////////
	ses, err := getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 2)
	origSes := getUserSessionWithId(ses, t.Session.ID())
	otherSes := getUserSessionWithId(ses, otherTestSuite.Session.ID())
	//check that OrigTestSuite session is at count 5
	t.AssertEqual(len(origSes.ChangeState), 5)
	t.AssertEqual(origSes.ChangeState[4].SessionAction, userdb.REOPENED)
	t.AssertEqual(origSes.Open, true)
	//check that OrigTestSuite session is at count 1
	t.AssertEqual(len(otherSes.ChangeState), 1)
	t.AssertEqual(otherSes.ChangeState[0].SessionAction, userdb.OPENED)
	t.AssertEqual(otherSes.Open, true)
	//check session cache
	var cacheSes controllers.CacheSession
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertEqual(nil, err)
	t.AssertEqual(2, len(cacheSes.Sessions))
	origTime, ok := cacheSes.Sessions[t.Session.ID()]
	t.AssertEqual(ok, true)
	t.AssertEqual(origSes.LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT), origTime.UTC().Format(userdb.SESSION_FORMAT))
	otherTime, ok := cacheSes.Sessions[otherTestSuite.Session.ID()]
	t.AssertEqual(ok, true)
	t.AssertEqual(otherSes.LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT), otherTime.UTC().Format(userdb.SESSION_FORMAT))

	///////////////
	//refresh both sessions assure that they are updated
	/////////////// making request
	t.Get("/dashboard") //+0 session count, OrigTestSuite total = 5
	t.AssertOk()
	otherTestSuite.Get("/dashboard") //+0 session count, OtherTestSuite total 1
	otherTestSuite.AssertOk()
	//validating change in time and session
	ses, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 2)
	origSes = getUserSessionWithId(ses, t.Session.ID())
	otherSes = getUserSessionWithId(ses, otherTestSuite.Session.ID())
	//check that OrigTestSuite session is still at count 5
	t.AssertEqual(len(origSes.ChangeState), 5)
	t.AssertEqual(origSes.ChangeState[4].SessionAction, userdb.REOPENED)
	t.AssertEqual(origSes.Open, true)
	//check that OrigTestSuite session is still at count 1
	t.AssertEqual(len(otherSes.ChangeState), 1)
	t.AssertEqual(otherSes.ChangeState[0].SessionAction, userdb.OPENED)
	t.AssertEqual(otherSes.Open, true)
	//check session cache
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertEqual(nil, err)
	t.AssertEqual(2, len(cacheSes.Sessions))
	origTime, ok = cacheSes.Sessions[t.Session.ID()]
	t.AssertEqual(ok, true)
	t.AssertEqual(origSes.LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT), origTime.UTC().Format(userdb.SESSION_FORMAT))
	otherTime, ok = cacheSes.Sessions[otherTestSuite.Session.ID()]
	t.AssertEqual(ok, true)
	t.AssertEqual(otherSes.LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT), otherTime.UTC().Format(userdb.SESSION_FORMAT))

	///////////////
	//refresh orig session timeout refresh original
	///////////////
	otherTestSuite.Get("/dashboard") //+0 session count, OtherTestSuite total 1
	otherTestSuite.AssertOk()
	//validating change in time and session
	ses, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 2)
	t.AssertEqual(getUserSessionWithId(ses, t.Session.ID()).LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT), origSes.LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT))
	t.AssertNotEqual(getUserSessionWithId(ses, otherTestSuite.Session.ID()).LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT), otherSes.LastRenewalTime.UTC().Format(userdb.SESSION_FORMAT))
	//check session cache
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.AssertEqual(nil, err)
	t.AssertEqual(2, len(cacheSes.Sessions))
	beforeTime, ok := cacheSes.Sessions[t.Session.ID()]
	t.AssertEqual(ok, true)
	_, ok = cacheSes.Sessions[otherTestSuite.Session.ID()]
	t.AssertEqual(ok, true)
	//sleep until orig session is too late
	time.Sleep(time.Duration(expTime.Nanoseconds() - time.Since(beforeTime).Nanoseconds()))
	//it should fail for orig but not other test
	t.Get("/dashboard")
	t.AssertStatus(403)
	otherTestSuite.Get("/dashboard")
	otherTestSuite.AssertOk()

	///////////////
	//logout other session
	///////////////
	otherTestSuite.Get("/logout")
	otherTestSuite.AssertOk()
	//session should be closed
	ses, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 2)
	us := getUserSessionWithId(ses, t.Session.ID())
	t.AssertEqual(len(us.ChangeState), 6)
	t.AssertEqual(us.ChangeState[5].SessionAction, userdb.CLOSED)
	t.AssertEqual(us.Open, false)
	us = getUserSessionWithId(ses, otherTestSuite.Session.ID())
	t.AssertEqual(len(us.ChangeState), 2)
	t.AssertEqual(us.ChangeState[1].SessionAction, userdb.CLOSED)
	t.AssertEqual(us.Open, false)

	//cache should be deleted
	err = cache.Get(GetDefaultLoginValues().Get("email"), &cacheSes)
	t.Assertf(nil != err, "Error should be empty", cacheSes)

}

func login(t *AppTest, v url.Values) {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")

	t.PostForm("/login", v)
	t.AssertOk()
	t.AssertContentType("application/json; charset=utf-8")
}

func logout(t *AppTest) {
	t.Get("/logout")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func GetDefaultLoginValues() url.Values {
	v := url.Values{}
	v.Set("email", "test@hodl.zone")
	v.Set("pass", "pass")
	return v
}

func resetUserDB() error {
	_, err := mongo.CreateBlankTestUserDB("127.0.0.1:27017", "", "")
	return err
}

func resetStatDB() error {
	_, err := mongo.CreateBlankTestStatDB("127.0.0.1:27017", "", "")
	return err
}

//will only return first one
func getUserSessionWithId(ses []userdb.Session, sessionId string) *userdb.Session {
	for i, s := range ses {
		if s.SessionId == sessionId {
			return &ses[i]
		}
	}
	return nil
}

func getUserSessionWithEmail(ses []userdb.Session, email string) *userdb.Session {
	for i, s := range ses {
		if s.Email == email {
			return &ses[i]
		}
	}
	return nil
}

func getAllUserSessions(email string) ([]userdb.Session, error) {
	dbGiven, err := mongo.CreateTestUserDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, fmt.Errorf("Error getting userdb: %s", err.Error())
	}
	ud := userdb.NewMongoUserDatabaseGiven(dbGiven)
	ses, err := ud.GetAllUserSessions(email, 0, 1000)
	return *ses, err
}

func getAllUsers() ([]userdb.User, error) {
	dbGiven, err := mongo.CreateTestUserDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, err
	}
	ud := userdb.NewMongoUserDatabaseGiven(dbGiven)
	return ud.FetchAllUsers()
}

func getUser(email string) (*userdb.User, error) {
	dbGiven, err := mongo.CreateTestUserDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, err
	}
	ud := userdb.NewMongoUserDatabaseGiven(dbGiven)
	return ud.FetchUser(email)
}

func (t *AppTest) After() {
	println("Tear down")
}
