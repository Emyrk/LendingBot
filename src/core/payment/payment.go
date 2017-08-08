package payment

import (
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

type PaymentDatabase struct {
	db *mongo.MongoDB

	//mux for generating code
	referralMux sync.Mutex
}

func NewPaymentDatabaseMap(uri, dbu, dbp string) (*PaymentDatabase, error) {
	db, err := mongo.CreateTestPaymentDB(uri, dbu, dbp)
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}
	s, c, err := db.GetCollection(mongo.C_Status)
	if err != nil {
		return nil, fmt.Errorf("NewPaymentDatabaseMap: status: createSession: %s", err)
	}
	err = c.Remove(bson.M{})
	if err != nil {
		return nil, err
	}
	s.Close()
	s, c, err = db.GetCollection(mongo.C_Debt)
	if err != nil {
		return nil, fmt.Errorf("NewPaymentDatabaseMap: debt: createSession: %s", err)
	}
	err = c.Remove(bson.M{})
	if err != nil {
		return nil, err
	}
	s.Close()
	s, c, err = db.GetCollection(mongo.C_Paid)
	if err != nil {
		return nil, fmt.Errorf("NewPaymentDatabaseMap: paid: createSession: %s", err)
	}
	err = c.Remove(bson.M{})
	if err != nil {
		return nil, err
	}
	s.Close()

	return &PaymentDatabase{db: db}, err
}

func NewPaymentDatabase(uri, dbu, dbp string) (*PaymentDatabase, error) {
	db, err := mongo.CreatePaymentDB(uri, dbu, dbp)
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}
	return &PaymentDatabase{db: db}, err
}

func NewPaymentDatabaseGiven(db *mongo.MongoDB) *PaymentDatabase {
	return &PaymentDatabase{db: db}
}

func (p *PaymentDatabase) Close() error {
	// if p.db == nil {
	// 	return p.db.Close()
	// }
	return nil
}

type Status struct {
	Username              string    `json:"email" bson:"_id"`
	TotalDebt             float64   `json:"tdebt" bson:"tdebt"`
	UnspentCredits        float64   `json:"unspentcred" bson:"unspentcred"`
	SpentCredits          float64   `json:"spentcred" bson:"spentcred"`
	CustomChargeReduction float64   `json:"customchargereduc" bson:"customchargereduc"`
	RefereeCode           string    `json:"refereecode" bson:"refereecode"` //(Person code who referred you)
	RefereeTime           time.Time `json:"refereetime" bson:"refereetime"` //NOTE time is set to start of time until refereecode is set
	ReferralCode          string    `json:"referralcode" bson:"referralcode"`
}

func (u *Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		TotalDebt             float64   `json:"tdebt"`
		UnspentCredits        float64   `json:"unspentcred"`
		SpentCredits          float64   `json:"spentcred"`
		CustomChargeReduction float64   `json:"customchargereduc"`
		RefereeCode           string    `json:"refereecode"` //(Person code who referred you)
		RefereeTime           time.Time `json:"refereetime"`
		ReferralCode          string    `json:"referralcode"`
	}{
		u.TotalDebt,
		u.UnspentCredits,
		u.SpentCredits,
		u.CustomChargeReduction,
		u.RefereeCode,
		u.RefereeTime,
		u.ReferralCode,
	})
}

func (p *PaymentDatabase) GetUserReferralsIfFound(username string) ([]Status, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		var sr []Status
		return sr, fmt.Errorf("AddUserReferral: createSession: %s", err.Error())
	}
	defer s.Close()
	ref, err := p.getUserReferralsGiven(username, c)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return ref, nil
	} else if err != nil {
		return ref, err
	}
	return ref, nil
}

func (p *PaymentDatabase) GetUserReferrals(username string) ([]Status, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		var sr []Status
		return sr, fmt.Errorf("AddUserReferral: createSession: %s", err.Error())
	}
	defer s.Close()
	return p.getUserReferralsGiven(username, c)
}

