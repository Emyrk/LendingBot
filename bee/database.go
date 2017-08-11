package bee

import (
	"time"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/payment"
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
	}
	return b.userDB.FetchUserWithSelector(username, selector)
}

func (b *Bee) InsertNewDebt(debt payment.Debt) error {
	return b.paymentDB.InsertNewDebt(debt)
}
