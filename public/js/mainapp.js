var app = angular.module('lendingApp', ['ngMaterial']);
app.controller('indexController', ['$scope', '$http', function($scope, $http) {
	var indexScope = $scope;

	indexScope.login = function() {
		$http({
			email: indexScope.email
			pass: indexScope.pass
		})
		.then((res) => {
			if (res.error) {
				//error in rpc
				$log.error("jsonRpcService: postTorrentStreamSeek: Error: [" + JSON.stringify(res.error) + "]");
			} else {
				//success
				$log.info("jsonRpcService: postTorrentStreamSeek: Success.");
			}
		}, (res) => {
			//error on call SHOULD NEVER HAPPEN
			$log.error("jsonRpcService: Error: [" + JSON.stringify(err) + "]");
		});
	}
}]);