func (p *PaymentDatabase) getUserReferralsGiven(username string, c *mgo.Collection) ([]Status, error) {
	var result []Status
	find := bson.M{"_id": username}
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

type Debt struct {
	//ID is set by database and is unique
	ID                    *bson.ObjectId      `json:"_id,omitempty" bson:"_id,omitempty"`
	LoanDate              time.Time           `json:"loandate" bson:"loandate"`
	Charge                float64             `json:"charge" bson:"charge"`
	AmountLoaned          float64             `json:"amountloaned" bson:"amountloaned"`
	LoanRate              float64             `json:"loanrate" bson:"loanrate"`
	GrossAmountEarned     float64             `json:"gae" bson:"gae"`
	Currency              string              `json:"cur" bson:"cur"`
	CurrencyToBTC         float64             `json:"curBTC" bson:"curBTC"`
	CurrencyToETH         float64             `json:"curETH" bson:"curETH"`
	Exchange              userdb.UserExchange `json:"exch" bson:"exch"`
	Username              string              `json:"email" bson:"email"`
	FullPaid              bool                `json:"fullpaid" bson:"fullpaid"`
	PaymentPercentageRate float64             `json:"ppr" bson:"ppr"`
}

func (u *Debt) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		LoanDate              time.Time `json:"loandate"`
		Charge                float64   `json:"charge"`
		AmountLoaned          float64   `json:"amountloaned"`
		LoanRate              float64   `json:"loanrate"`
		GrossAmountEarned     float64   `json:"gae"`
		Currency              string    `json:"cur"`
		CurrencyToBTC         float64   `json:"curBTC"`
		CurrencyToETH         float64   `json:"curETH"`
		Exchange              string    `json:"exch"`
		FullPaid              bool      `json:"fullpaid"`
		PaymentPercentageRate float64   `json:"ppr"`
	}{
		u.LoanDate,
		u.Charge,
		u.AmountLoaned,
		u.LoanRate,
		u.GrossAmountEarned,
		u.Currency,
		u.CurrencyToBTC,
		u.CurrencyToETH,
		u.Exchange.ExchangeToFullName(),
		u.FullPaid,
		u.PaymentPercentageRate,
	})
}

func (p *PaymentDatabase) SetDebt(debt Debt) error {
	return p.SetMultiDebt([]Debt{debt})
}

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

	find := bson.M{"_id": username}
	if paid == 1 {
		find["fullpaid"] = true
	} else if paid == 2 {
		find["fullpaid"] = false
	}
	err = c.Find(find).All(&results)
	if err != nil {
		return nil, fmt.Errorf("GetAllDebts: all: %s", err.Error())
	}
	return results, nil
}

func (p *PaymentDatabase) GetDebtsLimitSortIfFound(username string, paid, limit int) ([]Debt, error) {
	results, err := p.GetDebtsLimitSort(username, paid, limit)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return results, nil
	} else if err != nil {
		return results, err
	}
	return results, nil
}

// LIMIT
//   <= 0 will return all
// PAID
// 0 - Both paid and unpaid
// 1 - paid
// 2 - not paid
func (p *PaymentDatabase) GetDebtsLimitSort(username string, paid, limit int) ([]Debt, error) {
	var results []Debt

	s, c, err := p.db.GetCollection(mongo.C_Debt)
	if err != nil {
		return results, fmt.Errorf("GetDebtsLimitSort: getcol: %s", err)
	}
	defer s.Close()

	find := bson.M{"_id": username}
	if paid == 1 {
		find["fullpaid"] = true
	} else if paid == 2 {
		find["fullpaid"] = false
	}
	if limit <= 0 {
		err = c.Find(find).Sort("-loandate").All(&results)
	} else {
		err = c.Find(find).Sort("-loandate").Limit(limit).All(&results)
	}
	if err != nil {
		return nil, fmt.Errorf("GetDebtsLimitSort: all: %s", err.Error())
	}
	return results, nil
}

