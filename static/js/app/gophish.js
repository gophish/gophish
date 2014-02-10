var app = angular.module('gophish', ['ngTable', 'ngResource']);

app.factory('CampaignService', function($resource) {
    return $resource('/api/campaigns/:id?api_key=' + API_KEY, {
        id: "@id"
    }, {
        update: {
            method: 'PUT'
        }
    });
});

app.factory('GroupService', function($resource) {
    return $resource('/api/groups/:id?api_key=' + API_KEY, {}, {
        update: {
            method: 'PUT'
        }
    });
});

app.controller('CampaignCtrl', function($scope, CampaignService) {
    CampaignService.query(function(campaigns) {
        $scope.campaigns = campaigns
    })
});

app.controller('GroupCtrl', function($scope, GroupService, ngTableParams) {


    $scope.tableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc'     // initial sorting
        }
    }, {
        total: 0, // length of data
        getData: function($defer, params) {
            GroupService.query(function(groups) {
                $scope.groups = groups
                params.total(groups.length)
                $defer.resolve(groups.slice((params.page() - 1) * params.count(), params.page() * params.count()));
            })
        }
    });

    $scope.editGroup = function(group) {
        if (group === 'new') {
            $scope.newGroup = true;
            $scope.group = {
                name: '',
                targets: [],
                id: 0
            };

        } else {
            $scope.newGroup = false;
            $scope.group = group;
        }
    };

    $scope.addTarget = function() {
        if ($scope.newTarget.email != "") {
            $scope.group.targets.push({
                email: $scope.newTarget.email
            });
            $scope.newTarget.email = ""
        }
    };
    $scope.removeTarget = function(target) {
        $scope.group.targets.splice($scope.group.targets.indexOf(target), 1);
    };
    $scope.saveGroup = function(group) {
        var newGroup = new GroupService($scope.group);
        if ($scope.newGroup) {
            newGroup.$save(function() {
                $scope.groups.push(newGroup);
            });
        } else {
            newGroup.$update()
        }
    }
})
