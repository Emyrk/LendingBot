package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("============= Human Readable =============")
	log.SetLevel(log.DebugLevel)
	Print()
	fmt.Println()
	fmt.Println("============= JSON Format =============")
	log.SetFormatter(&log.JSONFormatter{})
	Print()
}

func Print() {
	log.Debug("A message with debug level")
	log.Info("A message with info level")
	log.Warn("A message with warn level")
	log.Error("A message with error level")

	log.WithField("package", "Package").Info("The package field shows what golang package made the log")
	log.WithField("method", "Function()").Info("The function field shows what function made the log")
	log.WithField("user", "Username@email.com").Info("The user field indicates the log is for the user")
	log.WithField("subpackage", "Lender").Info("Additonal info about which subpackage. With new infrastructure, this will kinda go away")
	log.WithField("bee", "AAAAAAAAAAAAAAAAAA").Info("Bee ID in hex")
	log.WithField("instancetype", "bee").Info("Types are 'hive', 'bee', and 'revel'. (Revel is webserver)")

	fmt.Println()
	fmt.Println("      ---------------- Example Actual Logs ----------------")
	fmt.Println()

	plog := log.WithFields(log.Fields{"packager": "Package", "method": "CreateLoan()", "user": "username@email.com", "bee": "AAAAAAAAAAAAAAAAAA", "instancetype": "bee"})
	plog.WithFields(log.Fields{"currency": "BTC", "rate": "0.0031", "amount": ".12"}).Info("Loan created")
	plog.WithField("retry", "1").Error("Error getting balances: Connection timed out. Please try again.")
}
