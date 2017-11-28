angular

.module('Reproxy', [])

.controller('MainCtrl', function($scope, $http) {

	$scope.items = [];
	$scope.selected = null;
	$scope.response = null;
	$scope.types = [
		{name: 'custom', description: 'Custom static response',},
		{name: 'statics', description: 'Serve static files',},
        {name: 'proxy', description: 'Proxy to an http server',},
        {name: 'script', description: 'Run script',},
	];

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

	$scope.save_selected = function() {
		var item = $scope.selected;
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

	$scope.try_selected = function() {
		var item = $scope.selected;

		var on_receive = function(response) {
			$scope.response = response;
			console.log(response);
			console.log(response.headers());
		};

		$http.get(item.prefix).then(on_receive, on_receive);
	};

	$scope.select = function(item) {
		$scope.selected = item;
		$scope.response = null;
	};

	$scope.remove_selected = function() {
		var item = $scope.selected;
		if (confirm('You are going to remove this rule \''+$scope.selected.prefix+'\' are you sure?')) {
			$http({
				method: 'DELETE',
				url: 'config/'+$scope.selected.id,
			}).then(function(response) {
				$scope.selected = null;
				$scope.reload();
			});
		}
	};

	$scope.add_header = function(headers) {
		headers.push({
			key: 'New-Key',
			value: 'new value',
		});
	};

	$scope.remove_header = function(headers, id) {
		headers.splice(id, 1);
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
