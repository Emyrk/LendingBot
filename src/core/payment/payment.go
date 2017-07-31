package payment

import (
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PaymentDatabase struct {
	db *mongo.MongoDB
}

func NewPaymentDatabase(uri, dbu, dbp string) (*PaymentDatabase, error) {
	db, err := mongo.CreatePaymentDB(uri, dbu, dbp)
	if err != nil {
		return nil, fmt.Errorf("Error creating payment db: %s\n", err.Error())
	}
	return &PaymentDatabase{db}, err
}

func NewPaymentDatabaseGiven(db *mongo.MongoDB) *PaymentDatabase {
	return &PaymentDatabase{db}
}

func (p *PaymentDatabase) Close() error {
	// if p.db == nil {
	// 	return p.db.Close()
	// }
	return nil
}

type Status struct {
	Username              string           `bson:"_id"`
	TotalDebt             float64          `bson:"tdebt"`
	UnspentCredits        float64          `bson:"unspentcred"`
	SpentCredits          float64          `bson:"spentcred"`
	CustomChargeReduction float64          `bson:"customchargereduc"`
	Referee               string           `bson:"referee"` //(Person who referred you)
	ReferralReductions    []StatusReferral `bson:"referralreducs"`
}

type StatusReferral struct {
	Username      string    `bson:"email"`
	ReductionTime time.Time `bson:"reductime"`
}

func (p *PaymentDatabase) SetUserReferee(username, refereeUsername string) error {
	if username == refereeUsername {
		return fmt.Errorf("Cannot use referee as referral username")
	}

	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return fmt.Errorf("SetUserReferee: createSession: %s", err.Error())
	}
	defer s.Close()

	st, err := p.getStatusGiven(username, c)
	if err != nil {
		return fmt.Errorf("SetUserReferee: getRef: %s", err.Error())
	}

	if st.Referee != "" {
		fmt.Errorf("Referee already set for user[%s]", username)
	}

	//CAN OPTIMIZE LATER
	upsertKey := bson.M{
		"_id": username,
	}
	upsertAction := bson.M{"$set": refereeUsername}
	_, err = c.Upsert(upsertKey, upsertAction)
	if err != nil {
		return fmt.Errorf("SetUserReferee: upsert: %s", err)
	}
	return nil
}

func (p *PaymentDatabase) AddUserReferral(username, referralUsername string) error {
	if username == referralUsername {
		return fmt.Errorf("Cannot use username as referral username")
	}

	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return fmt.Errorf("AddUserReferral: createSession: %s", err.Error())
	}
	defer s.Close()

	sr, err := p.getUserReferralsGiven(username, c)
	if err != nil && err.Error() != mgo.ErrNotFound.Error() {
		return fmt.Errorf("AddUserReferral: getRef: %s", err.Error())
	} else {
		for _, o := range sr {
			if o.Username == username {
				return fmt.Errorf("Error username[%s] already added as referee[%s]", username, referralUsername)
			}
		}
	}
	//CAN OPTIMIZE LATER
	upsertKey := bson.M{
		"_id": username,
	}
	upsertAction := bson.M{
		"$push": bson.M{
			"referralreducs": &StatusReferral{
				Username:      referralUsername,
				ReductionTime: time.Now().UTC(),
			},
		},
	}

	_, err = c.Upsert(upsertKey, upsertAction)
	if err != nil {
		return fmt.Errorf("AddUserReferral: upsert: %s", err)
	}
	return nil
}

func (p *PaymentDatabase) GetUserReferralsIfFound(username string) ([]StatusReferral, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		var sr []StatusReferral
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

func (p *PaymentDatabase) GetUserReferrals(username string) ([]StatusReferral, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		var sr []StatusReferral
		return sr, fmt.Errorf("AddUserReferral: createSession: %s", err.Error())
	}
	defer s.Close()
	return p.getUserReferralsGiven(username, c)
}

func (p *PaymentDatabase) getUserReferralsGiven(username string, c *mgo.Collection) ([]StatusReferral, error) {
	var result struct {
		ReferralReductions []StatusReferral `bson:"referralreducs"`
	}

	find := bson.M{"_id": username}
	sel := bson.M{"_id": 0}
	err := c.Find(find).Select(sel).One(&result)
	return result.ReferralReductions, err
}

func (p *PaymentDatabase) GetStatus(username string) (*Status, error) {
	s, c, err := p.db.GetCollection(mongo.C_Status)
	if err != nil {
		return nil, fmt.Errorf("GetStatus: getcol: %s", err)
	}
	defer s.Close()
	return p.getStatusGiven(username, c)
}

func (p *PaymentDatabase) getStatusGiven(username string, c *mgo.Collection) (*Status, error) {
	var result Status
	err := c.Find(bson.M{"_id": username}).One(&result)
	if err != nil {
		return nil, fmt.Errorf("GetStatus: one: %s", err.Error())
	}
	return &result, nil
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

type Debt struct {
	ID                    *bson.ObjectId      `bson:"_id,omitempty"`
	LoanDate              time.Time           `bson:"loandate"`
	Charge                float64             `bson:"charge"`
	AmountLoaned          float64             `bson:"amountloaned"`
	LoanRate              float64             `bson:"loanrate"`
	GrossAmountEarned     float64             `bson:"gae"`
	Currency              string              `bson:"cur"`
	CurrencyToBTC         float64             `bson:"curBTC"`
	CurrencyToETH         float64             `bson:"curETH"`
	Exchange              userdb.UserExchange `bson:"exch"`
	Username              string              `bson:"email"`
	FullPaid              bool                `bson:"fullpaid"`
	PaymentPercentageRate float64             `bson:"ppr"`
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

type Paid struct {
	ID                 *bson.ObjectId `bson:"_id,omitempty"`
	PaymentDate        time.Time      `bson:"paymentdate"`
	BTCPaid            float64        `bson:"btcpaid"`
	BTCTransactionDate time.Time      `bson:"btctrandate"`
	BTCTransactionID   int64          `bson:"btctranid"`
	ETHPaid            float64        `bson:"ethpaid"`
	ETHTransactionDate time.Time      `bson:"ethtrandate"`
	ETHTransactionID   int64          `bson:"ethtranid"`
	AddressPaidFrom    string         `bson:"addr"`
	Username           string         `bson:"email"`
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
	err = c.Find(find).All(&results)
	if err != nil {
		return nil, fmt.Errorf("GetAllPaid: all: %s", err.Error())
	}
	return results, nil
}
