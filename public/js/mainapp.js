var app = angular.module('lendingApp', ['ngMaterial']);

app.controller('indexController', ['$scope', '$http', '$log', function($scope, $http, $log) {
	var indexScope = $scope;

	indexScope.login = function() {
		$http({
			method: 'POST',
			url: '/login',
			data : {
				email: indexScope.login.email,
				pass: indexScope.login.pass,
			}
		})
		.then((res) => {
			if (res.error) {
				//error in rpc
				$log.error("login: Error: [" + JSON.stringify(res.error) + "]");
			} else {
				//success
				$log.info("login: Success.");
			}
		}, (res) => {
			$log.error("login: Error: [" + JSON.stringify(err) + "]");
		})
		.then(() => {
			indexScope.login.email = "";
			indexScope.login.pass = "";	
		});
	}

	indexScope.register = function() {
		$http(
		{
			method: 'POST',
			url: '/register',
			data : {
				email: indexScope.register.email,
				pass: indexScope.register.pass,
			}
		})
		.then((res) => {
			if (res.error) {
				//error in rpc
				$log.error("register: Error: [" + JSON.stringify(res.error) + "]");
			} else {
				//success
				$log.info("register: Success.");
			}
		}, (res) => {
			$log.error("register: Error: [" + JSON.stringify(err) + "]");
		})
		.then(() => {
			indexScope.register.email = "";
			indexScope.register.pass = "";	
		});
	}
}]);