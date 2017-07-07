package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time" // or "runtime"

	"github.com/Emyrk/LendingBot/bee"
)

func main() {
	var (
		address = flag.String("a", "dev.hodl.zone:9100", "Address to connect to the balancer")
		dba     = flag.String("dba", "mongo.hodl.zone:27017", "Address for db to connect")
		dbu     = flag.String("dbu", "", "Username for db to connect")
		dbp     = flag.String("dbp", "", "Password for db to connect")
		test    = flag.Bool("test", false, "Testmode")
	)

	flag.Parse()

	be := bee.NewBee(*address, *dba, *dbu, *dbp, *test)
	be.Run()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Recievived ctrl+c. Closing balancer")
		be.Shutdown()
		os.Exit(1)
	}()

	log.Printf("Now running Bee")
	for {
		log.Printf("Using %d users", len(be.Users))
		time.Sleep(10 * time.Minute)
	}
}
