package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Emyrk/LendingBot/app/controllers"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	// "github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"github.com/revel/revel/testing"
	"gopkg.in/mgo.v2/bson"
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

func (t *AppTest) TestRegister() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")

	json, err := json.Marshal(controllers.JSONUser{
		"test@hodl.zone",
		"testpass",
	})
	t.AssertEqual(false, err != nil)
	reader := bytes.NewReader([]byte(json))
	t.Post("/register", "application/json; charset=utf-8", reader)
	t.AssertOk()
	t.AssertContentType("application/json; charset=utf-8")
}

func (t *AppTest) TestSession() {
	err := setupSessionDB()
	if err != nil {
		t.Assertf(true, fmt.Sprintf("Error starting up db: %s", err.Error()))
	}
	uss, err := getAllUserSessions("test")
	if err != nil {
		t.Assertf(false, fmt.Sprintf("Error getting all users: %s", err.Error()))
	}
	t.Assertf(len(uss) == 0, fmt.Sprintf("Error length of users is not 0 is %d", len(uss)))

	//login user
	t.Get("/")
	t.AssertOk()
	t.PostForm("/login", GetDefaultLoginValues())
	t.AssertOk()

	//validating user session was opened
	uss, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	if err != nil {
		t.Assertf(false, fmt.Sprintf("Error getting all users: %s", err.Error()))
	}
	t.Assertf(len(uss) == 1, fmt.Sprintf("Error length of users is not 1 is %d", len(uss)))
	loginSes := uss[0]
	if loginSes.Open != true || loginSes.Email != GetDefaultLoginValues().Get("email") {
		t.Assertf(false, fmt.Sprintf("Error login session doesnt match: %s", loginSes))
	}
	// /login user

	//logout
	t.Get("/logout")
	t.AssertOk()
	uss, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	if err != nil {
		t.Assertf(false, fmt.Sprintf("Error getting all users: %s", err.Error()))
	}
	t.Assertf(len(uss) == 1, fmt.Sprintf("Error length of users is not 1 is %d", len(uss)))
	if uss[0].Open != false {
		t.Assertf(false, fmt.Sprintf("Error open should be false: %d", uss[0].Open))
	}
	if len(uss[0].ChangeState) != 2 {
		t.Assertf(false, fmt.Sprintf("Error should be 2 is %d", len(uss[0].ChangeState)))
	}
	if uss[0].ChangeState[0].SessionAction != userdb.OPENED {
		t.Assertf(false, fmt.Sprintf("Error first state should be OPENED: %d", uss[0].ChangeState[0].SessionAction))
	}
	if uss[0].ChangeState[1].SessionAction != userdb.CLOSED {
		t.Assertf(false, fmt.Sprintf("Error first state should be CLOSED: %d", uss[1].ChangeState[0].SessionAction))
	}
	// /logout user

	//login user and timeout
	t.PostForm("/login", GetDefaultLoginValues())
	t.AssertOk()
	uss, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	if err != nil {
		t.Assertf(false, fmt.Sprintf("Error getting all users: %s", err.Error()))
	}
	t.Assertf(len(uss) == 1, fmt.Sprintf("Error length of users is not 1 is %d", len(uss)))
	if uss[0].Open != true {
		t.Assertf(false, fmt.Sprintf("Error open should be false: %v", uss[0].Open))
	}
	if len(uss[0].ChangeState) != 3 {
		t.Assertf(false, fmt.Sprintf("Error should be 3 is %d", len(uss[0].ChangeState)))
	}
	if uss[0].ChangeState[0].SessionAction != userdb.OPENED {
		t.Assertf(false, fmt.Sprintf("Error first state should be OPENED: %d", uss[0].ChangeState[0].SessionAction))
	}
	if uss[0].ChangeState[1].SessionAction != userdb.CLOSED {
		t.Assertf(false, fmt.Sprintf("Error first state should be CLOSED: %d", uss[1].ChangeState[0].SessionAction))
	}
	if uss[0].ChangeState[2].SessionAction != userdb.REOPENED {
		t.Assertf(false, fmt.Sprintf("Error first state should be REOPEND: %d", uss[2].ChangeState[0].SessionAction))
	}

	time.Sleep(2000 * time.Millisecond)

	t.Get("/dashboard")
	t.AssertStatus(403)
	uss, err = getAllUserSessions(GetDefaultLoginValues().Get("email"))
	if err != nil {
		t.Assertf(false, fmt.Sprintf("Error getting all users: %s", err.Error()))
	}
	t.Assertf(len(uss) == 1, fmt.Sprintf("Error length of users is not 1 is %d", len(uss)))
	if uss[0].Open != false {
		t.Assertf(false, fmt.Sprintf("Error open should be false: %v", uss[0].Open))
	}
	if len(uss[0].ChangeState) != 4 {
		t.Assertf(false, fmt.Sprintf("Error should be 4 is %d", len(uss[0].ChangeState)))
	}
	if uss[0].ChangeState[0].SessionAction != userdb.OPENED {
		t.Assertf(false, fmt.Sprintf("Error first state should be OPENED: %d", uss[0].ChangeState[0].SessionAction))
	}
	if uss[0].ChangeState[1].SessionAction != userdb.CLOSED {
		t.Assertf(false, fmt.Sprintf("Error first state should be CLOSED: %d", uss[1].ChangeState[0].SessionAction))
	}
	if uss[0].ChangeState[2].SessionAction != userdb.REOPENED {
		t.Assertf(false, fmt.Sprintf("Error first state should be REOPENED: %d", uss[2].ChangeState[0].SessionAction))
	}
	if uss[0].ChangeState[3].SessionAction != userdb.CLOSED {
		t.Assertf(false, fmt.Sprintf("Error first state should be CLOSED: %d", uss[3].ChangeState[0].SessionAction))
	}
	// /login user and timeout
}

