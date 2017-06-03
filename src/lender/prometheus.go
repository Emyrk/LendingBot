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

	// Poloniex Stats
	//		Avg
	PoloniexStatsHourlyAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_avg_hourly",
		Help: "Hourly Avg for poloniex data",
	})

	PoloniexStatsDailyAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_avg_daily",
		Help: "Daily Avg for poloniex data",
	})

	PoloniexStatsWeeklyAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_avg_weekly",
		Help: "Weekly Avg for poloniex data",
	})

	PoloniexStatsMonthlyAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_avg_monthly",
		Help: "Monthly Avg for poloniex data",
	})

	//		Std
	PoloniexStatsHourlyStd = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_std_hourly",
		Help: "Hourly Std for poloniex data",
	})

	PoloniexStatsDailyStd = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_std_daily",
		Help: "Daily Std for poloniex data",
	})

	PoloniexStatsWeeklyStd = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_std_weekly",
		Help: "Weekly Std for poloniex data",
	})

	PoloniexStatsMonthlyStd = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_std_monthly",
		Help: "Monthly Std for poloniex data",
	})

	// Update Ticker
	LenderUpdateTicker = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_lender_update_ticker",
		Help: "Every ticker update",
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
	prometheus.MustRegister(PoloniexStatsHourlyAvg)
	prometheus.MustRegister(PoloniexStatsDailyAvg)
	prometheus.MustRegister(PoloniexStatsWeeklyAvg)
	prometheus.MustRegister(PoloniexStatsMonthlyAvg)
	prometheus.MustRegister(PoloniexStatsHourlyStd)
	prometheus.MustRegister(PoloniexStatsDailyStd)
	prometheus.MustRegister(PoloniexStatsWeeklyStd)
	prometheus.MustRegister(PoloniexStatsMonthlyStd)
	prometheus.MustRegister(LenderUpdateTicker)
}
