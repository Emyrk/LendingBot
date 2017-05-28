var app = angular.module('lendingApp', ['ngMaterial']);

app.controller('indexController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var indexScope = $scope;

		indexScope.login = function() {
			$http({
				method: 'POST',
				url: '/login',
				data : {
					email: indexScope.login.email,
					pass: indexScope.login.pass,
				},
				withCredentials: true
			})
			.then((res, status, headers, config) => {
			//success
			$log.info("login: Success.");
			window.location = LOC + '/dashboard';
		}, (err, status, headers, config) => {
			//error
			$log.error("login: Error: [" + JSON.stringify(err) + "] Status [" + status + "]");
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
				},
				withCredentials: true
			})
			.then((res, status, headers, config) => {
			//success
			$log.info("register: Success.");
			window.location = LOC + '/dashboard';
		}, (err, status, headers, config) => {
			//error
			$log.error("register: Error: [" + JSON.stringify(err) + "] Status [" + status + "]");
		})
			.then(() => {
				indexScope.register.email = "";
				indexScope.register.pass = "";	
			});
		}
	}]);