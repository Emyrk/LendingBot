package payment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var EMAIL_HALT_THROTTLE_TIMES = []time.Duration{ //when sending halt emails this is the time waited each time
	-24 * time.Hour,  //1 day
	-48 * time.Hour,  //2 days
	-168 * time.Hour, //7 days
	-336 * time.Hour, //14 days
	-720 * time.Hour, //30 days
}

const (
	DEFAULT_REDUCTION_ROUND = 3 //when calculating reduction, default place to round float to
)

type PaymentDatabase struct {
	db *mongo.MongoDB

	//mux for generating code
	referralMux sync.Mutex

	//mux for status updating
	paidlock *MapLock
}

func NewPaymentDatabaseEmpty(uri, dbu, dbp string) (*PaymentDatabase, error) {
	db, err := mongo.CreateTestPaymentDB(uri, dbu, dbp)
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}
	s, c, err := db.GetCollection(mongo.C_Status)
	if err != nil {
		return nil, fmt.Errorf("NewPaymentDatabaseMap: status: createSession: %s", err)
	}
	_, err = c.RemoveAll(bson.M{})
	if err != nil && err.Error() != mgo.ErrNotFound.Error() {
		return nil, err
	}
	s.Close()
	s, c, err = db.GetCollection(mongo.C_Debt)
	if err != nil {
		return nil, fmt.Errorf("NewPaymentDatabaseMap: debt: createSession: %s", err)
	}
	_, err = c.RemoveAll(bson.M{})
	if err != nil && err.Error() != mgo.ErrNotFound.Error() {
		return nil, err
	}
	s.Close()
	s, c, err = db.GetCollection(mongo.C_Paid)
	if err != nil {
		return nil, fmt.Errorf("NewPaymentDatabaseMap: paid: createSession: %s", err)
	}
	_, err = c.RemoveAll(bson.M{})
	if err != nil && err.Error() != mgo.ErrNotFound.Error() {
		return nil, err
	}
	s.Close()

	pb := &PaymentDatabase{db: db}
	pb.paidlock = NewMapLock()

	return pb, nil
}

func NewPaymentDatabase(uri, dbu, dbp string) (*PaymentDatabase, error) {
	db, err := mongo.CreatePaymentDB(uri, dbu, dbp)
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}

	pb := &PaymentDatabase{db: db}
	pb.paidlock = NewMapLock()

	return pb, err
}

func NewPaymentDatabaseGiven(db *mongo.MongoDB) *PaymentDatabase {
	pb := &PaymentDatabase{db: db}
	pb.paidlock = NewMapLock()

	return pb
}

func (p *PaymentDatabase) Close() error {
	// if p.db == nil {
	// 	return p.db.Close()
	// }
	return nil
}

type Status struct {
	Username                     string            `json:"email" bson:"_id"`
	UnspentCredits               int64             `json:"unspentcred" bson:"unspentcred"`
	SpentCredits                 int64             `json:"spentcred" bson:"spentcred"`
	CustomChargeReduction        float64           `json:"customchargereduc" bson:"customchargereduc"`
	RefereeCode                  string            `json:"refereecode" bson:"refereecode"` //(Person code who referred you)
	RefereeTime                  time.Time         `json:"refereetime" bson:"refereetime"` //NOTE time is set to start of time until refereecode is set
	ReferralCode                 string            `json:"referralcode" bson:"referralcode"`
	CustomChargeReductionReasons []ReductionReason `json:"customchargereducreasons" bson:"customchargereducreasons"`
}

type ReductionReason struct {
	Discount float64   `json:"discount" bson:"discount"`
	Reason   string    `json:"reason" bson:"reason"`
	Time     time.Time `json:"time" bson:"time"`
}

func (u *Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UnspentCredits               string            `json:"unspentcred"`
		SpentCredits                 string            `json:"spentcred"`
		CustomChargeReduction        float64           `json:"customchargereduc"`
		RefereeCode                  string            `json:"refereecode"` //(Person code who referred you)
		RefereeTime                  time.Time         `json:"refereetime"`
		ReferralCode                 string            `json:"referralcode"`
		CustomChargeReductionReasons []ReductionReason `json:"customchargereducreasons"`
	}{
		Int64SatoshiToString(u.UnspentCredits),
		Int64SatoshiToString(u.SpentCredits),
		u.CustomChargeReduction,
		u.RefereeCode,
		u.RefereeTime,
		u.ReferralCode,
		u.CustomChargeReductionReasons,
	})
}

