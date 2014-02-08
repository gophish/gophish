var app = angular.module('gophish', ['ngTable', 'ngResource']);

app.factory('CampaignService', function($resource) {
    return $resource('/api/campaigns/:id?api_key=' + API_KEY);
});

app.factory('GroupService', function($resource) {
    return $resource('/api/groups/:id?api_key=' + API_KEY);
});

app.controller('CampaignCtrl', function($scope, CampaignService) {
    CampaignService.query(function(campaigns){
        $scope.campaigns = campaigns
    })
});

app.controller('GroupCtrl', function($scope, GroupService) {
	GroupService.query(function(groups) {
		$scope.groups = groups
	})

    $scope.editGroup = function(group) {
        if (group === 'new') {
            $scope.newGroup = true;
            $scope.group = {
                name: '',
                targets: [{
                    email: ''
                }],
                id: 0
            };

        } else {
            $scope.newGroup = false;
            $scope.group = group;
        }
    };
})
