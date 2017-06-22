package mongo

import (
	"gopkg.in/mgo.v2"
)

type IMongoSession interface {
	CreateSession() (*mgo.Session, error)
}
