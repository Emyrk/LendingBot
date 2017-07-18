package tests

import (
	// "bytes"
	// "encoding/json"
	"fmt"
	// "net/url"
	// "time"
	// "github.com/Emyrk/LendingBot/app/controllers"
	// "github.com/Emyrk/LendingBot/src/core/database/mongo"
	// "github.com/Emyrk/LendingBot/src/core/email"
	// "github.com/Emyrk/LendingBot/src/core/userdb"
	// "github.com/revel/revel/testing"
	// "gopkg.in/mgo.v2/bson"
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

func (t *AppTest) After() {
	println("Tear down")
}

// func (t *AppTest) TestEmail() {

// 	r := email.NewHTMLRequest(email.SMTP_EMAIL_USER, []string{
// 		"stevenmasley@gmail.com",
// 		"masley.dean@gmail.com",
// 	}, "This is a test email")

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

// func (t *AppTest) TestRegister() {
// 	t.Get("/")
// 	t.AssertOk()
// 	t.AssertContentType("text/html; charset=utf-8")

// 	json, err := json.Marshal(controllers.JSONUser{
// 		"test@hodl.zone",
// 		"testpass",
// 	})
// 	t.AssertEqual(false, err != nil)
// 	reader := bytes.NewReader([]byte(json))
// 	t.Post("/register", "application/json; charset=utf-8", reader)
// 	t.AssertOk()
// 	t.AssertContentType("application/json; charset=utf-8")
// }

// func (t *AppTest) TestSession() {
// 	err := SetupSessionDB()
// 	t.Assertf(err != nil, fmt.Sprintf("Error starting up db: %s", err.Error()))
// 	uss, err := GetAllUserSessions("test")
// 	t.Assertf(err != nil, fmt.Sprintf("Error getting all users: %s", err.Error()))
// 	t.Assertf(len(*uss) != 1, fmt.Sprintf("Error length of users is not 1 is %d", len(*uss)))

// 	//login user
// 	t.Get("/")
// 	t.AssertOk()
// 	t.PostForm("/login", GetLoginValues())
// 	t.AssertOk()
// 	// /login user

// 	//logout user
// 	t.Get("/logout")
// 	t.AssertOk()
// 	// /logout user
// }

// func GetLoginValues() url.Values {
// 	v := url.Values{}
// 	v.Set("email", "test")
// 	v.Set("pass", "pass")
// 	return v
// }

// func SetupSessionDB() error {
// 	nsNotFoundErr := errors.New("ns not found")

// 	db, err := mongo.CreateTestUserDB("127.0.0.1:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
// 	if err != nil {
// 		return fmt.Errorf("Error setting up userdb: %s", err.Error())
// 	}
// 	s, c, err := db.GetCollection(mongo.C_USER)
// 	if err != nil {
// 		return fmt.Errorf("createSession: %s", err.Error())
// 	}
// 	_, err = c.RemoveAll(bson.M{})
// 	if err != nil && err.Error() != nsNotFoundErr.Error() {
// 		return fmt.Errorf("Error removing all users: %s", err.Error())
// 	}

// 	u := userdb.NewUser("test", "pass")
// 	u.SessionExpiryTime = 1 * time.Minute
// 	err = c.Insert(u)
// 	if err != nil {
// 		return fmt.Errorf("Error adding test user: %s", err.Error())
// 	}

// 	s, c, err = db.GetCollection(mongo.C_Session)
// 	if err != nil {
// 		return fmt.Errorf("createSession: %s", err.Error())
// 	}
// 	_, err = c.RemoveAll(bson.M{})
// 	if err != nil && err.Error() != nsNotFoundErr.Error() {
// 		return fmt.Errorf("Error removing all users: %s", err.Error())
// 	}
// 	s.Close()
// 	return nil
// }

// func GetAllUserSessions(email string) (*[]Session, error) {
// 	ud, err := userdb.NewMongoUserDatabase("127.0.0.1:27017", "revel", os.Getenv("MONGO_REVEL_PASS"))
// 	if err != nil {
// 		return fmt.Errorf("Error creating get all user serssion: %s", err.Error())
// 	}
// 	return ud.GetAllUserSessions(email, 0, 1000)
// }

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