func (p *PaymentDatabase) GetUserReferralsIfFound(refereeCode string) ([]Status, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		var sr []Status
		return sr, fmt.Errorf("AddUserReferral: createSession: %s", err.Error())
	}
	defer s.Close()
	ref, err := p.getUserReferralsGiven(refereeCode, c)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return ref, nil
	} else if err != nil {
		return ref, err
	}
	return ref, nil
}

func (p *PaymentDatabase) GetUserReferrals(refereeCode string) ([]Status, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		var sr []Status
		return sr, fmt.Errorf("AddUserReferral: createSession: %s", err.Error())
	}
	defer s.Close()
	return p.getUserReferralsGiven(refereeCode, c)
}

func (p *PaymentDatabase) getUserReferralsGiven(refereeCode string, c *mgo.Collection) ([]Status, error) {
	var result []Status
	find := bson.M{"refereecode": refereeCode}
	//CAN OPTIMIZE to use less data
	err := c.Find(find).All(&result)
	return result, err
}

func (p *PaymentDatabase) GetStatusIfFound(username string) (*Status, error) {
	status, err := p.GetStatus(username)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return status, nil
}

func (p *PaymentDatabase) GetStatus(username string) (*Status, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return nil, fmt.Errorf("GetStatus: getcol: %s", err)
	}
	defer s.Close()
	return p.getStatusGiven(username, c)
}

func (p *PaymentDatabase) SetStatus(status Status) error {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return fmt.Errorf("SetStatus: getcol: %s", err)
	}
	defer s.Close()

	_, err = c.UpsertId(status.Username, bson.M{"$set": status})
	if err != nil {
		return fmt.Errorf("SetStatus: upsert: %s", err)
	}
	return nil
}

func (p *PaymentDatabase) ReferralCodeExists(referralCode string) (bool, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return false, err
	}
	defer s.Close()
	_, err = p.getStatusReferralGiven(referralCode, c)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (p *PaymentDatabase) getStatusGiven(username string, c *mgo.Collection) (*Status, error) {
	if username == "" {
		return nil, fmt.Errorf("Username is empty.")
	}

	var result Status
	err := c.Find(bson.M{"_id": username}).One(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *PaymentDatabase) getStatusReferralGiven(referralCode string, c *mgo.Collection) (*Status, error) {
	var result Status
	err := c.Find(bson.M{"referralcode": referralCode}).One(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *PaymentDatabase) RecalcAllStatusCredits(username string) error {
	return p.RecalcMultiAllStatusCredits([]string{username})
}

func (p *PaymentDatabase) RecalcMultiAllStatusCredits(usernames []string) error {
	var (
		debtPaid  int64
		debtTotal int64
		paidTotal int64
	)

	s, c, err := p.db.GetCollection(mongo.C_Debt)
	if err != nil {
		return fmt.Errorf("GetAllDebts: getcol: %s", err)
	}
	defer s.Close()
	for _, username := range usernames {
		//lock
		lock, _ := p.paidlock.GetLocked(username)

		////
		//DEBT PAID
		/////
		o1 := bson.D{{
			"$match", bson.M{
				"$and": []bson.M{
					bson.M{"email": username},
					bson.M{"ppa": bson.M{"$gt": 0}},
				},
			},
		}}
		o2 := bson.D{{
			"$group", bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$ppa"},
			},
		}}
		ops := []bson.D{o1, o2}

		var result bson.M
		err = c.Pipe(ops).One(&result)
		if err != nil && err.Error() != mgo.ErrNotFound.Error() {
			p.paidlock.UnlockPayment(username, lock)
			return fmt.Errorf("Error debtPaid: %s", err.Error())
		}

		if err == mgo.ErrNotFound {
			debtPaid = 0
		} else {
			debtPaid = result["total"].(int64)
		}

		////
		//DEBT TOTAL
		/////
		o1 = bson.D{{
			"$match", bson.M{"email": username},
		}}
		o2 = bson.D{{
			"$group", bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$charge"},
			},
		}}
		ops = []bson.D{o1, o2}

		err = c.Pipe(ops).One(&result)
		if err != nil && err.Error() != mgo.ErrNotFound.Error() {
			p.paidlock.UnlockPayment(username, lock)
			return fmt.Errorf("Error debtTotal: %s", err.Error())
		}

		if err == mgo.ErrNotFound {
			debtTotal = 0
		} else {
			debtTotal = result["total"].(int64)
		}

		////
		//PAID
		/////
		o1 = bson.D{{"$match", bson.M{"email": username}}}
		o2 = bson.D{{
			"$group", bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$btcpaid"},
			},
		}}
		ops = []bson.D{o1, o2}

		err = s.DB(p.db.DbName).C(mongo.C_Paid).Pipe(ops).One(&result)
		if err != nil && err.Error() != mgo.ErrNotFound.Error() {
			p.paidlock.UnlockPayment(username, lock)
			return fmt.Errorf("Error total paid: %s", err.Error())
		}

		if err == mgo.ErrNotFound {
			paidTotal = 0
		} else {
			paidTotal = result["total"].(int64)
		}

		////
		//UPSERT
		/////
		update := bson.M{
			"$set": bson.M{
				"unspentcred": paidTotal - debtTotal,
				"spentcred":   debtPaid,
			},
		}

		_, err = s.DB(p.db.DbName).C(mongo.C_Status).UpsertId(username, update)
		if err != nil {
			p.paidlock.UnlockPayment(username, lock)
			return fmt.Errorf("Error setting status: %s", err.Error())
		}

		//unlock
		p.paidlock.UnlockPayment(username, lock)
	}
	return nil
}

