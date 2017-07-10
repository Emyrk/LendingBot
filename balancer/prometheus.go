package balancer

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var _ = fmt.Println

var (
	PoloPrivateCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_poloniex_private_calls_total",
		Help: "Number of public polo calls",
	})

	PoloCallTakeWait = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "hodlzone_balancer_poloniex_take_wait",
		Help: "Wait for a polo call",
	})

	PoloPublicCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hodlezone_poloniex_public_calls_total",
		Help: "Number of public polo calls",
	})

	// Polo Bot
	CompromisedBTC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_compromise_rate_btc",
		Help: "Compromised rate",
	})

	PoloBotRateBTC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_btc",
		Help: "BTC For polobot",
	})

	PoloBotRateETH = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_eth",
		Help: "BTC For polobot",
	})

	PoloBotRateXMR = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_xmr",
		Help: "BTC For polobot",
	})

	PoloBotRateXRP = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_xrp",
		Help: "BTC For polobot",
	})

	PoloBotRateDASH = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_dash",
		Help: "BTC For polobot",
	})

	PoloBotRateLTC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_ltc",
		Help: "BTC For polobot",
	})

	PoloBotRateDOGE = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_doge",
		Help: "BTC For polobot",
	})

	PoloBotRateBTS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_polobot_rate_bts",
		Help: "BTC For polobot",
	})

	// Lending Rates
	CurrentLoanRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate",
		Help: "Average based lend rate",
	})

	// Lending Rates Other
	CurrentLoanRateBTS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_bts",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateBTS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_bts",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateCLAM = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_clam",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateCLAM = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_clam",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateDOGE = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_doge",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateDOGE = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_doge",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateDASH = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_dash",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateDASH = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_dash",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateLTC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_ltc",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateLTC = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_ltc",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateMAID = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_maid",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateMAID = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_maid",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateSTR = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_str",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateSTR = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_str",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateXMR = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_xmr",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateXMR = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_xmr",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateXRP = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_xrp",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateXRP = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_xrp",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateETH = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_eth",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateETH = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_eth",
		Help: "Average based lend rate",
	})

	//
	CurrentLoanRateFCT = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_lend_rate_fct",
		Help: "Shows the current lending rate when it is calculated",
	})

	LenderCurrentAverageBasedRateFCT = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_current_avgbased_lend_rate_fct",
		Help: "Average based lend rate",
	})

	// Poloniex Stats
	//		Avg
	PoloniexStatsFiveMinAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hodlzone_lender_poloniex_stats_avg_fivemin",
		Help: "Hourly Avg for poloniex data",
	})

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
	prometheus.MustRegister(CompromisedBTC)

	prometheus.MustRegister(CurrentLoanRate)
	prometheus.MustRegister(PoloniexStatsFiveMinAvg)
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

	prometheus.MustRegister(CurrentLoanRateBTS)
	prometheus.MustRegister(LenderCurrentAverageBasedRateBTS)
	prometheus.MustRegister(CurrentLoanRateCLAM)
	prometheus.MustRegister(LenderCurrentAverageBasedRateCLAM)
	prometheus.MustRegister(CurrentLoanRateDOGE)
	prometheus.MustRegister(LenderCurrentAverageBasedRateDOGE)
	prometheus.MustRegister(CurrentLoanRateDASH)
	prometheus.MustRegister(LenderCurrentAverageBasedRateDASH)
	prometheus.MustRegister(CurrentLoanRateLTC)
	prometheus.MustRegister(LenderCurrentAverageBasedRateLTC)
	prometheus.MustRegister(CurrentLoanRateMAID)
	prometheus.MustRegister(LenderCurrentAverageBasedRateMAID)
	prometheus.MustRegister(CurrentLoanRateSTR)
	prometheus.MustRegister(LenderCurrentAverageBasedRateSTR)
	prometheus.MustRegister(CurrentLoanRateXMR)
	prometheus.MustRegister(LenderCurrentAverageBasedRateXMR)
	prometheus.MustRegister(CurrentLoanRateXRP)
	prometheus.MustRegister(LenderCurrentAverageBasedRateXRP)
	prometheus.MustRegister(CurrentLoanRateETH)
	prometheus.MustRegister(LenderCurrentAverageBasedRateETH)
	prometheus.MustRegister(CurrentLoanRateFCT)
	prometheus.MustRegister(LenderCurrentAverageBasedRateFCT)

	prometheus.MustRegister(PoloBotRateBTC)
	prometheus.MustRegister(PoloBotRateETH)
	prometheus.MustRegister(PoloBotRateXMR)
	prometheus.MustRegister(PoloBotRateXRP)
	prometheus.MustRegister(PoloBotRateDASH)
	prometheus.MustRegister(PoloBotRateLTC)
	prometheus.MustRegister(PoloBotRateDOGE)
	prometheus.MustRegister(PoloBotRateBTS)

	prometheus.MustRegister(PoloPrivateCalls)
	prometheus.MustRegister(PoloCallTakeWait)
	prometheus.MustRegister(PoloPublicCalls)
}

func SetSimple(currency string, rate float64) {
	switch currency {
	case "BTC":
		CurrentLoanRate.Set(rate)
	case "BTS":
		CurrentLoanRateBTS.Set(rate)
	case "CLAM":
		CurrentLoanRateCLAM.Set(rate)
	case "DOGE":
		CurrentLoanRateDOGE.Set(rate)
	case "DASH":
		CurrentLoanRateDASH.Set(rate)
	case "LTC":
		CurrentLoanRateLTC.Set(rate)
	case "MAID":
		CurrentLoanRateMAID.Set(rate)
	case "STR":
		CurrentLoanRateSTR.Set(rate)
	case "XMR":
		CurrentLoanRateXMR.Set(rate)
	case "XRP":
		CurrentLoanRateXRP.Set(rate)
	case "ETH":
		CurrentLoanRateETH.Set(rate)
	case "FCT":
		CurrentLoanRateFCT.Set(rate)
	}
}

func SetAvg(currency string, rate float64) {
	switch currency {
	case "BTC":
		LenderCurrentAverageBasedRate.Set(rate)
	case "BTS":
		LenderCurrentAverageBasedRateBTS.Set(rate)
	case "CLAM":
		LenderCurrentAverageBasedRateCLAM.Set(rate)
	case "DOGE":
		LenderCurrentAverageBasedRateDOGE.Set(rate)
	case "DASH":
		LenderCurrentAverageBasedRateDASH.Set(rate)
	case "LTC":
		LenderCurrentAverageBasedRateLTC.Set(rate)
	case "MAID":
		LenderCurrentAverageBasedRateMAID.Set(rate)
	case "STR":
		LenderCurrentAverageBasedRateSTR.Set(rate)
	case "XMR":
		LenderCurrentAverageBasedRateXMR.Set(rate)
	case "XRP":
		LenderCurrentAverageBasedRateXRP.Set(rate)
	case "ETH":
		LenderCurrentAverageBasedRateETH.Set(rate)
	case "FCT":
		LenderCurrentAverageBasedRateFCT.Set(rate)
	}
}
