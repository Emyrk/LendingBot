var app=angular.module("lendingApp",["ngRoute"]);

app.config(['$routeProvider', '$locationProvider', function($routeProvider, $locationProvider) {
	$routeProvider
	.when("/",{
		templateUrl : "/dashboard/info",
		controller : "dashInfoController"
	})
	.when("/info/:coin",{
		templateUrl : "/dashboard/info/:id",
		controller : "dashInfoAdvancedController"
	})
	.when("/settings",{
		templateUrl : "/dashboard/settings",
		controller : "dashSettingsController"
	})
	.when("/logs",{
		templateUrl : "/dashboard/logs",
		controller : "dashLogsController"
	})
	.otherwise({redirectTo:'/'});
	
	$locationProvider.html5Mode({enabled: false, requireBase: false});
}]);

app.controller('dashBaseController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashBaseScope = $scope;

		dashBaseScope.getCurrentUserStats = function() {
			$http(
			{
				method: 'GET',
				url: '/dashboard/data/currentuserstats',
				data : {
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				console.log(res.data)
				dashBaseScope.currentUserStats = res.data.CurrentUserStats;
				dashBaseScope.balances = res.data.Balances;
				init_chart_doughnut(dashBaseScope.balances.currencymap)
			}, (err) => {
				//error
				$log.error("CurrentUserStats: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
			});
		}

		dashBaseScope.getLendingHistory = function() {
			$http(
			{
				method: 'GET',
				url: '/dashboard/data/lendinghistory',
				data : {
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				console.log(res.data)
				dashBaseScope.lendHist = res.data.CompleteLoans
				if (dashBaseScope.lendHist == null) {
					dashBaseScope.lendHist = [];
				}
			}, (err) => {
				//error
				$log.error("LendingHistory: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
			});
		}

		//init
		dashBaseScope.logout = LOC + "/logout";
		dashBaseScope.getCurrentUserStats();
		dashBaseScope.getLendingHistory();
		dashBaseScope.backgroundColor = backgroundColor;
		//----
	}]);

app.controller('dashInfoController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashInfoScope = $scope;
	}]);

app.controller('dashInfoAdvancedController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashInfoAdvScope = $scope;
	}]);

app.controller('dashSettingsController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		//init
		var dashSettingsScope = $scope;
		dashSettingsScope.pass2FA = '';
		dashSettingsScope.token = '';
		//-----

		dashSettingsScope.create2FA = function() {
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/create2fa',
				data : {
					pass: dashSettingsScope.pass2FA,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("2FA: Success.");
				dashSettingsScope.qrcode = 'data:image/png;base64,' + res.data.data
				dashSettingsScope.has2FA = true;
			}, (err) => {
				//error
				$log.error("2FA: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				dashSettingsScope.pass2FA = '';
			});
		}

		dashSettingsScope.enable2FA = function(bool) {
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/enable2fa',
				data : {
					pass: dashSettingsScope.pass2FA,
					token: dashSettingsScope.token,
					enable: bool,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("2FA: Success.");
				dashSettingsScope.enabled2FA = (res.data.data === 'true')
			}, (err) => {
				//error
				$log.error("2FA: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				dashSettingsScope.pass2FA = '';
				dashSettingsScope.token = '';
			});
		}

		dashSettingsScope.setPoloniexKeys = function() {
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/setpoloniexkeys',
				data : {
					poloniexkey: dashSettingsScope.poloniexKey,
					poloniexsecret: dashSettingsScope.poloniexSecret,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("SetPoloniexKeys: Success.");
				var tempData = JSON.parse(res.data.data);
				dashSettingsScope.poloniexKey = tempData.poloniexkey;
				dashSettingsScope.poloniexSecret = tempData.poloniexsecret;
			}, (err) => {
				//error
				$log.error("SetPoloniexKeys: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		dashSettingsScope.verifyEmail = function() {
			$http(
			{
				method: 'GET',
				url: '/verify/request',
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("VerifyEmail: Success.");
				dashSettingsScope.verifyEmail = true;
			}, (err) => {
				//error
				$log.error("VerifyEmail: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}
	}]);

app.controller('dashLogsController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashLogsScope = $scope;
	}]);



function init_chart_doughnut(data){
				
		if( typeof (Chart) === 'undefined'){ return; }
		
		console.log('init_chart_doughnut');
	 
		if ($('.canvasDoughnut').length){
		var chart_doughnut_settings = {
				type: 'doughnut',
				tooltipFillColor: "rgba(51, 51, 51, 0.55)",
				data: {
					labels: Object.keys(data),
					datasets: [{
						data: Object.values(data),
						backgroundColor: backgroundColor,
					}]
				},
				options: { 
					legend: false, 
					responsive: false 
				}
			}
		
			$('.canvasDoughnut').each(function(){
				
				var chart_element = $(this);
				var chart_doughnut = new Chart( chart_element, chart_doughnut_settings);
				
			});			
		
		}  
	   
	}

var backgroundColor =[
	"#00BFFF",
	"#FF69B4",
	"#7CFC00",
	"#800000",
	"#FFA500",
	"#FF4500",
	"#800080",
	"#00FF7F",
	"#FFFF00",
	"#9ACD32",
	"#FF6347"
]