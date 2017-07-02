app.controller('adminControllerQuererStatus', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var adminAuditScope = $scope;

		adminAuditScope.conductAudit = function(e) {
			$http(
			{
				method: 'POST',
				url: '/dashboard/admin/conductAudit',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				adminAuditScope.auditData = res.data.data
			}, (err) => {
				//error
				$log.error("getUsers: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}


	}]);

app.controller('adminControllerLogs', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var adminLogs = $scope;

		adminLogs.getUserStats = function(e) {
			$http(
			{
				method: 'GET',
				url: '/dashboard/admin/getlogs',
				data: $.param({
					email: e,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				adminLogs.logFile = res.data.log;
			}, (err) => {
				//error
				$log.error("getUsers: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}
	}]);

app.controller('adminController', ['$scope', '$http', '$log', '$timeout',
	function($scope, $http, $log, $timeout) {
		var adminScope = $scope;
		var userTable;
		var inviteTable;

		adminScope.selectUser = function(i) {
			adminScope.selectedUser = angular.copy(adminScope.users[i]);
			adminScope.selectedUser.index = i;
			adminScope.getUserStats(adminScope.selectedUser.email)
		}

		adminScope.getUserStats = function(e) {
			$http(
			{
				method: 'POST',
				url: '/dashboard/admin/getuserstats',
				data: $.param({
					email: e,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				adminScope.userStats = res.data
			}, (err) => {
				//error
				$log.error("getUsers: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		adminScope.getUsers = function() {
			$http(
			{
				method: 'GET',
				url: '/dashboard/admin/getusers',
				data : {},
				withCredentials: true
			})
			.then((res) => {
				//success
				adminScope.users = res.data.data.users;
				adminScope.lev = res.data.data.lev;
				$timeout(() => {
					if (!$.fn.DataTable.isDataTable('#userTable')) {
						userTable = $('#userTable').DataTable({
							filter: true,
							select: 'single',
						});
					} else {
						userTable.rows().invalidate('data');
					}
				});
			}, (err) => {
				//error
				$log.error("getUsers: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}


		//--init
		adminScope.getUsers();
		adminScope.adminPass = "";
		adminScope.updateUserError = '';
		adminScope.userStats = null
		//------
	}]);