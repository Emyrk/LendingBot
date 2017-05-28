package lender

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CurrentLoanRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate",
		Help: "Shows the current lending rate when it is calculated",
	})

	LoansCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlzone_lender_loans_created_count",
		Help: "Count of loans created",
	})

	LoansCanceled = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlzone_lender_loans_canceled_count",
		Help: "Count of loans created",
	})

	JobsDone = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlzone_lender_jobs_complete",
		Help: "The counter of how many jobs are done",
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

	prometheus.MustRegister(CurrentLoanRate)
	prometheus.MustRegister(JobsDone)
	prometheus.MustRegister(LoansCreated)
	prometheus.MustRegister(LoansCanceled)
}