type Debt struct {
	//ID is set by database and is unique
	ID                   *bson.ObjectId      `json:"_id,omitempty" bson:"_id,omitempty"`
	ExchangeID           int                 `json:"exchid" bson:"exchid"`
	LoanOpenDate         time.Time           `json:"loanopendate" bson:"loanopendate"`
	LoanCloseDate        time.Time           `json:"loandate" bson:"loandate"`
	Charge               int64               `json:"charge" bson:"charge"`
	ChargeRate           float64             `json:"chargerate" bson:"chargerate"`
	AmountLoaned         int64               `json:"amountloaned" bson:"amountloaned"`
	LoanRate             float64             `json:"loanrate" bson:"loanrate"`
	GrossAmountEarned    int64               `json:"gae" bson:"gae"`
	GrossBTCAmountEarned int64               `json:"gaebtc" bson:"gaebtc"`
	Currency             string              `json:"cur" bson:"cur"`
	CurrencyToBTC        int64               `json:"curBTC" bson:"curBTC"`
	CurrencyToETH        int64               `json:"curETH" bson:"curETH"`
	Exchange             userdb.UserExchange `json:"exch" bson:"exch"`
	Username             string              `json:"email" bson:"email"`
	FullPaid             bool                `json:"fullpaid" bson:"fullpaid"`
	PaymentPaidAmount    int64               `json:"ppa" bson:"ppa"`
}

