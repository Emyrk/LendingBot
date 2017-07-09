
app.controller('coinController', ['$scope', '$http', '$log', '$timeout','$routeParams',
	function($scope, $http, $log, $timeout, $routeParams) {
		var coinScope = $scope;
		var earnedFeeChart;

		coinScope.resetPoloniexKeys = function() {
			coinScope.poloniexKey = coinScope.poloniexKeyOrig;
			coinScope.poloniexSecret = coinScope.poloniexSecretOrig;
		}

		coinScope.getLendingHistoryOption = function() {
			return  {
				dataZoom : {
					show : true,
					realtime: true,
					start : 50,
					end : 100
				},
				tooltip : {
					trigger: 'axis',
					axisPointer : {
						type : 'shadow'
					},
					formatter: function (params){
						return "Date: " + params[0].name + '<br/>'
						+ params[0].seriesName + ' : ' + params[0].value.toFixed(6) + '<br/>'
						+ params[1].seriesName + ' : ' + params[1].value.toFixed(6);
					}
				},
				legend: {
					selectedMode:false,
					data:['Earned', 'Fee']
				},
				toolbox: {
					show : true,
					feature : {
						mark : {show: true},
						dataView : {show: true, readOnly: false},
						restore : {show: true},
						saveAsImage : {show: true}
					}
				},
				calculable : true,
				xAxis : [
				{
					type : 'category',
					data : dates,
					name : "Month/Day",
				}
				],
				yAxis : [
				{
					type : 'value',
					boundaryGap: [0, 0.1],
					name :  (coinScope.isLendingHistoryCrypto ? coinScope.coin : "Dollar") + " Amount",
				}
				],
				series : [
				{
					name:'Fee',
					type:'bar',
					stack: 'sum',
					itemStyle: {
						normal: {
							color: '#ff4c4c',
							barBorderColor: '#ff4c4c',
							barBorderWidth: 6,
							barBorderRadius:0,
						}
					},
					data: (coinScope.isLendingHistoryCrypto ? coinScope.loanHistory.fee : coinScope.loanHistory.feeDollar),
				},
				{
					name:'Earned',
					type:'bar',
					stack: 'sum',
					barCategoryGap: '50%',
					itemStyle: {
						normal: {
							color: '#46D246',
							barBorderColor: '#46D246',
							barBorderWidth: 6,
							barBorderRadius:0,
							label : {
								show: true, position: 'top',
								formatter: function (params) {
									return params.value.toFixed(5);
								},
							}
						}
					},
					data: (coinScope.isLendingHistoryCrypto ? coinScope.loanHistory.earned : coinScope.loanHistory.earnedDollar),
				},
				]
			}
		}

		coinScope.getLendingHistory = function() {
			$http(
			{
				method: 'GET',
				url: '/dashboard/data/lendinghistorysummary', //' + coinScope.coin
				data : {
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				coinScope.hasCompleteLoans = res.data.LoanHistory ? true : false;
				if (coinScope.hasCompleteLoans) {
					earnedFeeChart = echarts.init(document.getElementById('lendingHistoryChart')),
					earned = [],
					fee = [],
					dates = [],
					earnedDollar = [],
					feeDollar = [];
					// res.data.LoanHistory
					for(i = res.data.LoanHistory.length-1; i >= 0; i--) {
						if (new Date(res.data.LoanHistory[i].time).getFullYear() > 2000) {
							var f = parseFloat(res.data.LoanHistory[i].poloniexdata[coinScope.coin].fees),
							e = parseFloat(res.data.LoanHistory[i].poloniexdata[coinScope.coin].earned);
							fee.push(f);
							earned.push(e);
							var usdRate = res.data.USDRates["USD_"+coinScope.coin]
							if (usdRate == null) {
								usdRate = 1
							}
							console.log(res.data)

							feeDollar.push(f*parseFloat(usdRate));
							earnedDollar.push(e*parseFloat(usdRate));
							var t = new Date(res.data.LoanHistory[i].time)
							var mon = t.getMonth()+1;
							dates.push(mon + "/" + t.getDate());
						}
					}
					if (earned.length == 0) {
						coinScope.hasCompleteLoans = false;
						return;
					}
					coinScope.loanHistory = {
						earned : earned,
						fee : fee,
						dates : dates,
						earnedDollar : earnedDollar,
						feeDollar : feeDollar,
					}
					
					earnedFeeChart.setOption(coinScope.getLendingHistoryOption());
					$timeout(() => {
						earnedFeeChart.resize();
					});
				}
			}, (err) => {
				//error
				$log.error("LendingHistory: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		coinScope.getDetailedLendingHistory = function() {
			coinScope.loadingDetailedLendingHistory = true;
			$http(
			{
				method: 'GET',
				url: '/dashboard/data/detstats',
				data : {
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				var poloAveragePoints = [];
				var poloRangePoints = [];
				var bitfinAveragePoints = [];
				var bitfincRangePoints = [];

				var poloLent = [];
				var poloNotLent = [];
				var bitfinLent = [];
				var bitfinNotLent = [];
				// for(i =0; i < 10; i++) {
				// 	averagePoints.push([i, i])
				// 	rangePoints.push([i, i,i+5])
				// }

				coinScope.detailedLendingHistory = res.data.data ? true : false;

				for(i = 0; i < res.data.data.length; i++) {
					if (res.data.data[i] == undefined) {
						continue
					}
					var prevLowest = 0
					for(c = 0; c < res.data.data[i].length; c++) {
						var cur = res.data.data[i][c].currencies[coinScope.coin]
						if (cur == undefined || cur == null) {
							continue
						}
						var unix = new Date(cur.time).getTime()
						if ((cur.activerate*100) > 2 ||  cur.activerate == 0){
							continue
						}
						var a = cur.activerate*100
						var avg = [unix, numberFix(a)]
						var lowest = cur.lowestrate*100
						if(lowest == 0) {
							lowest = prevLowest
						}
						if(lowest == 0) {
							lowest = a
						}
						prevLowest = lowest
						var highest = cur.highestrate*100
						if (highest == 0) {
							highest = a
						}
						var range = [unix, numberFix(lowest), numberFix(highest)]
						if (res.data.data[i][c].exchange == "bit") {
							bitfinAveragePoints.push(avg)
							bitfincRangePoints.push(range)
							bitfinLent.push([unix, cur.availlent])
							bitfinNotLent.push([unix, cur.availbal + cur.onorder])
						} else {
							poloAveragePoints.push(avg)
							poloRangePoints.push(range)
							poloLent.push([unix, cur.availlent])
							poloNotLent.push([unix, cur.availbal + cur.onorder])
						}
					}
				}


				initLineRangeGraph(poloAveragePoints, poloRangePoints, bitfinAveragePoints, bitfincRangePoints)
				initPercentLentGraph(bitfinLent, bitfinNotLent, poloLent, poloNotLent)


			}, (err) => {
				//error
				$log.error("LendingHistory: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				coinScope.loadingDetailedLendingHistory = false;
			});
		}

		coinScope.swapLendingHistoryType = function() {
			coinScope.isLendingHistoryCrypto = !coinScope.isLendingHistoryCrypto;
			earnedFeeChart.setOption(coinScope.getLendingHistoryOption());
		}


		// /Coin
		coinScope.coin = $routeParams.coin;
		coinScope.isLendingHistoryCrypto = true;
		coinScope.getLendingHistory();
		coinScope.getDetailedLendingHistory();
		//resize charts
		window.onresize = function() {
			if (coinScope.hasCompleteLoans) {
				$timeout(() => {
					earnedFeeChart.resize();
				});
			}
		};
		//------

	}]);

function numberFix(n) {
	return Number(n.toFixed(5))
}

function initPercentLentGraph(bitfinLent, bitfinNotLent, poloLent, poloNotLent) {
	Highcharts.chart('lent-totals-graph', {

		title: {
			text: 'Total Currency Being Lent'
		},

		xAxis: {
			type: 'datetime'
		},

		yAxis: {
			title: {
				text: null
			}
		},

		tooltip: {
			crosshairs: true,
			shared: true,
			valueSuffix: '%'
		},

		legend: {
		},

		series: [{
			name: 'Poloniex Currency Lent',
			data: poloLent,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}, {
			name: 'Poloniex Currency NotLent',
			data: poloNotLent,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}, {
			name: 'Bitfinex Currency Lent',
			data: bitfinLent,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}, {
			name: 'Bitfinex Currency NotLent',
			data: bitfinNotLent,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}]
	});	
}

function initLineRangeGraph(poloAverages, poloRanges, bitfinexAvgerages, bitfinexRanges) {
	Highcharts.chart('lending-rate-graph', {

		title: {
			text: 'Lending Rates in Percent'
		},

		xAxis: {
			type: 'datetime'
		},

		yAxis: {
			title: {
				text: null
			}
		},

		tooltip: {
			crosshairs: true,
			shared: true,
			valueSuffix: '%'
		},

		legend: {
		},

		series: [{
			name: 'Poloniex Average Rate',
			data: poloAverages,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}, {
			name: 'Poloniex Range',
			data: poloRanges,
			type: 'arearange',
			lineWidth: 0,
			linkedTo: ':previous',
			color: Highcharts.getOptions().colors[0],
			fillOpacity: 0.3,
			zIndex: 0,
			marker: {
				enabled: false
			}
		},
		{
			name: 'Bitfinex Average Rate',
			data: bitfinexAvgerages,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}, {
			name: 'Bitfinex Range',
			data: bitfinexRanges,
			type: 'arearange',
			lineWidth: 0,
			linkedTo: ':previous',
			color: Highcharts.getOptions().colors[0],
			fillOpacity: 0.3,
			zIndex: 0,
			marker: {
				enabled: false
			}
		}]
	});
}