package bee

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var _ = fmt.Println

var (

	// Polo Bot
	CompromisedBTC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_compromise_rate_btc",
		Help: "Compromised rate",
	})

	// Jobs
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

	JobPart1 = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "hodlzone_lender_job_part1_ns",
		Help: "Part 1 of job",
	})

	JobPart2 = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "hodlzone_lender_job_part2_ns",
		Help: "Part 2 of job",
	})

	// Jobs
	JobQueueCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_job_queue_length",
		Help: "Number of jobs to be processed",
	})

	JobProcessDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "hodlzone_lender_job_duration",
		Help: "How long to process a Job",
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

	prometheus.MustRegister(LoansCreated)
	prometheus.MustRegister(LoansCanceled)
	prometheus.MustRegister(JobsDone)
	prometheus.MustRegister(JobPart1)
	prometheus.MustRegister(JobPart2)
	prometheus.MustRegister(JobQueueCurrent)
	prometheus.MustRegister(JobProcessDuration)
	prometheus.MustRegister(CompromisedBTC)
}
