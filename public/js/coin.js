
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
							var f = parseFloat(res.data.LoanHistory[i].data[coinScope.coin].fees),
							e = parseFloat(res.data.LoanHistory[i].data[coinScope.coin].earned);
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
				var averagePoints = [];
				var rangePoints = [];
				// for(i =0; i < 10; i++) {
				// 	averagePoints.push([i, i])
				// 	rangePoints.push([i, i,i+5])
				// }

				for(i = 0; i < res.data.data.length; i++) {
					if (res.data.data[i] == undefined) {
						continue
					}
					for(c = 0; c < res.data.data[i].length; c++) {
						var cur = res.data.data[i][c].currencies[coinScope.coin]
						var unix = new Date(cur.time).getTime()
						if (cur.activerate > 2) {
							continue
						}
						var avg = [unix, (cur.activerate*100)]
						var range = [unix, (cur.lowestrate*100), (cur.highestrate*100)]
						averagePoints.push(avg)
						rangePoints.push(range)
					}
				}


				initLineRangeGraph(averagePoints, rangePoints)


			}, (err) => {
				//error
				$log.error("LendingHistory: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
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

function initLineRangeGraph(averages, ranges) {
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
			name: 'Average Rate',
			data: averages,
			zIndex: 1,
			marker: {
				fillColor: 'white',
				lineWidth: 2,
				lineColor: Highcharts.getOptions().colors[0]
			}
		}, {
			name: 'Range',
			data: ranges,
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