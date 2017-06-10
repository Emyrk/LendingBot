app.controller('sysAdminController', ['$scope', '$http', '$log', '$timeout',
	function($scope, $http, $log, $timeout) {
		var sysAdminScope = $scope;
		var userTable;

		sysAdminScope.selectUser = function(i) {
			sysAdminScope.selectedUser = angular.copy(sysAdminScope.users[i]);
			sysAdminScope.selectedUser.index = i;
		}

		sysAdminScope.getUsers = function() {
			$http(
			{
				method: 'GET',
				url: '/dashboard/sysadmin/getusers',
				data : {},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.users = res.data.data.users;
				sysAdminScope.lev = res.data.data.lev;
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

		sysAdminScope.deleteUser = function(email, pass) {
			$http(
			{
				method: 'GET',
				url: '/dashboard/sysadmin/deleteuser',
				data : {
					email: email,
					pass: pass,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.users = res.data.data;
			}, (err) => {
				//error
				$log.error("deleteUser: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		sysAdminScope.getInvites = function() {
			$http(
			{
				method: 'GET',
				url: '/dashboard/sysadmin/getinvites',
				data : {},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.invites = res.data.data;
			}, (err) => {
				//error
				$log.error("getInvites: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		sysAdminScope.makeInvite = function(hr, cap, code) {
			$http(
			{
				method: 'GET',
				url: '/dashboard/sysadmin/makeinvite',
				data : {
					hr: hours,
					cap: cap,
					code: code,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.invites = res.data.data;
			}, (err) => {
				//error
				$log.error("makeInvite: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		sysAdminScope.changeUserPriv = function() {
			sysAdminScope.updateUserError = '';
			$http(
			{
				method: 'POST',
				url: '/dashboard/sysadmin/changeuserpriv',
				data : $.param({
					email: sysAdminScope.selectedUser.email,
					priv: sysAdminScope.selectedUser.priv,
					pass: sysAdminScope.adminPass,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.users[sysAdminScope.selectedUser.index] = sysAdminScope.selectedUser;
				userTable.row(sysAdminScope.selectedUser.index).invalidate();
				sysAdminScope.selectedUser = null;
			}, (err) => {
				//error
				$log.error("changeUserPriv: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				sysAdminScope.updateUserError = err.data.error;
			})
			.then(() => {
				sysAdminScope.adminPass = "";
			});
		}

		//--init
		sysAdminScope.getUsers();
		sysAdminScope.adminPass = "";
		sysAdminScope.updateUserError = '';
		//------
	}]);
