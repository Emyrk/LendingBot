package core

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	PoloPublicCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_poloniex_public_calls_total",
		Help: "Number of public polo calls",
	})
)

var registered bool = false

// RegisterPrometheus registers the variables to be exposed. This can only be run once, hence the
// boolean flag to prevent panics if launched more than once. This is called in NetStart
func RegisterPrometheus() {
	if registered {
		return
	}
	registered = true

	prometheus.MustRegister(PoloPublicCalls)
}