func GetDefaultLoginValues() url.Values {
	v := url.Values{}
	v.Set("email", "test")
	v.Set("pass", "pass")
	return v
}

func setupSessionDB() error {
	nsNotFoundErr := errors.New("ns not found")

	db, err := mongo.CreateTestUserDB("127.0.0.1:27017", "", "")
	if err != nil {
		return fmt.Errorf("Error setting up userdb: %s", err.Error())
	}
	s, c, err := db.GetCollection(mongo.C_USER)
	if err != nil {
		return fmt.Errorf("createSession: %s", err.Error())
	}
	_, err = c.RemoveAll(bson.M{})
	if err != nil && err.Error() != nsNotFoundErr.Error() {
		return fmt.Errorf("Error removing all users: %s", err.Error())
	}

	u, err := userdb.NewUser("test", "pass")
	if err != nil {
		return err
	}
	u.SessionExpiryTime = 1 * time.Second
	err = c.Insert(u)
	if err != nil {
		return fmt.Errorf("Error adding test user: %s", err.Error())
	}
	s.Close()
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

func (t *AppTest) After() {
	println("Tear down")
}

// func (t *AppTest) TestLoginLogout() {
// 	t.Get("/")
// 	t.AssertOk()
// 	t.AssertContentType("text/html; charset=utf-8")

// 	json, err := json.Marshal(controllers.JSONUser{
// 		"test@hodl.zone",
// 		"testpass",
// 	})
// 	t.AssertEqual(false, err != nil)
// 	reader := bytes.NewReader([]byte(json))
// 	t.Post("/login", "application/json; charset=utf-8", reader)
// 	t.AssertOk()

// 	t.Get("/logout")
// 	t.AssertOk()

// 	t.Get("/dashboard")
// 	t.AssertOk()
// 	url, err := t.Response.Location()
// 	t.AssertEqual(false, err != nil)
// 	t.AssertEqual("/", url.Path)
// }

// func (t *AppTest) TestLoginTimeout() {
// 	t.Get("/")
// 	t.AssertOk()
// 	t.AssertContentType("text/html; charset=utf-8")

// 	json, err := json.Marshal(controllers.JSONUser{
// 		"test@hodl.zone",
// 		"testpass",
// 	})
// 	t.AssertEqual(false, err != nil)
// 	reader := bytes.NewReader([]byte(json))
// 	t.Post("/login", "application/json; charset=utf-8", reader)
// 	t.AssertOk()

// 	time.Sleep(3 * time.Second)

// 	t.Get("/dashboard")
// 	t.AssertOk()
// 	url, err := t.Response.Location()
// 	t.AssertEqual(false, err != nil)
// }
