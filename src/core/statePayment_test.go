package core_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/core/payment"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = fmt.Print

const (
	SATOSHI_FLOAT float64 = float64(100000000)
	SATOSHI_INT   int64   = int64(100000000)
	SATOSHI_5     int64   = int64(SATOSHI_FLOAT * 0.5)
)

func TestMakePayment_OneDebt_NoReferrals(t *testing.T) {
	state := NewStateWithMongoEmpty()
	//setup debt
	email := "test@hodl.zone"
	debt := payment.Debt{
		LoanDate:          time.Now(),
		Charge:            0.0,
		AmountLoaned:      int64(SATOSHI_FLOAT * 10.0),
		LoanRate:          0.5,
		GrossAmountEarned: int64(SATOSHI_FLOAT * 5.0),
		Currency:          "BTC",
		CurrencyToBTC:     int64(SATOSHI_FLOAT * 10.0),
		CurrencyToETH:     int64(SATOSHI_FLOAT * 100.0),
		Exchange:          userdb.PoloniexExchange,
		Username:          email,
		FullPaid:          false,
		PaymentPaidAmount: 0,
	}
	paymentdb, err := payment.NewPaymentDatabaseEmpty("mongodb://localhost:27017", "", "")
	if err != nil {
		t.Errorf("Error connecting to database: %s", err.Error())
	}
	err = paymentdb.InsertNewDebt(debt)
	if err != nil {
		t.Errorf("Error inserting new debt: %s", err.Error())
	}
	// /setup debt
	//validate debt
	debts, err := paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 1 {
		t.Errorf("Debts should have length 1 is %d", len(debts))
	}
	// /validate debt

	//setup paid
	paid := payment.Paid{
		ID:                 nil,
		PaymentDate:        time.Now(),
		BTCPaid:            SATOSHI_5,
		BTCTransactionDate: time.Now(),
		BTCTransactionID:   0,
		ETHPaid:            0.0,
		ETHTransactionDate: time.Now(),
		ETHTransactionID:   0,
		AddressPaidFrom:    "IAMANADDRESS",
		Username:           email,
	}
	err = state.MakePayment(email, paid)
	if err != nil {
		t.Errorf("Error making payment: %s", err.Error())
	}
	// /setup paid
	//validate paid
	debts, err = paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 1 {
		t.Errorf("Debts should have length 1 is %d", len(debts))
	}
	if err = checkDebt(debts[0], SATOSHI_5, true, SATOSHI_5); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	// /validate paid

	//validate new status
	status, err := paymentdb.GetStatus(email)
	if err != nil {
		t.Errorf("Error getting status: %s", err.Error())
	}
	if err = checkStatus(status, 0, SATOSHI_5); err != nil {
		t.Errorf("Error checking status: %s", err.Error())
	}
	// /validate new status
	state.Close()
}

