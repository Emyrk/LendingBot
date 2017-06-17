package controllers

import (
	"fmt"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/lender"
	"github.com/Emyrk/LendingBot/src/queuer"

	// For Prometheus
	"github.com/prometheus/client_golang/prometheus"
	"net/http"

	// Init logger
	_ "github.com/Emyrk/LendingBot/src/log"

	"github.com/revel/revel"
)

var _ = revel.Equal

var Queuer *queuer.Queuer
var Lender *lender.Lender

func Launch() {
	// Prometheus
	lender.RegisterPrometheus()
	queuer.RegisterPrometheus()

	state = core.NewState()
	lenderBot := lender.NewLender(state)
	queuerBot := queuer.NewQueuer(state, lenderBot)

	err := state.VerifyState()
	if err != nil {
		panic(err)
	}

	Queuer = queuerBot
	Lender = lenderBot

	// if revel.DevMode {
	// 	return
	// }

	// Start go lending
	go lenderBot.Start()
	go queuerBot.Start()
	go launchPrometheus(9911)
}

func LaunchFake() {
	// Prometheus
	lender.RegisterPrometheus()
	queuer.RegisterPrometheus()

	state = core.NewState()
	lenderBot := lender.NewLender(state)
	queuerBot := queuer.NewQueuer(state, lenderBot)

	err := state.VerifyState()
	if err != nil {
		panic(err)
	}

	Queuer = queuerBot
	Lender = lenderBot

	// if revel.DevMode {
	// 	return
	// }

	// Start go lending
	go lenderBot.Start()
	go queuerBot.Start()
	go launchPrometheus(9911)
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
