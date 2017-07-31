package payment_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "github.com/Emyrk/LendingBot/src/core/payment"
)

func Test_user_referral(t *testing.T) {
	testUser := "test@hodl.zone"
	userReferrals := []string{"admin@admin.com", "roger@houston.com"}
	db, err := getPaymentDBAndClearCollection(mongo.C_Status)
	if err != nil {
		t.Error(err)
	}

	//add user referral, one
	err = db.AddUserReferral(testUser, userReferrals[0])
	if err != nil {
		t.Error(err)
	}
	referrals, err := db.GetUserReferralsIfFound(testUser)
	if err != nil {
		t.Error(err)
	}
	if len(referrals) != 1 {
		t.Errorf("Should have 1 referral, is %d", len(referrals))
	} else if referrals[0].Username != userReferrals[0] {
		t.Errorf("Referral name is not correct [%s]!=[%s]", referrals[0])
	}

	//add 2nd referral
	err = db.AddUserReferral(testUser, userReferrals[1])
	if err != nil {
		t.Error(err)
	}
	referrals, err = db.GetUserReferralsIfFound(testUser)
	if err != nil {
		t.Error(err)
	}
	if len(referrals) != 2 {
		t.Errorf("Should have 2 referral, is %d", len(referrals))
	} else if referrals[0].Username != userReferrals[0] {
		t.Errorf("Referral name is not correct [%s]!=[%s]", referrals[0])
	} else if referrals[1].Username != userReferrals[1] {
		t.Errorf("Referral name is not correct [%s]!=[%s]", referrals[1])
	}

	//test fail to add yourself as referral
	err = db.AddUserReferral(testUser, testUser)
	if err == nil {
		t.Errorf("Should not be able to add yourself as referee.")
	}
	referrals, err = db.GetUserReferralsIfFound(testUser)
	if err != nil {
		t.Error(err)
	}
	if len(referrals) != 2 {
		t.Errorf("Should still have 2 referral, is %d", len(referrals))
	}
}

func Test_user_referee(t *testing.T) {
	testUser := "test@hodl.zone"
	userReferee := "admin@admin.com"
	db, err := getPaymentDBAndClearCollection(mongo.C_Status)
	if err != nil {
		t.Error(err)
	}

	//set user referee to ones self should fail
	err = db.SetUserReferee(testUser, testUser)
	if err == nil {
		t.Errorf("Should fail to set oneself as user referee")
	}

	//set user referee
	err = db.SetUserReferee(testUser, userReferee)
	if err == nil {
		t.Error(err)
	}

	//set again should fail because already set
	err = db.SetUserReferee(testUser, userReferee)
	if err == nil {
		t.Errorf("Should fail to change referee")
	}
}

func Test_user_debt(t *testing.T) {
	testUser := "test@hodl.zone"
	db, err := getPaymentDBAndClearCollection(mongo.C_Debt)
	if err != nil {
		t.Error(err)
	}

	//create multidebt
	rd := make([]Debt, 10, 10)
	for i := 0; i < len(rd); i++ {
		if i < len(rd)/2 {
			rd[i] = newRandomizedDebt(testUser, true)
		} else {

			rd[i] = newRandomizedDebt(testUser, false)
		}
	}
	err = db.SetMultiDebt(rd)
	if err != nil {
		t.Error(err)
	}

	//get all debts
	debts, err := db.GetAllDebts(testUser, 2)
	if len(debts) == 5 {
		t.Errorf("Should get 5 back is %d", len(debts))
	}
	debts, err = db.GetAllDebts(testUser, 1)
	if len(debts) == 5 {
		t.Errorf("Should get 5 back is %d", len(debts))
	}
	debts, err = db.GetAllDebts(testUser, 0)
	if len(debts) == 10 {
		t.Errorf("Should get 10 back is %d", len(debts))
	}

	//add in blank one to test if it adds
	debt := newRandomizedDebt(testUser, true)
	err = db.SetMultiDebt([]Debt{debt})
	if err != nil {
		t.Error(err)
	}
	debts, err = db.GetAllDebts(testUser, 2)
	if len(debts) == 5 {
		t.Errorf("Should get 5 back is %d", len(debts))
	}
	debts, err = db.GetAllDebts(testUser, 1)
	if len(debts) == 6 {
		t.Errorf("Should get 6 back is %d", len(debts))
	}
	debts, err = db.GetAllDebts(testUser, 0)
	if len(debts) == 11 {
		t.Errorf("Should get 11 back is %d", len(debts))
	}
}

