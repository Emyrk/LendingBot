package core

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/payment"
)

func (s *State) SetUserReferee(username, refereeCode string) *primitives.ApiError {
	//calls get payment status to set referral code automatically if status does not exist
	exists, err := s.paymentDB.ReferralCodeExists(refereeCode)
	if err != nil {
		errMes := fmt.Errorf("Error checking referral code exists: %s", err.Error())
		return primitives.NewAPIErrorInternalError(errMes)
	}
	if !exists {
		return &primitives.ApiError{
			fmt.Errorf("RefereeCode [%s] does not exist", refereeCode),
			fmt.Errorf("The referee code entered does not exist."),
		}
	}

	status, err := s.GetPaymentStatus(username)
	if err != nil {
		errMes := fmt.Errorf("Error getting payment status: %s", err.Error())
		return primitives.NewAPIErrorInternalError(errMes)
	}

	if status.RefereeCode != "" {
		return &primitives.ApiError{
			fmt.Errorf("RefereeCode for user[%s] already set", username),
			fmt.Errorf("Your referee code has already been set."),
		}
	}

	status.RefereeCode = refereeCode

	err = s.paymentDB.SetStatus(*status)
	if err != nil {
		errMes := fmt.Errorf("Error setting status: %s", err.Error())
		return primitives.NewAPIErrorInternalError(errMes)
	}
	return nil
}

func (s *State) GetUserReferrals(username string) ([]payment.Status, error) {
	return s.paymentDB.GetUserReferrals(username)
}

func (s *State) GetPaymentDebtHistory(username string, limit int) ([]payment.Debt, error) {
	return s.paymentDB.GetDebtsLimitSortIfFound(username, 2, limit)
}

func (s *State) GetPaymentPaidHistory(username, dateAfterStr string) ([]payment.Paid, error) {
	var p []payment.Paid
	if dateAfterStr == "" {
		//get all payments
		return s.paymentDB.GetAllPaid(username, nil)
	}
	//parse time
	layout := "2017-07-31T18:35:34.970Z"
	dateAfter, err := time.Parse(layout, dateAfterStr)
	if err != nil {
		return p, fmt.Errorf("Failed to parse dateAfter: %s", err.Error())
	}
	return s.paymentDB.GetAllPaid(username, &dateAfter)
}

// Gets payment status, if none is found than will create new status with predefined
// referral code and username.
func (s *State) GetPaymentStatus(username string) (*payment.Status, error) {
	status, err := s.paymentDB.GetStatusIfFound(username)
	if err != nil {
		return nil, fmt.Errorf("Failed to get status: %s", err.Error())
	} else if status == nil {
		status, err = s.paymentDB.GenerateReferralCode(username)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate referral code: %s", err.Error())
		}
	}
	return status, nil
}

//returns true if referee code has been set
func (s *State) HasSetReferee(username string) bool {
	status, err := s.paymentDB.GetStatusIfFound(username)
	if err != nil || status == nil {
		return false
	}
	return status.RefereeCode != ""
}

func (s *State) MakePayment(username string, paid payment.Paid) error {
	err := s.paymentDB.AddPaid(paid)
	if err != nil {
		return fmt.Errorf("Error adding pay debt: %s", err.Error())
	}
	err = s.paymentDB.PayDebts(username, paid)
	if err != nil {
		return fmt.Errorf("Error paying debts: %s", err.Error())
	}
	return nil
}