func (u *Debt) MarshalJSON() ([]byte, error) {
	formatFloat := func(f float64) string {
		return fmt.Sprintf("%.8f", f)
	}

	return json.Marshal(&struct {
		LoanCloseDate        string `json:"loandate"`
		LoanOpenDate         string `json:"loanopendate"`
		Charge               string `json:"charge"`
		ChargeRate           string `json:"chargerate" bson:"chargerate"`
		AmountLoaned         string `json:"amountloaned"`
		LoanRate             string `json:"loanrate"`
		GrossAmountEarned    string `json:"gae"`
		GrossBTCAmountEarned string `json:"gaebtc"`
		Currency             string `json:"cur"`
		CurrencyToBTC        int64  `json:"curBTC"`
		CurrencyToETH        int64  `json:"curETH"`
		Exchange             string `json:"exch"`
		FullPaid             bool   `json:"fullpaid"`
		PaymentPaidAmount    string `json:"ppa"`
	}{
		u.LoanCloseDate.Format("2006-01-02 15:04:05"),
		u.LoanOpenDate.Format("2006-01-02 15:04:05"),
		formatFloat(float64(u.Charge) / 1e8),
		fmt.Sprintf("%.4f%%", u.ChargeRate*100),
		formatFloat(float64(u.AmountLoaned) / 1e8),
		fmt.Sprintf("%.4f%%", u.LoanRate*100),
		formatFloat(float64(u.GrossAmountEarned) / 1e8),
		formatFloat(float64(u.GrossBTCAmountEarned) / 1e8),
		u.Currency,
		u.CurrencyToBTC,
		u.CurrencyToETH,
		u.Exchange.ExchangeToFullName(),
		u.FullPaid,
		formatFloat(float64(u.PaymentPaidAmount) / 1e8),
	})
}

//DO NOT CALL THIS FOR NEW DEBTS!!! USE InsertNewDebt
func (p *PaymentDatabase) SetDebt(debt Debt) error {
	return p.SetMultiDebt([]Debt{debt})
}

//DO NOT CALL THIS FOR NEW DEBTS!!! USE InsertNewDebt
func (p *PaymentDatabase) SetMultiDebt(debt []Debt) error {
	s, c, err := p.db.GetCollection(mongo.C_Debt)
	if err != nil {
		return fmt.Errorf("SetMultiDebt: getcol: %s", err)
	}
	defer s.Close()

	bulk := c.Bulk()
	for _, o := range debt {
		if o.ID == nil {
			//if nil id assume that this is new record so insert
			bulk.Insert(o)
		} else {
			//upsert to prevent update vs insert error for dups
			bulk.Upsert(
				bson.M{"_id": o.ID},
				bson.M{"$set": o},
			)
		}
	}

	_, err = bulk.Run()
	if err != nil {
		return fmt.Errorf("SetMultiDebt: run: %s", err)
	}
	return nil
}

// PAID
// 0 - Both paid and unpaid
// 1 - paid
// 2 - not paid
func (p *PaymentDatabase) GetAllDebts(username string, paid int) ([]Debt, error) {
	var results []Debt

	s, c, err := p.db.GetCollection(mongo.C_Debt)
	if err != nil {
		return results, fmt.Errorf("GetAllDebts: getcol: %s", err)
	}
	defer s.Close()

	find := bson.M{"email": username}
	if paid == 1 {
		find["fullpaid"] = true
	} else if paid == 2 {
		find["fullpaid"] = false
	}
	err = c.Find(find).All(&results)
	if err != nil {
		return results, fmt.Errorf("GetAllDebts: all: %s", err.Error())
	}
	return results, nil
}

//refer to GetDebtsLimitSort for args info
func (p *PaymentDatabase) GetDebtsLimitSortIfFound(username string, paid, limit, sort int) ([]Debt, error) {
	results, err := p.GetDebtsLimitSort(username, paid, limit, sort)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return results, nil
	} else if err != nil {
		return results, err
	}
	return results, nil
}

// PAID
// 0 - Both paid and unpaid
// 1 - paid
// 2 - not paid
// LIMIT
//   <= 0 will return all
// SORT
//   <= 0 will return "-loandate" (Descending order, largest to smallest)
//   > 0 will return "loandate" (Ascending order, smallest to largest)
func (p *PaymentDatabase) GetDebtsLimitSort(username string, paid, limit, sort int) ([]Debt, error) {
	var results []Debt

	s, c, err := p.db.GetCollection(mongo.C_Debt)
	if err != nil {
		return results, fmt.Errorf("GetDebtsLimitSort: getcol: %s", err)
	}
	defer s.Close()

	find := bson.M{"email": username}
	if paid == 1 {
		find["fullpaid"] = true
	} else if paid == 2 {
		find["fullpaid"] = false
	}
	sortOrder := "-loandate"
	if sort > 0 {
		sortOrder = "loandate"
	}
	if limit <= 0 {
		err = c.Find(find).Sort(sortOrder).All(&results)
	} else {
		err = c.Find(find).Sort(sortOrder).Limit(limit).All(&results)
	}
	if err != nil {
		return nil, fmt.Errorf("GetDebtsLimitSort: all: %s", err.Error())
	}
	return results, nil
}