func Test_user_paid(t *testing.T) {
	testUser := "test@hodl.zone"
	db, err := getPaymentDBAndClearCollection(mongo.C_Paid)
	if err != nil {
		t.Error(err)
	}

	//create multipaid
	rd := make([]Paid, 10, 10)
	now := time.Now()
	for i := 0; i < len(rd); i++ {
		rd[i] = newRandomizedPaid(testUser, now.Add(time.Duration(i)*time.Second))
	}
	err = db.SetMultiPaid(rd)
	if err != nil {
		t.Error(err)
	}

	//get all paid
	paids, err := db.GetAllPaid(testUser, &rd[4].PaymentDate)
	if len(paids) == 5 {
		t.Errorf("Should get 5 back is %d", len(paids))
	}
	paids, err = db.GetAllPaid(testUser, nil)
	if len(paids) == 10 {
		t.Errorf("Should get 10 back is %d", len(paids))
	}

	//add in blank one to test if it adds
	paid := newRandomizedPaid(testUser, now.Add(time.Duration(10)*time.Second))
	err = db.SetMultiPaid([]Paid{paid})
	if err != nil {
		t.Error(err)
	}

	//get all paid
	paids, err = db.GetAllPaid(testUser, &rd[4].PaymentDate)
	if len(paids) == 6 {
		t.Errorf("Should get 6 back is %d", len(paids))
	}
	paids, err = db.GetAllPaid(testUser, nil)
	if len(paids) == 11 {
		t.Errorf("Should get 11 back is %d", len(paids))
	}
}

func removeAllFromCollection(c *mgo.Collection) error {
	_, err := c.RemoveAll(bson.M{})
	if err != nil {
		return err
	}
	return nil
}

func getPaymentDB() (*PaymentDatabase, error) {
	db, err := mongo.CreateTestPaymentDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}
	return NewPaymentDatabaseGiven(db), nil
}

func getPaymentDBAndClearCollection(collectionName string) (*PaymentDatabase, error) {
	db, err := mongo.CreateTestPaymentDB("127.0.0.1:27017", "", "")
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}

	s, c, err := db.GetCollection(collectionName)
	if err != nil {
		return nil, fmt.Errorf("createSession: %s", err.Error())
	}
	defer s.Close()

	_, err = c.RemoveAll(bson.M{})
	if err != nil {
		return nil, fmt.Errorf("removeAll: %s", err.Error())
	}

	return NewPaymentDatabaseGiven(db), nil
}

//all currency in btc... did not randomize that
//did not randomize full paid
func newRandomizedDebt(username string, fp bool) Debt {
	rand.Seed(time.Now().UTC().UnixNano())
	return Debt{
		LoanDate:              time.Now().UTC(),
		Charge:                rand.Float64(),
		AmountLoaned:          rand.Float64(),
		LoanRate:              rand.Float64(),
		GrossAmountEarned:     rand.Float64(),
		Currency:              "BTC",
		CurrencyToBTC:         rand.Float64(),
		CurrencyToETH:         rand.Float64(),
		Exchange:              userdb.PoloniexExchange,
		Username:              username,
		FullPaid:              fp,
		PaymentPercentageRate: rand.Float64(),
	}
}

func newRandomizedPaid(username string, t time.Time) Paid {
	rand.Seed(time.Now().UTC().UnixNano())
	return Paid{
		PaymentDate:        t.UTC(),
		BTCPaid:            rand.Float64(),
		BTCTransactionDate: t.UTC(),
		BTCTransactionID:   rand.Int63(),
		ETHPaid:            rand.Float64(),
		ETHTransactionDate: t.UTC(),
		ETHTransactionID:   rand.Int63(),
		AddressPaidFrom:    "",
		Username:           username,
	}
}
