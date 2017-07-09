package bee_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	. "github.com/Emyrk/LendingBot/bee"
	// "github.com/Emyrk/LendingBot/src/core/userdb"
	log "github.com/sirupsen/logrus"
)

var _ = time.Now
var _ = os.Readlink
var _ = json.Marshal

func TestLH(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	address := "dev.hodl.zone:9100"
	dba := "mongo.hodl.zone:27017"
	dbu := "bee"
	dbp := os.Getenv("MONGO_BEE_PASS")

	b := NewBee(address, dba, dbu, dbp, false)
	l := NewLendingHistoryKeeper(b)
	a := ""
	s := ""

	n := time.Now()
	top := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	top = top.Add(time.Hour * 24).Add(-1 * time.Second)
	top = top.Add(-24 * time.Hour)

	t.Log(l.FindStart("stevenmasley@gmail.com", top))
	return
	l.SavePoloniexMonth("stevenmasley@gmail.com", a, s)
}