type Paid struct {
	ID *bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	// Email of account to be credited
	Username string `json:"email" bson:"email"`
	// Email given by user (Where receipt is sent)
	ContactUsername string `json:"contactemail" bson:"contactemail"`

	// Payment Data -- Data about coinbase transaction
	PaymentDate      time.Time `json:"paymentdate" bson:"paymentdate"`
	PaymentCreatedAt time.Time `json:"paymentcreatedat" bson:"paymentcreatedat"`
	PaymentExpiresAt time.Time `json:"paymentexpiresat" bson:"paymentexpiresat"`

	// Transaction Data -- Data about bitcoin transaction
	BTCPaid            int64     `json:"btcpaid" bson:"btcpaid"`
	BTCTransactionID   string    `json:"btctranid" bson:"btctranid"`
	BTCTransactionDate time.Time `json:"btctrandate" bson:"btctrandate"`

	// CoinbaseData
	//		Notification -- When we hear about the payment
	CoinbaseNotificationID string    `json:"noteid" bson:"noteid"`
	CoinbaseUserID         string    `json:"userid" bson:"userid"`
	CoinbaseAccountID      string    `json:"accountid" bson:"accountid"`
	NotificationCreatedAt  time.Time `json:"notificationcreatetime" bson:"notificationcreatetime"`
	NotificationDelivedAt  time.Time `json:"notificationdelivertime" bson:"notificationdelivertime"`
	DeliveryAttempts       int       `json:"deliveryattempts" bson:"deliveryattempts"`

	//		Payment
	CoinbasePaymentID string `json:"paymentid" bson:"paymentid"`
	ReceiptUrl        string `json:"receipt" bson:"receipt"`
	Code              string `json:"code" bson:"code"`
	BTCAddress        string `json:"btcaddr" bson:"btcaddr"`
	RefundAddress     string `json:"refundaddr" bson:"refundaddr"`

	RawData json.RawMessage `json:"rawdata" bson:"rawdata"`
}

func (u *Paid) String() string {
	d, _ := json.Marshal(u)
	return string(d)
}

func (u *Paid) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Username              string `json:"email" bson:"email"`
		ContactUsername       string `json:"contactemail" bson:"contactemail"`
		PaymentDate           string `json:"paymentdate" bson:"paymentdate"`
		BTCPaid               string `json:"btcpaid" bson:"btcpaid"`
		BTCTransactionID      string `json:"btctranid" bson:"btctranid"`
		BTCTransactionDate    string `json:"btctrandate" bson:"btctrandate"`
		NotificationDelivedAt string `json:"notificationdelivertime" bson:"notificationdelivertime"`
		ReceiptUrl            string `json:"receipt" bson:"receipt"`
		Code                  string `json:"code" bson:"code"`
	}{
		u.Username,
		u.ContactUsername,
		u.PaymentDate.Format("2006-01-02 15:04:05"),
		Int64SatoshiToString(u.BTCPaid),
		u.BTCTransactionID,
		u.BTCTransactionDate.Format("2006-01-02 15:04:05"),
		u.NotificationDelivedAt.Format("2006-01-02 15:04:05"),
		u.ReceiptUrl,
		u.Code,
	})
}

func (p *PaymentDatabase) SetPaid(paid Paid) error {
	return p.SetMultiPaid([]Paid{paid})
}

func (p *PaymentDatabase) SetMultiPaid(paid []Paid) error {
	s, c, err := p.db.GetCollection(mongo.C_Paid)
	if err != nil {
		return fmt.Errorf("SetMultiPaid: getcol: %s", err)
	}
	defer s.Close()

	bulk := c.Bulk()
	for _, o := range paid {
		if o.ID == nil {
			//if nil id assume that this is new record so insert
			bulk.Insert(o)
		} else {
			//upsert to prevent update vs insert error for dups
			bulk.Upsert(
				bson.M{"_id": o.ID},
				bson.M{"$set": o},
			)
		}
	}

	_, err = bulk.Run()
	if err != nil {
		return fmt.Errorf("SetMultiPaid: run: %s", err)
	}
	return nil
}

