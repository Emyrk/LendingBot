package log_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/Emyrk/LendingBot/src/log"
	log "github.com/sirupsen/logrus"
)

var _ = ExportLogs

func TestLog(t *testing.T) {
	return
	f, err := os.OpenFile("unittest.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	log.SetOutput(f)
	UsingFile = true
	LogFile = f
	log.Println("ASDASDSD")

	nf, err := os.OpenFile("unittest.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	str2, err := ioutil.ReadAll(nf)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("LOG:", str2)

	str, err := ReadLogs()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("LOG:", str)
	os.Remove("unittest.txt")
}

func TestContext(t *testing.T) {
	log.SetOutput(os.Stdout)
	var contextLogger = log.WithFields(log.Fields{
		"package": "Lender",
	})

	contextLogger.WithField("method", "Dog").Println("ASDASD")
}

func TestRead(t *testing.T) {
	fmt.Println(ReadLogFile("/Users/stevenmasley/go/src/github.com/Emyrk/LendingBot/app/views/index/belowhero.html"))
}
