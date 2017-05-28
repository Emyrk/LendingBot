package queuer

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	QueuerCycles = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlzone_queuer_cycles_count",
		Help: "How many times the queuer enters it's decision cycle",
	})

	QueuerJobsMade = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlzone_queuer_newjobs_count",
		Help: "How many jobs the queuer queues",
	})

	QueuerTotalUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_queuer_users_total",
		Help: "How many users the queuer sees",
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

	prometheus.MustRegister(QueuerCycles)
	prometheus.MustRegister(QueuerJobsMade)
	prometheus.MustRegister(QueuerTotalUsers)
}