// DATE AFTER
// dateAfter
//  - if nil then will return all dates
//  - else will get all paid after date given
func (p *PaymentDatabase) GetAllPaid(username string, dateAfter *time.Time) ([]Paid, error) {
	var results []Paid

	s, c, err := p.db.GetCollection(mongo.C_Paid)
	if err != nil {
		return results, fmt.Errorf("GetAllPaid: getcol: %s", err)
	}
	defer s.Close()

	find := bson.M{"email": username}
	if dateAfter != nil {
		find["paymentdate"] = bson.M{"$gt": dateAfter}
	}
	err = c.Find(find).Sort("-paymentdate").All(&results)
	if err != nil {
		return nil, fmt.Errorf("GetAllPaid: all: %s", err.Error())
	}
	return results, nil
}

func (p *PaymentDatabase) GenerateReferralCode(username string) (*Status, error) {
	//must lock to avoid conflicts
	p.referralMux.Lock()
	defer p.referralMux.Unlock()

	st, err := p.GetStatus(username)
	if err != nil && err.Error() != mgo.ErrNotFound.Error() {
		return nil, err
	}
	if st == nil {
		st = &Status{
			Username:              username,
			UnspentCredits:        0,
			SpentCredits:          0,
			CustomChargeReduction: 0,
			RefereeCode:           "",
			RefereeTime:           time.Unix(0, 0), //Sets unix time to 1970 init time until refereecode set
			ReferralCode:          "",
		}
	}
	if st.ReferralCode != "" {
		return nil, fmt.Errorf("Referral code already set")
	}

	baseBytes := sha256.Sum256([]byte(username))
	base := hex.EncodeToString(baseBytes[:])
	base = "REF" + strings.ToUpper(base[:10])

	st.ReferralCode = base
	i := 0
	for {
		b, err := p.ReferralCodeExists(st.ReferralCode)
		if err != nil {
			return nil, fmt.Errorf("Error checking if code exists: %s", err.Error())
		}
		if b == false {
			break
		}
		st.ReferralCode = fmt.Sprintf("%s%d", base, i)
		i++
	}
	return st, p.SetStatus(*st)
}

func (p *PaymentDatabase) PayDebts(username string) error {
	status, err := p.GetStatusIfFound(username)
	if err != nil {
		return fmt.Errorf("Error grabbing user stats: %s", err.Error())
	}

	//if unspent is negative, then they owe money
	btcLeft := status.UnspentCredits
	if btcLeft <= 0 {
		return nil
	}

	//only grab non-paid debts
	debts, err := p.GetDebtsLimitSortIfFound(username, 2, -1, -1)
	if err != nil {
		return fmt.Errorf("Error getting all debts: %s", err.Error())
	}
	//pay off debts one at a time
	for i := len(debts) - 1; i >= 0; i-- {
		if btcLeft >= debts[i].Charge-debts[i].PaymentPaidAmount {
			//if btcPaid is greater then this one debt
			btcLeft = btcLeft - (debts[i].Charge - debts[i].PaymentPaidAmount)
			debts[i].PaymentPaidAmount = debts[i].Charge
			debts[i].FullPaid = true
		} else {
			//if btcPaid is less than this debt
			debts[i].PaymentPaidAmount += btcLeft
			debts[i].FullPaid = false
			btcLeft = 0.0
			break
		}
	}
	//save all debts back
	err = p.SetMultiDebt(debts)
	if err != nil {
		return fmt.Errorf("Error setting debts: %s", err.Error())
	}

	return p.updateStatusCredits(username, status.UnspentCredits-btcLeft, btcLeft)
}

func (p *PaymentDatabase) updateStatusCredits(username string, usedBTC, leftoverBTC int64) error {
	lock, _ := p.paidlock.GetLocked(username)
	defer p.paidlock.UnlockPayment(username, lock)

	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return fmt.Errorf("updateStatusCredits: getcol: %s", err)
	}
	defer s.Close()

	update := bson.M{
		"$inc": bson.M{
			"spentcred": usedBTC,
		},
		"$set": bson.M{
			"unspentcred": leftoverBTC,
		},
	}
	return c.UpdateId(username, update)
}