func TestMakePayment_MultiDebt_NoReferrals(t *testing.T) {
	state := NewStateWithMongoEmpty()
	//setup debt
	email := "test@hodl.zone"
	debt := payment.Debt{
		LoanDate:          time.Now(),
		Charge:            0.0,
		AmountLoaned:      int64(SATOSHI_FLOAT * 10.0),
		LoanRate:          0.5,
		GrossAmountEarned: int64(SATOSHI_FLOAT * 5.0),
		Currency:          "BTC",
		CurrencyToBTC:     int64(SATOSHI_FLOAT * 10.0),
		CurrencyToETH:     int64(SATOSHI_FLOAT * 100.0),
		Exchange:          userdb.PoloniexExchange,
		Username:          email,
		FullPaid:          false,
		PaymentPaidAmount: 0,
	}
	paymentdb, err := payment.NewPaymentDatabaseEmpty("mongodb://localhost:27017", "", "")
	if err != nil {
		t.Errorf("Error connecting to database: %s", err.Error())
	}
	err = paymentdb.InsertNewDebt(debt)
	if err != nil {
		t.Errorf("Error inserting new debt: %s", err.Error())
	}
	debt.LoanDate = time.Now().Add(-1 * time.Hour)
	err = paymentdb.InsertNewDebt(debt)
	if err != nil {
		t.Errorf("Error inserting new debt: %s", err.Error())
	}
	// /setup debt
	//validate debt
	debts, err := paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 2 {
		t.Errorf("Debts should have length 1 is %d", len(debts))
	}
	// /validate debt

	//setup paid
	paid := payment.Paid{
		ID:                 nil,
		PaymentDate:        time.Now(),
		BTCPaid:            int64(SATOSHI_FLOAT * 0.75),
		BTCTransactionDate: time.Now(),
		BTCTransactionID:   0,
		ETHPaid:            0,
		ETHTransactionDate: time.Now(),
		ETHTransactionID:   0,
		AddressPaidFrom:    "IAMANADDRESS",
		Username:           email,
	}
	err = state.MakePayment(email, paid)
	if err != nil {
		t.Errorf("Error making payment: %s", err.Error())
	}
	// /setup paid
	//validate paid
	debts, err = paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 2 {
		t.Errorf("Debts should have length 2 is %d", len(debts))
	}
	if err = checkDebt(debts[1], SATOSHI_5, true, SATOSHI_5); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	if err = checkDebt(debts[0], SATOSHI_5, false, int64(SATOSHI_FLOAT*0.25)); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	// /validate paid

	//validate new status
	status, err := paymentdb.GetStatus(email)
	if err != nil {
		t.Errorf("Error getting status: %s", err.Error())
	}
	if err = checkStatus(status, 0.0, int64(SATOSHI_FLOAT*0.75)); err != nil {
		t.Errorf("Error checking status: %s", err.Error())
	}
	// /validate new status

	//validate paid rest of debt
	err = state.MakePayment(email, paid)
	if err != nil {
		t.Errorf("Error making payment: %s", err.Error())
	}
	debts, err = paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 2 {
		t.Errorf("Debts should have length 2 is %d", len(debts))
	}
	if err = checkDebt(debts[1], SATOSHI_5, true, SATOSHI_5); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	if err = checkDebt(debts[0], SATOSHI_5, true, SATOSHI_5); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	// /validate paid rest of debt

	//validate new status
	status, err = paymentdb.GetStatus(email)
	if err != nil {
		t.Errorf("Error getting status: %s", err.Error())
	}
	if err = checkStatus(status, SATOSHI_5, SATOSHI_INT); err != nil {
		t.Errorf("Error checking status: %s", err.Error())
		fmt.Println(status)
	}
	// /validate new status
	state.Close()
}

