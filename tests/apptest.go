package tests

import (
	"bytes"
	"encoding/json"
	"github.com/Emyrk/LendingBot/app/controllers"
	"github.com/Emyrk/LendingBot/src/core/email"
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

func (t *AppTest) After() {
	println("Tear down")
}

func (t *AppTest) TestEmail() {

	r := email.NewHTMLRequest(email.SMTP_EMAIL_USER, []string{
		"stevenmasley@gmail.com",
		"masley.dean@gmail.com",
	}, "This is a test email")

	err := r.ParseTemplate("test.html", struct {
		NameOne string
		NameTwo string
	}{
		"steve",
		"dean",
	})
	t.AssertEqual(false, err != nil)

	err = r.SendEmail()
	t.AssertEqual(false, err != nil)
}

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
