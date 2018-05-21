package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	// "time" // or "runtime"

	"github.com/Emyrk/LendingBot/bee"
)

func main() {
	var (
		address = flag.String("a", "bee.hodl.zone:9200", "Address to connect to the balancer")
		dba     = flag.String("dba", "mongodb://mongo1.hodl.zone:4000", "Address for db to connect")
		dbu     = flag.String("dbu", "bee", "Username for db to connect")
		test    = flag.Bool("test", false, "Testmode")
	)

	flag.Parse()

	pass := os.Getenv("MONGO_BEE_PASS")
	if pass == "" && !(*test) {
		panic("No password given for bee")
	}

	be := bee.NewBee(*address, *dba, *dbu, pass, *test)
	be.Run()
	go StartProfiler()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Recievived ctrl+c. Closing bee")
		be.Shutdown()
		os.Exit(1)
	}()

	log.Printf("Now running Bee [%s]", be.ID)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		switch text {
		case "u":
			fmt.Printf("Using %d users\n", len(be.Users))
		case "s":
			fmt.Println(be.Report())
		case "l":
			// fmt.Println(be.LendingBot.FullReport())
		case "t":
			fmt.Println(be.LendingBot.BitfinLender.TickerInfo())
		}
	}
}