func (p *PaymentDatabase) AddReferralReduction(username string, reducReason ReductionReason) (*Status, error) {
	lock, _ := p.paidlock.GetLocked(username)
	defer p.paidlock.UnlockPayment(username, lock)

	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return nil, fmt.Errorf("AddReferralReduction: getcol: %s", err)
	}
	defer s.Close()

	upsertKey := bson.M{
		"_id": username,
	}
	upsertAction := bson.M{
		"$inc": bson.M{
			"customchargereduc": reducReason.Discount,
		},
		"$push": bson.M{
			"customchargereducreasons": bson.M{
				"$each": []ReductionReason{reducReason},
				"$sort": bson.M{"time": -1},
			},
		},
	}

	_, err = c.Upsert(upsertKey, upsertAction)
	if err != nil {
		return nil, fmt.Errorf("AddReferralReduction: upsert: %s", err)
	}

	var result Status
	err = c.FindId(username).One(&result)
	if err != nil {
		return nil, fmt.Errorf("Error setting status: %s", err.Error())
	}
	return &result, nil
}

const (
	SATOSHI_FLOAT      float64 = float64(100000000)
	SATOSHI_INT        int64   = int64(100000000)
	REDUCTION_CREDIT   int64   = int64(SATOSHI_FLOAT * 0.025)
	REDUCTION_REFERRAL int64   = int64(SATOSHI_FLOAT * 0.005)
	REDUCTION_PAID_01  int64   = int64(SATOSHI_FLOAT * 0.01)
	REDUCTION_PAID_001 int64   = int64(SATOSHI_FLOAT * 0.001)
	MAX_DISCOUNT       int64   = 3500000 //int64(SATOSHI_FLOAT * 0.035) DIDNT LIKE MATH?
	STARTING_CHARGE    int64   = int64(SATOSHI_FLOAT * 0.10)
)

//pass in debt that has the following set fields:
//		LoanDate
//		AmountLoaned
//		LoanRate
//		GrossAmountEarned
//		Currency
//		CurrencyToBTC
//		CurrencyToETH
//		Exchange
//		Username
// Method will set:
//		Charge
//		FullPaid
//		PaymentPaidAmount
func (p *PaymentDatabase) InsertNewDebt(debt Debt) error {
	//STEVE CALL THIS TO ADD NEW DEBT
	userStatus, err := p.GetStatusIfFound(debt.Username)
	if err != nil {
		return fmt.Errorf("Error finding user referrals: %s", err)
	} else if userStatus == nil {
		userStatus, err = p.GenerateReferralCode(debt.Username)
		if err != nil {
			return fmt.Errorf("Failed to create status and new referral code for user[%s]: %s", debt.Username, err.Error())
		}
	}

	refs, err := p.GetUserReferralsIfFound(userStatus.ReferralCode)
	if err != nil {
		return fmt.Errorf("Error finding user referrals: %s", err)
	}

	referralReduc := float64(0.0)
	for _, r := range refs {
		if r.SpentCredits+r.UnspentCredits > REDUCTION_CREDIT {
			referralReduc += 0.005
		}
	}

	paidUsReduc := float64(float64((userStatus.SpentCredits+userStatus.UnspentCredits)/REDUCTION_PAID_01) * 0.001)
	discount := float64(referralReduc + paidUsReduc)
	discount, err = RoundFloat(discount, DEFAULT_REDUCTION_ROUND)
	if err != nil {
		return err
	}
	if discount > 0.035 {
		discount = 0.035
	}
	final := float64(0.10 - userStatus.CustomChargeReduction - discount)

	if final < 0.0 {
		final = 0.0
	}

	debt.ChargeRate = final
	debt.Charge = int64(float64(debt.GrossBTCAmountEarned) * final)
	debt.FullPaid = false
	debt.PaymentPaidAmount = 0
	debt.ID = nil

	return p.SetDebt(debt)
}
