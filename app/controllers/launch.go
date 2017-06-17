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

const (
	DEV_FAKE  = "devFake"
	DEV_EMPTY = "devEmpty"
)

var _ = revel.Equal

var Queuer *queuer.Queuer
var Lender *lender.Lender

func Launch() {

	fmt.Println("MODE IS: ", revel.RunMode)
	if revel.RunMode == DEV_FAKE {
		//devFake mode
		//should be all in memory with user account

		//user: a@a pass:a
		//should be commonuser level
		//should have populated data

		//user: b@b pass:b
		//should be commonuser level
		//should have empty data

		//user: admin@admin pass:admin
		//should be sysadmin level
		//should have populated data

		//should be mainly used for gui creation
	} else if revel.RunMode == DEV_EMPTY {
		//devEmpty mode
		//should be all in memory with empty data

		//to be used for unit testing/regression testing
	} else {
		//dev Mode Normal

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
}

func launchPrometheus(port int) {
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
