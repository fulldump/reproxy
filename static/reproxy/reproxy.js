angular

.module('Reproxy', [])

.controller('MainCtrl', function($scope, $http) {

	$scope.items = [];
	$scope.selected = null;

	$scope.reload = function() {
		$http({
			method: 'GET',
			url: 'config',
		}).then(
			function ok(response) {
				$scope.items = response.data;
			},
			function error(response) {

			}
		);
	};

	$scope.save = function(item) {
		$http({
			method: 'PUT',
			url: 'config/'+item.id,
			data: item,
		}).then(
			function ok(response) {
				console.log(response);
			},
			function error(response) {

			}
		);
	};

	$scope.select = function(item) {
		$scope.selected = item;
	};

	$scope.add_header = function() {
		if (null !== $scope.selected) {
			$scope.selected.headers.push({
				key: 'New-Key',
				value: 'new value',
			});
		}
	};

	$scope.input_keyup = function(e) {
		if(13 == e.keyCode) {
			var item = {prefix:$scope.input};
			$http({
				method: 'POST',
				url: 'config/',
				data: item,
			}).then(
				function ok(response) {
					$scope.items.push(response.data);
				},
				function error(response) {

				}
			);
			$scope.input = '';
		}
	};

	var that = this;

	$scope.reload();
})
