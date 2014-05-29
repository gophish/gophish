app.controller('CampaignCtrl', function($scope, $modal, CampaignService, GroupService, TemplateService, ngTableParams, $http) {
    $scope.flashes = []
    $scope.mainTableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc' // initial sorting
        }
    }, {
        total: 0, // length of data
        getData: function($defer, params) {
            CampaignService.query(function(campaigns) {
                $scope.campaigns = campaigns
                params.total(campaigns.length)
                $defer.resolve(campaigns.slice((params.page() - 1) * params.count(), params.page() * params.count()));
            })
        }
    });

    GroupService.query(function(groups) {
        $scope.groups = groups;
    })

    TemplateService.query(function(templates) {
        $scope.templates = templates;
    })

    $scope.addGroup = function(group) {
        if (group.name != "") {
            $scope.campaign.groups.push({
                name: group.name
            });
            group.name = ""
            $scope.editGroupTableParams.reload()
        }
    };

    $scope.removeGroup = function(group) {
        $scope.campaign.groups.splice($scope.campaign.groups.indexOf(group), 1);
        $scope.editGroupTableParams.reload()
    };

    $scope.newCampaign = function() {
        $scope.campaign = {
            name: '',
            groups: []
        };
        $scope.editCampaign($scope.campaign)
    };

    $scope.editCampaign = function(campaign) {
        var modalInstance = $modal.open({
            templateUrl: '/js/app/partials/modals/campaignModal.html',
            controller: CampaignModalCtrl,
            scope: $scope
        });

        modalInstance.result.then(function(selectedItem) {
            $scope.selected = selectedItem;
        }, function() {
            console.log('closed')
        });
    };

    $scope.editGroupTableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc' // initial sorting
        }
    }, {
        total: 0, // length of data
        getData: function($defer, params) {
            params.total($scope.campaign.groups.length)
            $defer.resolve($scope.campaign.groups.slice((params.page() - 1) * params.count(), params.page() * params.count()));
        }
    });

    $scope.saveCampaign = function(campaign) {
        $scope.flashes = []
        $scope.validated = true
        var newCampaign = new CampaignService(campaign);
        newCampaign.$save({}, function() {
            $scope.successFlash("Campaign added successfully")
            $scope.campaigns.push(newCampaign);
            $scope.mainTableParams.reload()
        }, function(response) {
            $scope.errorFlash(response.data)
        });
        $scope.campaign = {
            groups: [],
        };
        $scope.editGroupTableParams.reload()
    }

    $scope.deleteCampaign = function(campaign) {
        var deleteCampaign = new CampaignService(campaign);
        deleteCampaign.$delete({
            id: deleteCampaign.id
        }, function() {
            $scope.mainTableParams.reload();
        });
    }

    $scope.errorFlash = function(message) {
        $scope.flashes.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes.push({
            "type": "success",
            "message": message,
            "icon": "fa-check-circle"
        })
    }
});

var CampaignModalCtrl = function($scope, $modalInstance) {
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
};

app.controller('CampaignResultsCtrl', function($scope, CampaignService, GroupService, ngTableParams, $http, $window) {
    id = $window.location.hash.split('/')[2];
    $scope.flashes = []
    $scope.mainTableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc' // initial sorting
        }
    }, {
        total: 0, // length of data
        getData: function($defer, params) {
            CampaignService.get({
                "id": id
            }, function(campaign) {
                $scope.campaign = campaign
                params.total(campaign.results.length)
                $defer.resolve(campaign.results.slice((params.page() - 1) * params.count(), params.page() * params.count()));
            })
        }
    });

    $scope.errorFlash = function(message) {
        $scope.flashes.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }
});

