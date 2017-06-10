app.controller('sysAdminController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var sysAdminScope = $scope;
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
				sysAdminScope.users = res.data.data;
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

		sysAdminScope.changeUserPriv = function(priv, pass) {
			$http(
			{
				method: 'GET',
				url: '/dashboard/sysadmin/changeuserpriv',
				data : {
					priv: priv,
					pass: pass,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.invites = res.data.data;
			}, (err) => {
				//error
				$log.error("changeUserPriv: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		//--init
		sysAdminScope.getUsers();
		//------
	}]);
