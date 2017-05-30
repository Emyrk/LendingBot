package log

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

var UsingFile bool = false
var LogFile *os.File
var LogPath string

func init() {
	log.SetOutput(os.Stdout)
	logPath := os.Getenv("LOG_PATH")
	if len(logPath) > 0 {
		logPath = logPath
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR, 0666)
		if err == nil {
			log.SetOutput(f)
			UsingFile = true
			LogFile = f
			LogPath = logPath
		}
	}
}

func Close() {
	LogFile.Close()
}

func ReadLogs() (string, error) {
	if !UsingFile {
		return "", fmt.Errorf("Cannot read logs unless its going to a file")
	}

	logs, err := ioutil.ReadAll(LogFile)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

func ExportLogs() (string, error) {
	if !UsingFile {
		return "", fmt.Errorf("Cannot export logs unless its going to a file")
	}

	buf := new(bytes.Buffer)
	log.SetOutput(buf)

	logs, err := ioutil.ReadAll(LogFile)
	if err != nil {
		return "", err
	}

	// Clear logs
	LogFile.Close()
	f, err := os.OpenFile(LogPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return "", err
	}
	LogFile = f

	// Write the logs that were collected in buffer
	f.Write(buf.Bytes())

	// Set output
	log.SetOutput(f)

	return string(logs), nil
}
