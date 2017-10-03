
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
						+ params[2].seriesName + ' : ' + params[2].value.toFixed(6) + '<br/>'
						+ params[3].seriesName + ' : ' + params[3].value.toFixed(6) + '<br/>'
						+ params[0].seriesName + ' : ' + params[0].value.toFixed(6) + '<br/>'
						+ params[1].seriesName + ' : ' + params[1].value.toFixed(6);
					}
				},
				legend: {
					selectedMode:false,
					data:['Poloniex Earned', 'Poloniex Fee', 'Bitfinex Earned', 'Bitfinex Fee']
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
					name:'Poloniex Fee',
					type:'bar',
					stack: 'poloniex',
					itemStyle: {
						normal: {
							color: '#ff4c4c',
							barBorderColor: '#ff4c4c',
							barBorderWidth: 6,
							barBorderRadius:0,
						}
					},
					data: (coinScope.isLendingHistoryCrypto ? coinScope.loanHistory.poloFee : coinScope.loanHistory.poloFeeDollar),
				},
				{
					name:'Poloniex Earned',
					type:'bar',
					stack: 'poloniex',
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
					data: (coinScope.isLendingHistoryCrypto ? coinScope.loanHistory.poloEarned : coinScope.loanHistory.poloEarnedDollar),
				},
				{
					name:'Bitfinex Fee',
					type:'bar',
					stack: 'bitfinex',
					itemStyle: {
						normal: {
							color: '#ff1a1a',
							barBorderColor: '#ff1a1a',
							barBorderWidth: 6,
							barBorderRadius:0,
						}
					},
					data: (coinScope.isLendingHistoryCrypto ? coinScope.loanHistory.bitfinexFee : coinScope.loanHistory.bitfinexFeeDollar),
				},
				{
					name:'Bitfinex Earned',
					type:'bar',
					stack: 'bitfinex',
					barCategoryGap: '50%',
					itemStyle: {
						normal: {
							color: '#2db92d',
							barBorderColor: '#2db92d',
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
					data: (coinScope.isLendingHistoryCrypto ? coinScope.loanHistory.bitfinexEarned : coinScope.loanHistory.bitfinexEarnedDollar),
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
					poloEarned = [],
					poloFee = [],
					bitfinexEarned = [],
					bitfinexFee = [],

					dates = [],

					poloEarnedDollar = [],
					poloFeeDollar = [],
					bitfinexEarnedDollar = [],
					bitfinexFeeDollar = [];
					// res.data.LoanHistory
					for(i = res.data.LoanHistory.length-1; i >= 0; i--) {
						if (new Date(res.data.LoanHistory[i].time).getFullYear() > 2000) {
							var pf = parseFloat(res.data.LoanHistory[i].poloniexdata[coinScope.coin].fees),
							pe = parseFloat(res.data.LoanHistory[i].poloniexdata[coinScope.coin].earned);
							bf = parseFloat(res.data.LoanHistory[i].bitfinexdata[coinScope.coin].fees),
							be = parseFloat(res.data.LoanHistory[i].bitfinexdata[coinScope.coin].earned);

							poloFee.push(pf);
							poloEarned.push(pe);
							bitfinexFee.push(bf);
							bitfinexEarned.push(be);
							var usdRate = res.data.USDRates["USD_"+coinScope.coin]
							if (usdRate == null) {
								usdRate = 1
							}
							console.log(res.data)

							poloFeeDollar.push(pf*parseFloat(usdRate));
							poloEarnedDollar.push(pe*parseFloat(usdRate));
							bitfinexFeeDollar.push(bf*parseFloat(usdRate));
							bitfinexEarnedDollar.push(be*parseFloat(usdRate));
							var t = new Date(res.data.LoanHistory[i].time)
							var mon = t.getMonth()+1;
							dates.push(mon + "/" + t.getDate());
						}
					}
					if (poloEarned.length == 0) {
						coinScope.hasCompleteLoans = false;
						return;
					}
					coinScope.loanHistory = {
						poloEarned : poloEarned,
						poloFee : poloFee,
						bitfinexEarned : bitfinexEarned,
						bitfinexFee : bitfinexFee,
						dates : dates,
						poloEarnedDollar : poloEarnedDollar,
						poloFeeDollar : poloFeeDollar,
						bitfinexEarnedDollar : bitfinexEarnedDollar,
						bitfinexFeeDollar : bitfinexFeeDollar,
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

				coinScope.detailedLendingHistory = false;

				for(i = 0; i < res.data.data.length; i++) {
					if (res.data.data[i] == null) {
						continue
					} else if (res.data.data[i].length == 0) {
						continue
					}
					coinScope.detailedLendingHistory = true;
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
						var avg = [unix, numberFix(a, 5)]
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
						var range = [unix, numberFix(lowest, 5), numberFix(highest, 5)]
						if (res.data.data[i][c].exchange == "bit") {
							bitfinAveragePoints.push(avg)
							bitfincRangePoints.push(range)
							bitfinLent.push([unix, numberFix(cur.availlent, 3)])
							bitfinNotLent.push([unix, numberFix(cur.availbal + cur.onorder, 3)])
						} else {
							poloAveragePoints.push(avg)
							poloRangePoints.push(range)
							poloLent.push([unix, numberFix(cur.availlent, 3)])
							poloNotLent.push([unix, numberFix(cur.availbal + cur.onorder, 3)])
						}
					}
				}


				initLineRangeGraph(poloAveragePoints, poloRangePoints, bitfinAveragePoints, bitfincRangePoints)
				initPercentLentGraph(bitfinLent, bitfinNotLent, poloLent, poloNotLent, coinScope.coin)


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

function numberFix(n, l) {
	return Number(n.toFixed(l))
}

function initPercentLentGraph(bitfinLent, bitfinNotLent, poloLent, poloNotLent, coin) {
	Highcharts.chart('lent-totals-graph', {
		colors: ['#33cc33', '#ff3300', '#248f24', '#b32400'],
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
			valueSuffix: " " + coin,
			thousandsSep: ','
		},

		legend: {
		},

		series: [{
			name: 'Poloniex Currency Lent',
			data: poloLent,
			zIndex: 1,
			marker: {
				fillColor: "#0a6970",
				lineWidth: 2,
				lineColor: "#0a6970"
			}
		}, {
			name: 'Poloniex Currency NotLent',
			data: poloNotLent,
			zIndex: 1,
			marker: {
				fillColor: "#0a6970",
				lineWidth: 2,
				lineColor: "#0a6970"
			}
		}, {
			name: 'Bitfinex Currency Lent',
			data: bitfinLent,
			zIndex: 1,
			marker: {
				fillColor: "#004d00",
				lineWidth: 2,
				lineColor: "#004d00"
			}
		}, {
			name: 'Bitfinex Currency NotLent',
			data: bitfinNotLent,
			zIndex: 1,
			marker: {
				fillColor: "#004d00",
				lineWidth: 2,
				lineColor: "#004d00"
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
			valueSuffix: '%',
			thousandsSep: ','
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