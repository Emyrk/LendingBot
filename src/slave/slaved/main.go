package main

import (
	"flag"

	. "github.com/Emyrk/LendingBot/src/slave"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		address = flag.String("a", "localhost:1234", "Master to connect to")
	)

	flag.Parse()
	log.SetLevel(log.InfoLevel)

	s := NewSlave(*address)
	s.Run()
}