func TestMakePayment_OneDebt_WithReferrals(t *testing.T) {
	state := NewStateWithMongoEmpty()
	paymentdb, err := payment.NewPaymentDatabaseEmpty("mongodb://localhost:27017", "", "")
	if err != nil {
		t.Errorf("Error connecting to database: %s", err.Error())
	}
	//generate status
	email := "test@hodl.zone"
	email1 := "test1@hodl.zone"
	email10 := "test1@hodll.zone"
	if err = setUpStatForUser(state, paymentdb, email, "", 0); err != nil {
		t.Errorf("Error setting stat for user: %s", err.Error())
	}
	if err = setUpStatForUser(state, paymentdb, email1, "test", int64(SATOSHI_FLOAT*10.0)); err != nil {
		t.Errorf("Error setting stat for user: %s", err.Error())
	}
	if err = setUpStatForUser(state, paymentdb, email10, "test", int64(SATOSHI_FLOAT*10.0)); err != nil {
		t.Errorf("Error setting stat for user: %s", err.Error())
	}
	// /generate status

	//setup debt
	debt := payment.Debt{
		LoanDate:          time.Now(),
		Charge:            0.0,
		AmountLoaned:      int64(SATOSHI_FLOAT * 10.0),
		LoanRate:          0.5,
		GrossAmountEarned: int64(SATOSHI_FLOAT * 5.0),
		Currency:          "BTC",
		CurrencyToBTC:     int64(SATOSHI_FLOAT * 10.0),
		CurrencyToETH:     int64(SATOSHI_FLOAT * 100.0),
		Exchange:          userdb.PoloniexExchange,
		Username:          email,
		FullPaid:          false,
		PaymentPaidAmount: 0,
	}
	err = paymentdb.InsertNewDebt(debt)
	if err != nil {
		t.Errorf("Error inserting new debt: %s", err.Error())
	}
	// /setup debt
	//validate debt
	debts, err := paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 1 {
		t.Errorf("Debts should have length 1 is %d", len(debts))
	}
	// /validate debt

	//setup paid
	paid := payment.Paid{
		ID:                 nil,
		PaymentDate:        time.Now(),
		BTCPaid:            SATOSHI_5,
		BTCTransactionDate: time.Now(),
		BTCTransactionID:   0,
		ETHPaid:            0,
		ETHTransactionDate: time.Now(),
		ETHTransactionID:   0,
		AddressPaidFrom:    "IAMANADDRESS",
		Username:           email,
	}
	err = state.MakePayment(email, paid)
	if err != nil {
		t.Errorf("Error making payment: %s", err.Error())
	}
	// /setup paid
	//validate paid
	debts, err = paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 1 {
		t.Errorf("Debts should have length 1 is %d", len(debts))
	}
	if err = checkDebt(debts[0], int64(SATOSHI_FLOAT*0.45), true, int64(SATOSHI_FLOAT*0.45)); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	// /validate paid

	//validate new status
	status, err := paymentdb.GetStatus(email)
	if err != nil {
		t.Errorf("Error getting status: %s", err.Error())
	}
	if err = checkStatus(status, int64(SATOSHI_FLOAT*0.05), int64(SATOSHI_FLOAT*0.45)); err != nil {
		t.Errorf("Error checking status: %s", err.Error())
	}
	// /validate new status
	state.Close()
}

