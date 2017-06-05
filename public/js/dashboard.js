var app=angular.module("lendingApp",["ngRoute","ngMask"]);

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
	.when("/settings/user",{
		templateUrl : "/dashboard/settings/user",
		controller : "dashSettingsUserController"
	})
	.when("/settings/lending",{
		templateUrl : "/dashboard/settings/lending",
		controller : "dashSettingsLendingController"
	})
	.when("/logs",{
		templateUrl : "/dashboard/logs",
		controller : "dashLogsController"
	})
	.otherwise({redirectTo:'/'});
	
	$locationProvider.html5Mode({enabled: false, requireBase: false});
}]);

app.controller('dashBaseController', ['$scope', '$http', '$log', "$location",
	function($scope, $http, $log, $location) {
		var dashBaseScope = $scope;
		//init
		dashBaseScope.logout = LOC + "/logout";
		//----
	}]);

app.controller('dashInfoController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashInfoScope = $scope;

		dashInfoScope.getCurrentUserStats = function() {
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
				dashInfoScope.currentUserStats = res.data.CurrentUserStats;
				dashInfoScope.balances = res.data.Balances;
				init_chart_doughnut(dashInfoScope.balances.currencymap)
			}, (err) => {
				//error
				$log.error("CurrentUserStats: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		dashInfoScope.getLendingHistory = function() {
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
				console.log(res.data);
				$.each(res.data.CompleteLoans, (index, val) => {
					var tr = $("<tr>")
					.append($("<td>").html(val.currency))
					.append($("<td>").html(parseFloat(val.rate).toFixed(4)))
					.append($("<td>").html(val.amount))
					.append($("<td>").html(val.earned))
					.append($("<td>").html(val.fee))
					.append($("<td>").html(val.close))
					.append($("<td>").html(val.duration));
					$("#lendingHistory").append(tr);
				});
				if (!$.fn.DataTable.isDataTable('#datatable-responsive')) {
					$('#datatable-responsive').DataTable({
						filter: false
					});
				}
			}, (err) => {
				//error
				$log.error("LendingHistory: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		//init
		dashInfoScope.getCurrentUserStats();
		dashInfoScope.getLendingHistory();
		dashInfoScope.backgroundColor = backgroundColor;
		//----
	}]);

app.controller('dashInfoAdvancedController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashInfoAdvScope = $scope;
	}]);

app.controller('dashSettingsUserController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		//init
		var dashSettingsUserScope = $scope;
		dashSettingsUserScope.pass2FA = '';
		dashSettingsUserScope.token = '';
		//-----

		dashSettingsUserScope.create2FA = function() {
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/create2fa',
				data : {
					pass: dashSettingsUserScope.pass2FA,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("2FA: Success.");
				dashSettingsUserScope.qrcode = 'data:image/png;base64,' + res.data.data
				dashSettingsUserScope.has2FA = true;
			}, (err) => {
				//error
				$log.error("2FA: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				dashSettingsUserScope.pass2FA = '';
			});
		}

		dashSettingsUserScope.enable2FA = function(bool) {
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/enable2fa',
				data : {
					pass: dashSettingsUserScope.pass2FA,
					token: dashSettingsUserScope.token,
					enable: bool,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("2FA: Success.");
				dashSettingsUserScope.enabled2FA = (res.data.data === 'true')
			}, (err) => {
				//error
				$log.error("2FA: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				dashSettingsUserScope.pass2FA = '';
				dashSettingsUserScope.token = '';
			});
		}

		dashSettingsUserScope.setPoloniexKeys = function() {
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/setpoloniexkeys',
				data : {
					poloniexkey: dashSettingsUserScope.poloniexKey,
					poloniexsecret: dashSettingsUserScope.poloniexSecret,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("SetPoloniexKeys: Success.");
				var tempData = JSON.parse(res.data.data);
				dashSettingsUserScope.poloniexKey = tempData.poloniexkey;
				dashSettingsUserScope.poloniexSecret = tempData.poloniexsecret;
			}, (err) => {
				//error
				$log.error("SetPoloniexKeys: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		dashSettingsUserScope.verifyEmail = function() {
			$http(
			{
				method: 'GET',
				url: '/verify/request',
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("VerifyEmail: Success.");
				dashSettingsUserScope.verifyEmail = true;
			}, (err) => {
				//error
				$log.error("VerifyEmail: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}
	}]);

app.controller('dashSettingsLendingController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashSettingsLendingScope = $scope;

		dashSettingsLendingScope.resetPoloniexKeys = function() {
				dashSettingsLendingScope.poloniexKeyOrig = dashSettingsLendingScope.poloniexKey;
				dashSettingsLendingScope.poloniexSecretOrig = dashSettingsLendingScope.poloniexSecret;
		}

		dashSettingsLendingScope.setPoloniexKeys = function() {
			dashSettingsLendingScope.loadingPoloniexKeys = true;
			$http(
			{
				method: 'POST',
				url: '/dashboard/settings/setpoloniexkeys',
				data : $.param({
					poloniexkey: dashSettingsLendingScope.poloniexKey,
					poloniexsecret: dashSettingsLendingScope.poloniexSecret,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("SetPoloniexKeys: Success.");
				var tempData = JSON.parse(res.data.data);
				dashSettingsLendingScope.poloniexKeyOrig = tempData.poloniexkey;
				dashSettingsLendingScope.poloniexSecretOrig = tempData.poloniexsecret;
				//resets to new originals
				dashSettingsLendingScope.resetPoloniexKeys();
			}, (err) => {
				//error
				$log.error("SetPoloniexKeys: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				dashSettingsLendingScope.loadingPoloniexKeys = false;
			});
		}

		//init
		init_InputMask();
		dashSettingsLendingScope.pass2FA = '';
		dashSettingsLendingScope.token = '';
		dashSettingsLendingScope.loadingPoloniexKeys = false;
		//------

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

function init_InputMask() {

	if( typeof ($.fn.inputmask) === 'undefined'){ return; }
	console.log('init_InputMask');

	$(":input").inputmask();

};

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