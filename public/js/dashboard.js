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
		dashBaseScope.logout = LOC + "/logout"
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
			.then((res, status, headers, config) => {
				//success
				$log.info("2FA: Success.");
				dashSettingsScope.qrcode = 'data:image/png;base64,' + res.data.data
				dashSettingsScope.has2FA = true;
			}, (err, status, headers, config) => {
				//error
				$log.error("2FA: Error: [" + JSON.stringify(err) + "] Status [" + status + "]");
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
			.then((res, status, headers, config) => {
				//success
				$log.info("2FA: Success.");
				dashSettingsScope.enabled2FA = (res.data.data === 'true')
			}, (err, status, headers, config) => {
				//error
				$log.error("2FA: Error: [" + JSON.stringify(err) + "] Status [" + status + "]");
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
			.then((res, status, headers, config) => {
				//success
				$log.info("SetPoloniexKeys: Success.");
				var tempData = JSON.parse(res.data.data);
				dashSettingsScope.poloniexKey = tempData.poloniexkey;
				dashSettingsScope.poloniexSecret = tempData.poloniexsecret;
			}, (err, status, headers, config) => {
				//error
				$log.error("SetPoloniexKeys: Error: [" + JSON.stringify(err) + "] Status [" + status + "]");
			});
		}
	}]);

app.controller('dashLogsController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashLogsScope = $scope;
	}]);