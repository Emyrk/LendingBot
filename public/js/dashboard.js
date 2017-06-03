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
				dashBaseScope.Balances = res.data.Balances;

				console.log(Math.abs(dashBaseScope.currentUserStats.loanratechange).toFixed(3) <= 0.00)
			}, (err) => {
				//error
				$log.error("CurrentUserStats: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
			});
		}

		//init
		dashBaseScope.logout = LOC + "/logout";
		dashBaseScope.getCurrentUserStats();
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
					labels: [
						"Symbian",
						"Blackberry",
						"Other",
						"Android",
						"IOS"
					],
					datasets: [{
						data: [15, 20, 30, 10, 30],
						backgroundColor: [
							"#BDC3C7",
							"#9B59B6",
							"#E74C3C",
							"#26B99A",
							"#3498DB"
						],
						hoverBackgroundColor: [
							"#CFD4D8",
							"#B370CF",
							"#E95E4F",
							"#36CAAB",
							"#49A9EA"
						]
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