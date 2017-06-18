package mongo

import (
	"fmt"
	"gopkg.in/mgo.v2"
)

/*
Book's model connection
*/
var Books *mgo.Collection

type MongoDB struct {
	uri      string
	dbName   string
	sessions map[string]*mgo.Session
}

func (c *MongoDB) GetURI() string {
	return c.uri
}

func (c *MongoDB) GetDBName() string {
	return c.dbName
}

func (c *MongoDB) GetSession(sessionName string) *mgo.Session {
	return c.sessions[sessionName]
}

func CreateMongoDB(uri, dbname string) *MongoDB {
	m := make(map[string]*mgo.Session)
	mongoDB := &MongoDB{
		uri,
		dbname,
		m,
	}
	return mongoDB
}

func (c *MongoDB) CreateSession(sessionName string) (*mgo.Session, error) {
	session, err := mgo.Dial(c.uri)
	if err != nil {
		return nil, err
	}

	c.sessions[sessionName] = new(mgo.Session)
	c.sessions[sessionName] = session

	// See https://godoc.org/labix.org/v2/mgo#Session.SetMode
	session.SetMode(mgo.Monotonic, true)

	return session, nil
}

func (c *MongoDB) CloseSession(name string) error {
	s, ok := c.sessions[name]
	if !ok {
		return fmt.Errorf("Session not found. Unable to close.")
	}
	s.Close()
	delete(c.sessions, name)
	return nil
}
