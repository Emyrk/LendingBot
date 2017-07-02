package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time" // or "runtime"

	"github.com/Emyrk/LendingBot/balancer"
)

func main() {
	var (
		port = flag.Int("p", "7000", "Port to listen on for balancer")
	)

	flag.Parse()

	bal := balancer.NewBalancer()
	bal.Run(*port)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Recievived ctrl+c. Closing balancer")
		bal.Close()
		os.Exit(1)
	}()

	log.Printf("Now running Balancer")
	for {
		log.Printf("Using %d bees", bal.ConnectionPool.Slaves.SwarmCount())
		time.Sleep(10 * time.Minute)
	}
}
