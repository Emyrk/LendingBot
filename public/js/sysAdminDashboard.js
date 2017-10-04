app.controller('sysAdminController', ['$scope', '$http', '$log', '$timeout',
	function($scope, $http, $log, $timeout) {
		var sysAdminScope = $scope;
		var userTable;
		var inviteTable;

		sysAdminScope.selectUser = function(i) {
			sysAdminScope.selectedUser = angular.copy(sysAdminScope.users[i]);
			sysAdminScope.selectedUser.index = i;
			sysAdminScope.getStatus(sysAdminScope.selectedUser.email);
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
				console.log("Retrieved invites :" + res.data.data)
				$timeout(() => {
					if (!$.fn.DataTable.isDataTable('#inviteTable')) {
						inviteTable = $('#inviteTable').DataTable({
							filter: true,
						});
					} else {
						inviteTable.rows().invalidate('data');
					}
				});
			}, (err) => {
				//error
				$log.error("getInvites: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		sysAdminScope.createInvite = function() {
			sysAdminScope.makeInviteError = '';
			$http(
			{
				method: 'POST',
				url: '/dashboard/sysadmin/makeinvite',
				data : $.param({
					rawc: sysAdminScope.makeInvite.rawc,
					cap: sysAdminScope.makeInvite.cap,
					hr: sysAdminScope.makeInvite.hr,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				sysAdminScope.invites = res.data.data;
				sysAdminScope.makeInvite = {};
				sysAdminScope.getInvites();
			}, (err) => {
				//error
				$log.error("makeInvite: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				sysAdminScope.makeInviteError = err.data.error;
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
				userTable.row(sysAdminScope.selectedUser.index).invalidate();
				sysAdminScope.users[sysAdminScope.selectedUser.index] = sysAdminScope.selectedUser;
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

		sysAdminScope.deleteInvite = function(rawc) {
			$http(
			{
				method: 'POST',
				url: '/dashboard/sysadmin/deleteinvite',
				data : $.param({
					rawc: rawc
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				console.log("DeletedInvite");
				sysAdminScope.getInvites();
			}, (err) => {
				//error
				$log.error("changeUserPriv: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				sysAdminScope.updateUserError = err.data.error;
			})
			.then(() => {
				sysAdminScope.adminPass = "";
			});
		}

		sysAdminScope.getStatus = function(username) {
			$http(
			{
				method: 'GET',
				url: '/dashboard/sysadmin/getuserstatus',
				params: {
					email: username,
				},
				withCredentials: true
			})
			.then((res) => {
				//success
				console.log("getStatus: Success");
				sysAdminScope.selectUser.status = JSON.stringify(res.data.status, null, 2);
			}, (err) => {
				//error
				$log.error("getUsers: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
			});
		}

		sysAdminScope.addCustomChargeReduction = function() {
			sysAdminScope.addReductionError = '';
			$http(
			{
				method: 'POST',
				url: '/dashboard/sysadmin/addcustomreduc',
				data : $.param({
					email: sysAdminScope.selectedUser.email,
					percAmount: sysAdminScope.reduc.percAmount,
					reason: sysAdminScope.reduc.reason,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				console.log("addCustomChargeReduction: Success");
				sysAdminScope.selectUser.status = JSON.stringify(res.data.status, null, 2);
				sysAdminScope.reduc.percAmount = 0.0;
				sysAdminScope.reduc.reason = "";
			}, (err) => {
				//error
				$log.error("addCustomChargeReduction: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				sysAdminScope.addReductionError = err.data.error;
			});
		}

		sysAdminScope.addPayment = function() {
			sysAdminScope.addPaymentError = '';
			$http(
			{
				method: 'POST',
				url: '/dashboard/sysadmin/addpayment',
				data : $.param({
					email: sysAdminScope.selectedUser.email,
					amount: sysAdminScope.payment.amount,
					reason: sysAdminScope.payment.reason,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				console.log("addPayment: Success");
				sysAdminScope.payment.status = JSON.stringify(res.data.status, null, 2);
				sysAdminScope.payment.amount = 0;
				sysAdminScope.payment.reason = "";
			}, (err) => {
				//error
				$log.error("addCustomChargeReduction: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				sysAdminScope.addPaymentError = err.data.error;
			});
		}

		//--init
		sysAdminScope.getUsers();
		sysAdminScope.adminPass = "";
		sysAdminScope.updateUserError = '';
		sysAdminScope.makeInviteError = '';
		sysAdminScope.addReductionError = '';
		sysAdminScope.addPaymentError = '';
		sysAdminScope.makeInvite = {};
		sysAdminScope.getInvites();
		//------
	}]);
