package tests

import (
	"fmt"
	"net/url"
	"time"

	"github.com/Emyrk/LendingBot/app/controllers"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/revel/revel/cache"
	// "github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel/testing"
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
	fmt.Println(t.Response.StatusCode)

	logout(t)
}

func (t *AppTest) TestRegister() {
	t.AssertEqual(resetUserDB(), nil)
	t.AssertEqual(resetStatDB(), nil)

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
	users, err := getAllUsers(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(nil, err)
	t.AssertEqual(2, len(users))

	//check that session is at count 1
	ses, err := getAllUserSessions(GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(ses), 1)
	us := getUserSession(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 1)
	t.AssertEqual(us.ChangeState[0].SessionAction, userdb.OPENED)

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
	us = getUserSession(ses, GetDefaultLoginValues().Get("email"))
	t.AssertEqual(len(us.ChangeState), 2)
	t.AssertEqual(us.ChangeState[0].SessionAction, userdb.OPENED)
	t.AssertEqual(us.ChangeState[1].SessionAction, userdb.CLOSED)

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

	t.Get("/logout")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
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

func (t *AppTest) TestSetExpiry() {
	t.TestRegister()

	login(t, GetDefaultLoginValues())
	v := url.Values{}
	v.Set("sesexp", fmt.Sprintf("%d", 500*time.Millisecond))
	t.PostForm("/dashboard/settings/changeexpiry", v)
	t.AssertOk()
	logout(t)
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

func getUserSession(ses []userdb.Session, email string) *userdb.Session {
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

func getAllUsers(email string) ([]userdb.User, error) {
	dbGiven, err := mongo.CreateTestUserDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, err
	}
	ud := userdb.NewMongoUserDatabaseGiven(dbGiven)
	return ud.FetchAllUsers()
}

func (t *AppTest) After() {
	println("Tear down")
}
