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
		dashSettingsScope.pass2fa = '';
		dashSettingsScope.enablePass = '';
		dashSettingsScope.disablePass = '';
		dashSettingsScope.qrcode = '';
		//-----
		dashSettingsScope.create2FA = function() {
			$http(
			{
				method: 'POST',
				url: '/dashboard/2fa/create2fa',
				data : {
					pass: dashSettingsScope.pass2fa,
				},
				withCredentials: true
			})
			.then((res, status, headers, config) => {
				//success
				$log.info("2fa: Success.");
				dashSettingsScope.qrcode = 'data:image/png;base64,' + res.data.data
				dashSettingsScope.has2FA = true;
			}, (err, status, headers, config) => {
				//error
				$log.error("2fa: Error: [" + JSON.stringify(err) + "] Status [" + status + "]");
			})
			.then(() => {
				dashSettingsScope.pass2fa = '';
			});
		}

		dashSettingsScope.disable2FA = function() {
			
			dashSettingsScope.disablePass = '';
		}
	}]);

app.controller('dashLogsController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashLogsScope = $scope;
	}]);