package tests

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/payment"
	"github.com/revel/revel/cache"
)

type HasReferee struct {
	Ref bool `json:"ref"`
}

func (t *AppTest) TestUserReferee() {
	cache.Flush()
	t.AssertEqual(nil, resetUserDB())
	t.AssertEqual(nil, resetPaymentDB())
	// setup
	testUser := register(t, "test@hodl.zone", "pass")
	test1User := register(t, "test@hodll.zone", "pass")
	login(t, testUser)
	t.Get("/dashboard/data/paymenthistory")
	t.AssertOk()
	logout(t)
	login(t, test1User)
	t.Get("/dashboard/data/paymenthistory")
	t.AssertOk()
	logout(t)
	login(t, testUser)
	status, err := getUserStatus("test@hodl.zone")
	t.AssertEqual(nil, err)
	t.AssertEqual(status.ReferralCode, "test")
	status, err = getUserStatus("test@hodll.zone")
	t.AssertEqual(nil, err)
	t.AssertEqual(status.ReferralCode, "test0")
	// /setup

	var hasRef HasReferee

	t.Get("/dashboard/settings/hasreferee")
	t.AssertOk()
	json.Unmarshal(t.ResponseBody, &hasRef)
	t.AssertEqual(hasRef.Ref, false)

	//trying to set to oneself
	v := url.Values{}
	v.Set("ref", "test")
	t.PostForm("/dashboard/settings/setreferee", v)
	t.AssertStatus(500)
	status, err = getUserStatus("test@hodl.zone")
	t.AssertEqual(nil, err)
	t.AssertEqual(status.RefereeCode, "")

	//set to another user
	v.Set("ref", "test0")
	t.PostForm("/dashboard/settings/setreferee", v)
	t.AssertOk()
	status, err = getUserStatus("test@hodl.zone")
	t.AssertEqual(nil, err)
	t.AssertEqual(status.RefereeCode, "test0")

	//try to set to another user
	v.Set("ref", "test0")
	t.PostForm("/dashboard/settings/setreferee", v)
	t.AssertStatus(500)
	status, err = getUserStatus("test@hodl.zone")
	t.AssertEqual(nil, err)
	t.AssertEqual(status.RefereeCode, "test0")
}

func register(t *AppTest, username, pass string) url.Values {
	v := url.Values{}
	v.Set("email", username)
	v.Set("pass", pass)
	v.Set("ic", "testcode")
	t.PostForm("/register", v)
	t.AssertOk()
	return v
}

func getUserStatus(email string) (*payment.Status, error) {
	dbGiven, err := mongo.CreateTestPaymentDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, fmt.Errorf("Error getting userdb: %s", err.Error())
	}
	p := payment.NewPaymentDatabaseGiven(dbGiven)
	return p.GetStatus(email)
}

func resetPaymentDB() error {
	_, err := payment.NewPaymentDatabaseEmpty("mongodb://localhost:27017", "", "")
	return err
}

// func (t *AppTest) TestUserReferee() {

// }
