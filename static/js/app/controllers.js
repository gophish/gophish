var gophishApp = angular.module('gophishApp', []);

gophishApp.controller('CampaignCtrl', function($scope, $http) {
	$http.get('/api/campaigns?api_key=' + API_KEY).success(function(data) {
		$scope.campaigns = data;
	})
})