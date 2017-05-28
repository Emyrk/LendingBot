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
	.when("/sysadmin",{
		templateUrl : "/dashboard/sysadmin",
		controller : "sysAdminController"
	})
	.when("/admin",{
		templateUrl : "/dashboard/admin",
		controller : "dashAdminController"
	})
	.otherwise({redirectTo:'/'});
	
	$locationProvider.html5Mode({enabled: false, requireBase: false});
}]);

app.controller('dashBaseController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashBaseScope = $scope;

		//removes all cookies and then redirects to index
		dashBaseScope.logout = function() {
			var cookies = $.cookie();
			for(var cookie in cookies) {
				$.removeCookie(cookie);
			}
			window.location = LOC;
		}
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
		var dashSettingsScope = $scope;
		$log.info("HERE");
	}]);

app.controller('dashSysAdminController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var dashSysAdminScope = $scope;
	}]);

app.controller('dashAdminController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var adminScope = $scope;
	}]);