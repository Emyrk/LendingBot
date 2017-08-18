package bee

import (
	"time"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2/bson"
)

var _ = mongo.AUDIT_DB

func (b *Bee) SaveUserStastics(stats *userdb.AllUserStatistic) error {
	return b.userStatDB.RecordData(stats)
}

func (b *Bee) FetchUser(username string) (*userdb.User, error) {
	selector := bson.M{
		"level":                1,
		"lendingstrategy":      1,
		"poloniexminiumlend":   1,
		"poloniexenabled":      1,
		"poloniexkeys":         1,
		"bitfinexminiumumlend": 1,
		"bitfinexenabled":      1,
		"bitfinexkeys":         1,
		"poloniexenabledtime":  1,
	}
	return b.userDB.FetchUserWithSelector(username, selector)
}

// --NOTE: Uses Loan Date to calculate time--
//pass in duration of time since oldest debt
//used for telling if should stop lending
func (b *Bee) IsDebtOverTimeLimit(username string, dur time.Duration) (bool, error) {
	debts, err := b.paymentDB.GetDebtsLimitSortIfFound(username, 2, 1, 1)
	if err != nil {
		return false, err
	}
	if len(debts) == 0 {
		return false, nil
	}
	if debts[0].LoanDate.UTC().Add(dur).UnixNano() > time.Now().UnixNano() {
		return true, nil
	}
	return false, nil
}
