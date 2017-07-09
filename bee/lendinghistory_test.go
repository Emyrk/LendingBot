package bee_test

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/Emyrk/LendingBot/bee"
	// "github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

func TestLH(t *testing.T) {
	return
	log.SetLevel(log.DebugLevel)
	address := "dev.hodl.zone:9100"
	dba := "mongo.hodl.zone:27017"
	dbu := "bee"
	dbp := os.Getenv("MONGO_BEE_PASS")

	b := NewBee(address, dba, dbu, dbp, false)
	l := NewLendingHistoryKeeper(b)
	a := ""
	s := ""
	l.SaveMonth("stevenmasley@gmail.com", a, s)
}
