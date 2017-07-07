package bee

import (
	"fmt"
	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = mongo.AUDIT_DB

func (b *Bee) SaveUserStastics(stats *userdb.AllUserStatistic, exchange int) error {
	switch exchange {
	case balancer.PoloniexExchange:
		b.userStatDB.RecordData(stats)
	case balancer.BitfinexExchange:
		//TODO

		// s, c, err = us.mdb.GetCollection(mongo.C_UserStat_POL)
		// if err != nil {
		// 	return fmt.Errorf("Mongo: RecordData: createSession: %s", err)
		// }
		// fallthrough
	default:
		return fmt.Errorf("Exchange not recognized: %d", exchange)
	}
	return nil
}

func (b *Bee) FetchUser(username string) (*userdb.User, error) {
	return b.userDB.FetchUser(username)
}
