package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/email"
	"github.com/Emyrk/LendingBot/src/core/payment"
	// "github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var statePaymentLog = log.WithFields(log.Fields{
	"package": "core",
	"file":    "statePayment",
})

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

	if status.ReferralCode == refereeCode {
		return &primitives.ApiError{
			fmt.Errorf("RefereeCode for user[%s] is same as users [%s]==[%s]", username, status.RefereeCode, refereeCode),
			fmt.Errorf("You can not set your code as referee code."),
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
	return s.paymentDB.GetDebtsLimitSortIfFound(username, 0, limit, -1)
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
	} else if status == nil || status.ReferralCode == "" {
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
	err := s.paymentDB.SetPaid(paid)
	if err != nil {
		return fmt.Errorf("Error adding pay debt: %s", err.Error())
	}

	if err = s.paymentDB.RecalcAllStatusCredits(username); err != nil {
		return fmt.Errorf("Error recalcing: %s", err.Error())
	}

	if err = s.paymentDB.PayDebts(username); err != nil {
		return fmt.Errorf("Error paying debts: %s", err.Error())
	}

	return s.updateUserLendingHalt(username)
}

//also sets if user should halt
func (s *State) RecalcStatus(username string) error {
	//recalcing
	err := s.paymentDB.RecalcAllStatusCredits(username)
	if err != nil {
		return fmt.Errorf("Error recalcing: %s", err.Error())
	}

	if err = s.paymentDB.PayDebts(username); err != nil {
		return fmt.Errorf("Error paying debts: %s", err.Error())
	}
	// /recalcing
	return s.updateUserLendingHalt(username)
}

func (s *State) updateUserLendingHalt(username string) error {
	llog := statePaymentLog.WithField("method", "updateUserLendingHalt")
	//update user
	status, err := s.paymentDB.GetStatusIfFound(username)
	if err != nil {
		return fmt.Errorf("Error grabbing status: %s", err.Error())
	}

	user, err := s.userDB.FetchUserIfFound(username)
	if err != nil {
		return fmt.Errorf("Error fetching user: %s", err.Error())
	}

	if status.UnspentCredits < 0.0 {
		user.LendingHalted.Reason = fmt.Sprintf("%s %s", payment.Int64SatoshiToString(status.UnspentCredits), " credits owed. Will not lend until credits are paid.")
		user.LendingHalted.Time = time.Now().UTC()
		user.LendingHalted.Halt = true
	} else {
		user.LendingHalted.Halt = false
		user.LendingHalted.EmailThrottleCount = 0
	}

	//if the the lending rate is halt and the last time email sent was greater than 18 hours
	// then send a new email
	if user.LendingHalted.EmailStop == false && user.LendingHalted.Halt == true {
		lastEmailTime := user.LendingHalted.EmailTime.UTC().UnixNano()

		//calculates the min time for the previous throttle
		tempTime := time.Now().UTC().Add(payment.EMAIL_HALT_THROTTLE_TIMES[len(payment.EMAIL_HALT_THROTTLE_TIMES)-1]).UnixNano()
		if int(user.LendingHalted.EmailThrottleCount) < len(payment.EMAIL_HALT_THROTTLE_TIMES) {
			tempTime = time.Now().UTC().Add(payment.EMAIL_HALT_THROTTLE_TIMES[user.LendingHalted.EmailThrottleCount]).UnixNano()
		}

		if lastEmailTime <= tempTime {
			emailRequest := email.NewHTMLRequest(email.SMTP_EMAIL_NO_REPLY, []string{username}, "Payment Needed")
			err = emailRequest.ParseTemplate("paymentneeded.html", nil)
			if err = emailRequest.SendEmail(); err != nil {
				llog.Errorf("Sending email: %s", err.Error())
			} else {
				// if no error update last time email was sent
				user.LendingHalted.EmailTime = time.Now().UTC()
			}

			if int(user.LendingHalted.EmailThrottleCount+1) < len(payment.EMAIL_HALT_THROTTLE_TIMES) {
				user.LendingHalted.EmailThrottleCount++
			}
		}
	}

	return s.userDB.PutUser(user)
}

func (s *State) AddCustomChargeReduction(username string, percentageAmount, reason string) (*payment.Status, *primitives.ApiError) {
	discount, err := strconv.ParseFloat(percentageAmount, 64)
	if err != nil {
		return nil, &primitives.ApiError{
			LogError:  fmt.Errorf("Error parsing percentageAmount: %.8f", percentageAmount),
			UserError: fmt.Errorf("Percentage amount invalid: %.8f", percentageAmount),
		}
	}

	status, err := s.paymentDB.AddReferralReduction(username, payment.ReductionReason{
		Discount: discount,
		Reason:   reason,
		Time:     time.Now().UTC(),
	})
	if err != nil {
		return nil, &primitives.ApiError{
			LogError:  fmt.Errorf("Error adding referral reduc: %s", err.Error()),
			UserError: fmt.Errorf("Error adding referral: %s", err.Error()),
		}
	}

	return status, nil
}

//struct used for showing a users referrals under the 'view more' button on the payment info page
type UserReferral struct {
	Username     string `json:"email"`
	ReachedLimit bool   `json:"reachedlimit" bson:"reachedlimit"`
}

func (s *State) GetReferrals(username string) ([]UserReferral, *primitives.ApiError) {
	referrals, err := s.paymentDB.GetUserReferralsIfFound(username)
	if err != nil {
		return nil, &primitives.ApiError{
			LogError:  fmt.Errorf("Error retrieving referrals: %s", err.Error()),
			UserError: fmt.Errorf("Internal error retrieving referrals. Please contact: support@hodl.zone"),
		}
	}

	userRef := make([]UserReferral, len(referrals), len(referrals))
	for i, ref := range referrals {
		if ref.SpentCredits+ref.UnspentCredits > payment.REDUCTION_CREDIT {
			userRef[i] = UserReferral{
				Username:     ref.Username,
				ReachedLimit: true,
			}
		} else {
			userRef[i] = UserReferral{
				Username:     ref.Username,
				ReachedLimit: false,
			}
		}
	}

	return userRef, nil
}
