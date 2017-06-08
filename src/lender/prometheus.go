package lender

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Lending Rates
	CurrentLoanRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate",
		Help: "Average based lend rate",
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

	// Tickers
	TickerFCTValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_fct_value",
		Help: "FCT_BTC",
	})

	TickerBTSValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_bts_value",
		Help: "BTS_BTC",
	})

	TickerCLAMValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_clam_value",
		Help: "CLAM_BTC",
	})

	TickerDOGEValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_doge_value",
		Help: "DOGE_BTC",
	})

	TickerLTCValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_ltc_value",
		Help: "LTC_BTC",
	})

	TickerMAIDValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_maid_value",
		Help: "MAID_BTC",
	})

	TickerSTRValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_str_value",
		Help: "STR_BTC",
	})

	TickerXMRValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_xmr_value",
		Help: "XMR_BTC",
	})

	TickerXRPValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_xrp_value",
		Help: "XRP_BTC",
	})

	TickerETHValue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlezone_lender_eth_value",
		Help: "ETH_BTC",
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
	prometheus.MustRegister(LenderCurrentAverageBasedRate)

	prometheus.MustRegister(TickerFCTValue)
	prometheus.MustRegister(TickerBTSValue)
	prometheus.MustRegister(TickerCLAMValue)
	prometheus.MustRegister(TickerDOGEValue)
	prometheus.MustRegister(TickerLTCValue)
	prometheus.MustRegister(TickerMAIDValue)
	prometheus.MustRegister(TickerSTRValue)
	prometheus.MustRegister(TickerXMRValue)
	prometheus.MustRegister(TickerXRPValue)
	prometheus.MustRegister(TickerETHValue)
}
