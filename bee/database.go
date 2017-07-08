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
		b.userStatDB.RecordData(stats, userdb.PoloniexExchange)
	case balancer.BitfinexExchange:
		b.userStatDB.RecordData(stats, userdb.BitfinexExchange)
	default:
		return fmt.Errorf("Exchange not recognized: %d", exchange)
	}
	return nil
}

func (b *Bee) FetchUser(username string) (*userdb.User, error) {
	return b.userDB.FetchUser(username)
}
