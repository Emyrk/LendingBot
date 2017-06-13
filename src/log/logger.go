package log

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var UsingFile bool = false
var LogFile *os.File
var LogPath string

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	logPath := os.Getenv("LOG_PATH")
	if len(logPath) > 0 {
		// Rename any existing file
		ext := fmt.Sprintf("-%d.back", time.Now().Unix())
		os.Rename(logPath, logPath+ext)

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR, 0666)
		if err == nil {
			log.SetOutput(f)
			UsingFile = true
			LogFile = f
			LogPath = logPath
		}
	}
	log.Info("Logger has initiated")
	log.Info("Logs will now be appended")
}

func Close() {
	LogFile.Close()
}

func ReadLogs() (string, error) {
	if !UsingFile {
		return "", fmt.Errorf("Cannot read logs unless its going to a file")
	}

	return ReadLogFile(LogPath)
}

func ReadLogFile(path string) (string, error) {
	rf, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not read err: %s", err.Error())
	}

	fi, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Could not stat file err: %s %s", path, err.Error())
	}

	max := int64(120000)
	var buf []byte
	if fi.Size() < max {
		max = fi.Size()
	}

	_, err = rf.Seek(fi.Size()-max, 0)
	if err != nil {
		return "", fmt.Errorf("Could not seek file err: %s", err.Error())
	}

	scanner := bufio.NewScanner(rf)
	for scanner.Scan() {
		buf = append([]byte(scanner.Text()+"\n"), buf...)
		// buf.WriteString(scanner.Text() + "\n")
	}

	// logs, err := ioutil.ReadAll(rf)
	// if err != nil {
	// 	return "", err
	// }

	return string(buf), nil
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
