package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/lender"
	"github.com/Emyrk/LendingBot/src/lender/otherBots/poloBot"
	"github.com/Emyrk/LendingBot/src/queuer"

	// For Prometheus
	"github.com/prometheus/client_golang/prometheus"
	"net/http"

	// Init logger
	_ "github.com/Emyrk/LendingBot/src/log"

	"github.com/revel/revel"
)

func Launch() {
	// Prometheus
	lender.RegisterPrometheus()
	queuer.RegisterPrometheus()

	state = core.NewState()
	lenderBot := lender.NewLender(state)
	queuerBot := queuer.NewQueuer(state, lenderBot)

	poloBotChannel := make(chan *poloBot.PoloBotParams)
	_, err := poloBot.NewPoloBot(poloBotChannel)
	if err != nil {
		fmt.Printf("ERRRROROROASDOFOASDOF", err)
	}

	err = state.VerifyState()
	if err != nil {
		panic(err)
	}

	if revel.DevMode {
		return
	}
	// Start go lending
	go lenderBot.Start()
	go queuerBot.Start()
	go launchPrometheus(9911)
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
