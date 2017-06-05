var app = angular.module('lendingApp', ['ngMaterial']);

app.controller('indexController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var indexScope = $scope;

		indexScope.login = function() {
			indexScope.failedLogin = false;
			indexScope.attemptingLogin = true;
			$http({
				method: 'POST',
				url: '/login',
				data: $.param({
					email: indexScope.login.email,
					pass: indexScope.login.pass,
					twofa: indexScope.login.twofa,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res, status, headers, config) => {
				//success
				$log.info("login: Success.");
				window.location = LOC + '/dashboard';
			}, (err, status, headers, config) => {
				//error
				$log.error("login: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				indexScope.failedLogin = true;
			})
			.then(() => {
				indexScope.attemptingLogin = false;
				indexScope.login.email = "";
				indexScope.login.pass = "";
				indexScope.login.twofa = "";
			});
		}

		indexScope.cancelLogin = function() {
			indexScope.attemptingLogin = false;
			indexScope.failedLogin = false;
			indexScope.login.email = "";
			indexScope.login.pass = "";
			indexScope.login.twofa = "";
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
			.then((res) => {
				//success
				$log.info("register: Success.");
				window.location = LOC + '/dashboard';
			}, (err) => {
				//error
				$log.error("register: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			})
			.then(() => {
				indexScope.register.email = "";
				indexScope.register.pass = "";	
			});
		}

		indexScope.cancelRegister = function() {
			indexScope.attemptingLogin = false;
			indexScope.failedLogin = false;
			indexScope.login.email = "";
			indexScope.login.pass = "";
			indexScope.login.twofa = "";
		}

		//--init
		indexScope.attemptingLogin = false;
		indexScope.failedLogin = false;
		//
	}]);