func TestMakePayment_MultiDebt_WithReferrals(t *testing.T) {
	state := NewStateWithMongoEmpty()
	paymentdb, err := payment.NewPaymentDatabaseEmpty("mongodb://localhost:27017", "", "")
	if err != nil {
		t.Errorf("Error connecting to database: %s", err.Error())
	}
	//generate status
	email := "test@hodl.zone"
	email1 := "test1@hodl.zone"
	email10 := "test1@hodll.zone"
	if err = setUpStatForUser(state, paymentdb, email, "", 0); err != nil {
		t.Errorf("Error setting stat for user: %s", err.Error())
	}
	if err = setUpStatForUser(state, paymentdb, email1, "test", int64(SATOSHI_FLOAT*10.0)); err != nil {
		t.Errorf("Error setting stat for user: %s", err.Error())
	}
	if err = setUpStatForUser(state, paymentdb, email10, "test", int64(SATOSHI_FLOAT*10.0)); err != nil {
		t.Errorf("Error setting stat for user: %s", err.Error())
	}
	// /generate status

	//setup debt
	debt := payment.Debt{
		LoanDate:          time.Now(),
		Charge:            0.0,
		AmountLoaned:      int64(SATOSHI_FLOAT * 10.0),
		LoanRate:          0.5,
		GrossAmountEarned: int64(SATOSHI_FLOAT * 5.0),
		Currency:          "BTC",
		CurrencyToBTC:     int64(SATOSHI_FLOAT * 10.0),
		CurrencyToETH:     int64(SATOSHI_FLOAT * 100.0),
		Exchange:          userdb.PoloniexExchange,
		Username:          email,
		FullPaid:          false,
		PaymentPaidAmount: 0,
	}
	err = paymentdb.InsertNewDebt(debt)
	if err != nil {
		t.Errorf("Error inserting new debt: %s", err.Error())
	}
	debt.LoanDate = time.Now().Add(-1 * time.Hour)
	err = paymentdb.InsertNewDebt(debt)
	if err != nil {
		t.Errorf("Error inserting new debt: %s", err.Error())
	}
	// /setup debt
	//validate debt
	debts, err := paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 2 {
		t.Errorf("Debts should have length 1 is %d", len(debts))
	}
	// /validate debt

	//setup paid
	paid := payment.Paid{
		ID:                 nil,
		PaymentDate:        time.Now(),
		BTCPaid:            int64(SATOSHI_FLOAT * 0.675),
		BTCTransactionDate: time.Now(),
		BTCTransactionID:   0,
		ETHPaid:            0,
		ETHTransactionDate: time.Now(),
		ETHTransactionID:   0,
		AddressPaidFrom:    "IAMANADDRESS",
		Username:           email,
	}
	err = state.MakePayment(email, paid)
	if err != nil {
		t.Errorf("Error making payment: %s", err.Error())
	}
	// /setup paid
	//validate paid
	debts, err = paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 2 {
		t.Errorf("Debts should have length 2 is %d", len(debts))
	}
	if err = checkDebt(debts[1], int64(SATOSHI_FLOAT*0.45), true, int64(SATOSHI_FLOAT*0.45)); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	if err = checkDebt(debts[0], int64(SATOSHI_FLOAT*0.45), false, int64(SATOSHI_FLOAT*0.225)); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	// /validate paid

	//validate new status
	status, err := paymentdb.GetStatus(email)
	if err != nil {
		t.Errorf("Error getting status: %s", err.Error())
	}
	if err = checkStatus(status, 0.0, int64(SATOSHI_FLOAT*0.675)); err != nil {
		t.Errorf("Error checking status: %s", err.Error())
	}
	// /validate new status

	//validate paid rest of debt
	err = state.MakePayment(email, paid)
	if err != nil {
		t.Errorf("Error making payment: %s", err.Error())
	}
	debts, err = paymentdb.GetAllDebts(email, 0)
	if err != nil {
		t.Errorf("Error getting all debts: %s", err.Error())
	}
	if len(debts) != 2 {
		t.Errorf("Debts should have length 2 is %d", len(debts))
	}
	if err = checkDebt(debts[1], int64(SATOSHI_FLOAT*0.45), true, int64(SATOSHI_FLOAT*0.45)); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	if err = checkDebt(debts[0], int64(SATOSHI_FLOAT*0.45), true, int64(SATOSHI_FLOAT*0.45)); err != nil {
		t.Errorf("Error checking debt: %s", err.Error())
	}
	// /validate paid rest of debt

	//validate new status
	status, err = paymentdb.GetStatus(email)
	if err != nil {
		t.Errorf("Error getting status: %s", err.Error())
	}
	if err = checkStatus(status, int64(SATOSHI_FLOAT*0.45), int64(SATOSHI_FLOAT*0.9)); err != nil {
		t.Errorf("Error checking status: %s", err.Error())
	}
	// /validate new status
	state.Close()
}

func checkDebt(debt payment.Debt, charge int64, fullpaid bool, ppa int64) error {
	if debt.Charge != charge {
		return fmt.Errorf("Charge should be %d is %d", charge, debt.Charge)
	}
	if debt.FullPaid != fullpaid {
		return fmt.Errorf("FullPaid is %t, should be %t", debt.FullPaid, fullpaid)
	}
	if debt.PaymentPaidAmount != ppa {
		return fmt.Errorf("PPA should be %d is %d", ppa, debt.PaymentPaidAmount)
	}
	return nil
}

func checkStatus(status *payment.Status, unspentcred, spentcred int64) error {
	if status.UnspentCredits != unspentcred {
		return fmt.Errorf("UnspentCredits should be %v is %v", unspentcred, status.UnspentCredits)
	}
	if status.SpentCredits != spentcred {
		return fmt.Errorf("SpentCredits should be %v is %v", spentcred, status.SpentCredits)
	}
	return nil
}

func setUpStatForUser(state *State, paymentdb *payment.PaymentDatabase, email, refereeEmail string, paidAmount int64) error {
	status, err := state.GetPaymentStatus(email)
	if err != nil {
		return fmt.Errorf("Error adding user[%s] %s", email, err.Error())
	}
	status.RefereeCode = refereeEmail
	status.SpentCredits = paidAmount / 2
	status.UnspentCredits = paidAmount / 2
	if err = paymentdb.SetStatus(*status); err != nil {
		return fmt.Errorf("Error setting %s status: %s", refereeEmail, err.Error())
	}
	return nil
}
