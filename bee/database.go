package bee

import (
	"fmt"
	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var _ = mongo.AUDIT_DB

func (b *Bee) SaveUserStastics(stats *userdb.AllUserStatistic, exchange int) error {
	var (
		s   *mgo.Session
		c   *mgo.Collection
		err error
	)
	switch exchange {
	case balancer.PoloniexExchange:
		// s, c, err = us.mdb.GetCollection(mongo.C_UserStat_POL)
		// if err != nil {
		// 	return err
		// }
	case balancer.BitfinexExchange:
		//TODO

		// s, c, err = us.mdb.GetCollection(mongo.C_UserStat_POL)
		// if err != nil {
		// 	return fmt.Errorf("Mongo: RecordData: createSession: %s", err)
		// }
		fallthrough
	default:
		return fmt.Errorf("Exchange not recognized: %d", exchange)
	}
	defer s.Close()

	upsertKey := bson.M{
		"$and": []bson.M{
			bson.M{"email": stats.Username},
			bson.M{"time": stats.Time},
		},
	}
	upsertAction := bson.M{"$set": stats}

	_, err = c.Upsert(upsertKey, upsertAction)
	if err != nil {
		return fmt.Errorf("Failed to upsert: %s", err)
	}
	return nil
}

func (b *Bee) FetchUser(username string) (*userdb.User, error) {
	s, c, err := ud.mdb.GetCollection(mongo.C_USER)
	if err != nil {
		return nil, fmt.Errorf("PutUser: getCol: %s", err.Error())
	}
	defer s.Close()

	var result User
	err = c.FindId(username).One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PutUser: find: %s", err.Error())
	}

	result.PoloniexKeys.SetEmptyIfBlank()
	return &result, nil
}
