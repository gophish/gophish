var gophishApp = angular.module('gophishApp', []);

gophishApp.controller('CampaignCtrl', function($scope, $http) {
	$http.get('/api/campaigns?api_key=' + API_KEY).success(function(data) {
		$scope.campaigns = data;
	})
})

gophishApp.controller('GroupCtrl', function($scope, $http) {
	$http.get('/api/groups?api_key=' + API_KEY).success(function(data) {
		$scope.groups = data;
	})
})