type Paid struct {
	ID                 *bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	PaymentDate        time.Time      `json:"paymentdate" bson:"paymentdate"`
	BTCPaid            float64        `json:"btcpaid" bson:"btcpaid"`
	BTCTransactionDate time.Time      `json:"btctrandate" bson:"btctrandate"`
	BTCTransactionID   int64          `json:"btctranid" bson:"btctranid"`
	ETHPaid            float64        `json:"ethpaid" bson:"ethpaid"`
	ETHTransactionDate time.Time      `json:"ethtrandate" bson:"ethtrandate"`
	ETHTransactionID   int64          `json:"ethtranid" bson:"ethtranid"`
	AddressPaidFrom    string         `json:"addr" bson:"addr"`
	Username           string         `json:"email" bson:"email"`
}

func (u *Paid) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PaymentDate        time.Time `json:"paymentdate"`
		BTCPaid            float64   `json:"btcpaid"`
		BTCTransactionDate time.Time `json:"btctrandate"`
		BTCTransactionID   int64     `json:"btctranid"`
		ETHPaid            float64   `json:"ethpaid"`
		ETHTransactionDate time.Time `json:"ethtrandate"`
		ETHTransactionID   int64     `json:"ethtranid"`
		AddressPaidFrom    string    `json:"addr"`
	}{
		u.PaymentDate,
		u.BTCPaid,
		u.BTCTransactionDate,
		u.BTCTransactionID,
		u.ETHPaid,
		u.ETHTransactionDate,
		u.ETHTransactionID,
		u.AddressPaidFrom,
	})
}

func (p *PaymentDatabase) AddPaid(paid Paid) error {
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

	find := bson.M{"_id": username}
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
			TotalDebt:             0,
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

	splitArr := strings.Split(username, "@")
	if len(splitArr) != 2 {
		return nil, fmt.Errorf("Error splitting username has more than one @ sign: %s", splitArr)
	}

	base := splitArr[0]
	if len(base) > 6 {
		base = base[0:6]
	}
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

func (p *PaymentDatabase) PayDebts(username string, paid Paid) error {
	//only grab non-paid debts
	debts, err := p.GetDebtsLimitSort(username, 2, -1)
	if err != nil {
		return fmt.Errorf("Error getting all debts: %s", err.Error())
	}

	//pay off debts one at a time
	btcLeft := paid.BTCPaid
	for i := len(debts) - 1; i >= 0; i++ {
		alreadyPaidBTC := debts[i].PaymentPercentageRate * debts[i].Charge
		if btcLeft >= debts[i].Charge-alreadyPaidBTC {
			//if btcPaid is greater then this one debt
			debts[i].PaymentPercentageRate = 1.0
			debts[i].FullPaid = true
			btcLeft = btcLeft - (debts[i].Charge - alreadyPaidBTC)
		} else {
			//if btcPaid is less than this debt
			debts[i].PaymentPercentageRate = (alreadyPaidBTC + btcLeft) / debts[i].Charge
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

	return p.updateStatusCredits(username, paid.BTCPaid-btcLeft, btcLeft)
}

func (p *PaymentDatabase) updateStatusCredits(username string, usedBTC, leftoverBTC float64) error {
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
//		PaymentPercentageRate
func (p *PaymentDatabase) InsertNewDebt(debt Debt) error {
	//STEVE CALL THIS TO ADD NEW DEBT
	userStatus, err := p.GetStatusIfFound(debt.Username)
	if err != nil {
		return fmt.Errorf("Error finding user referrals: %s", err)
	}

	refs, err := p.GetUserReferralsIfFound(debt.Username)
	if err != nil {
		return fmt.Errorf("Error finding user referrals: %s", err)
	}

	referralReduc := 0.0
	for _, r := range refs {
		referralReduc += 0.005 * ((r.SpentCredits + r.UnspentCredits) / 0.025)
	}

	paidUsReduc := ((userStatus.SpentCredits + userStatus.UnspentCredits) / 0.01) * 0.001
	discount := referralReduc + paidUsReduc
	if discount > 0.035 {
		discount = 0.035
	}
	finalPercentage := 0.10 - userStatus.CustomChargeReduction - discount

	debt.Charge = debt.CurrencyToBTC * finalPercentage
	debt.FullPaid = false
	debt.PaymentPercentageRate = 0.0
	debt.ID = nil

	return p.SetDebt(debt)
}
