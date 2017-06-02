package tests

import (
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
