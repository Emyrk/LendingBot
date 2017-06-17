package mongo_test

import (
	"testing"

	. "github.com/Emyrk/LendingBot/src/core/database/mongo"
)

var db *MongoDB
var sessionName = "apples"

func Test_mongo_database_create(t *testing.T) {
	uri := "mongodb://localhost:27017"
	name := "LendingBot"
	db = CreateMongoDB(uri, name)
}

func Test_mongo_create_session(t *testing.T) {
	_, err := db.CreateSession(sessionName)
	if err != nil {
		t.Errorf("Error creating session: %s\n", err.Error())
		t.FailNow()
	}
}

func Test_mongo_close_session(t *testing.T) {
	err := db.CloseSession(sessionName)
	if err != nil {
		t.Errorf("Error closing session: %s\n", err.Error())
	}

	if db.GetSession(sessionName) != nil {
		t.Errorf("Error session should be removed")
	}
}