app.controller('GroupCtrl', function($scope, $modal, GroupService, ngTableParams) {
    $scope.mainTableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc' // initial sorting
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

    $scope.editGroupTableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc' // initial sorting
        }
    }, {
        total: 0, // length of data
        getData: function($defer, params) {
            params.total($scope.group.targets.length)
            $defer.resolve($scope.group.targets.slice((params.page() - 1) * params.count(), params.page() * params.count()));
        }
    });

    $scope.editGroup = function(group) {
        if (group === 'new') {
            $scope.newGroup = true;
            $scope.group = {
                name: '',
                targets: [],
            };

        } else {
            $scope.newGroup = false;
            $scope.group = group;
            $scope.editGroupTableParams.reload()
        }
        $scope.newTarget = {};
        var modalInstance = $modal.open({
            templateUrl: '/js/app/partials/modals/userModal.html',
            controller: GroupModalCtrl,
            scope: $scope
        });
    };

    $scope.addTarget = function() {
        if ($scope.newTarget.email != "") {
            $scope.group.targets.push({
                email: $scope.newTarget.email
            });
            $scope.newTarget.email = ""
            $scope.editGroupTableParams.reload()
        }
    };
    $scope.removeTarget = function(target) {
        $scope.group.targets.splice($scope.group.targets.indexOf(target), 1);
        $scope.editGroupTableParams.reload()
    };
    $scope.saveGroup = function(group) {
        var newGroup = new GroupService(group);
        if ($scope.newGroup) {
            newGroup.$save({}, function() {
                $scope.groups.push(newGroup);
                $scope.mainTableParams.reload()
            });
        } else {
            newGroup.$update({
                id: newGroup.id
            })
        }
        $scope.group = {
            name: '',
            targets: [],
        };
        $scope.editGroupTableParams.reload()
    }
    $scope.deleteGroup = function(group) {
        var deleteGroup = new GroupService(group);
        deleteGroup.$delete({
            id: deleteGroup.id
        }, function() {
            $scope.mainTableParams.reload();
        });
    }
})

var GroupModalCtrl = function($scope, $modalInstance) {
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
}

app.controller('TemplateCtrl', function($scope, $modal, TemplateService, ngTableParams) {
    $scope.mainTableParams = new ngTableParams({
        page: 1, // show first page
        count: 10, // count per page
        sorting: {
            name: 'asc' // initial sorting
        }
    }, {
        total: 0, // length of data
        getData: function($defer, params) {
            TemplateService.query(function(templates) {
                $scope.templates = templates
                params.total(templates.length)
                $defer.resolve(templates.slice((params.page() - 1) * params.count(), params.page() * params.count()));
            })
        }
    });

    $scope.editTemplate = function(template) {
        if (template === 'new') {
            $scope.newTemplate = true;
            $scope.template = {
                name: '',
                html: '',
                text: '',
            };

        } else {
            $scope.newTemplate = false;
            $scope.template = template;
        }
        var modalInstance = $modal.open({
            templateUrl: '/js/app/partials/modals/templateModal.html',
            controller: TemplateModalCtrl,
            scope: $scope
        });

        modalInstance.result.then(function(selectedItem) {
            $scope.selected = selectedItem;
        }, function() {
            console.log('closed')
        });
    };

    $scope.saveTemplate = function(template) {
        var newTemplate = new TemplateService(template);
        if ($scope.newTemplate) {
            newTemplate.$save({}, function() {
                $scope.templates.push(newTemplate);
                $scope.mainTableParams.reload()
            });
        } else {
            newTemplate.$update({
                id: newTemplate.id
            })
        }
        $scope.template = {
            name: '',
            html: '',
            text: '',
        };
    }
    $scope.deleteTemplate = function(template) {
        var deleteTemplate = new TemplateService(template);
        deleteTemplate.$delete({
            id: deleteTemplate.id
        }, function() {
            $scope.mainTableParams.reload();
        });
    }
})

var TemplateModalCtrl = function($scope, $modalInstance) {
    console.log($scope.template)
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
};

app.controller('SettingsCtrl', function($scope, $http, $window) {
    $scope.flashes = [];
    $scope.user = user;
    $scope.errorFlash = function(message) {
        $scope.flashes = [];
        $scope.flashes.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes = [];
        $scope.flashes.push({
            "type": "success",
            "message": message,
            "icon": "fa-check-circle"
        })
    }
    $scope.form_data = {
        username: user.username,
        csrf_token: csrf_token
    }
    $scope.api_reset = function() {
        $http({
            method: 'POST',
            url: '/api/reset',
            data: $.param($scope.form_data), // pass in data as strings
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            } // set the headers so angular passing info as form data (not request payload)
        })
            .success(function(data) {
                $scope.user.api_key = data;
                $window.user.api_key = data;
                $scope.successFlash("API Key Successfully Reset")
            })
    }
    $scope.save_settings = function(){
        $http({
            method: 'POST',
            url: '/settings',
            data: $.param($scope.form_data),
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            }
        })
            .success(function(data) {
                if (data.success) {
                    $scope.successFlash(data.message)
                }
                else {
                    $scope.errorFlash(data.message)
                }
            })
    }
})
