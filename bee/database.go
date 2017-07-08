package bee

import (
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
)

var _ = mongo.AUDIT_DB

func (b *Bee) SaveUserStastics(stats *userdb.AllUserStatistic) error {
	return b.userStatDB.RecordData(stats)
}

func (b *Bee) FetchUser(username string) (*userdb.User, error) {
	return b.userDB.FetchUser(username)
}
