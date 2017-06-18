package main

import (
	"flag"

	. "github.com/Emyrk/LendingBot/src/slave"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		address = flag.String("a", "34.211.81.174:6667", "Master to connect to")
	)

	flag.Parse()
	log.SetLevel(log.InfoLevel)

	s := NewSlave(*address)
	s.Run()
}
