package bee

import (
	"fmt"
	"github.com/Emyrk/LendingBot/balancer"
	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"github.com/Emyrk/LendingBot/src/core/userdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (b *Bee) SaveUserStastics(stats *userdb.AllUserStatistic, exchange int) error {
	var (
		s   *mgo.Session
		c   *mgo.Collection
		err error
	)
	switch exchange {
	case balancer.PoloniexExchange:
		s, c, err = us.mdb.GetCollection(mongo.C_UserStat_POL)
		if err != nil {
			return err
		}
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
