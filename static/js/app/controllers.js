app.controller('DashboardCtrl', function($scope, $filter, $location, CampaignService, ngTableParams, $http) {
    $scope.campaigns = []
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
                var campaign_series = [];
                var avg = 0;
                angular.copy(campaigns, campaign_series)
                angular.forEach(campaigns, function(campaign, key) {
                    campaign.x = new Date(campaign.created_date)
                    campaign.y = 0
                    angular.forEach(campaign.results, function(result, r_key) {
                        if (result.status == "Success") {
                            campaign.y++;
                        }
                    })
                    campaign.y = Math.floor((campaign.y / campaign.results.length) * 100)
                    avg += campaign.y
                });
                avg = Math.floor(avg / campaigns.length);
                $scope.overview_chart = {
                    options: {
                        chart: {
                            type: 'area',
                            zoomType: "x"
                        },
                        tooltip: {
                            formatter: function() {
                                return "Name: " + this.point.name + "<br/>Successful Phishes: " + this.point.y + "%<br/>Date: " + $filter("date")(this.point.x, "medium")
                            },
                            style: {
                                padding: 10,
                                fontWeight: 'bold'
                            }
                        },
                        plotOptions: {
                            series: {
                                cursor: 'pointer',
                                point: {
                                    events: {
                                        click: function(e) {
                                            $location.path("/campaigns/" + this.id)
                                            $scope.$apply()
                                        }
                                    }
                                }
                            }
                        },
                        xAxis: {
                            type: 'datetime',
                            max: Date.now(),
                            title: {
                                text: 'Date'
                            }
                        },
                    },
                    series: [{
                        name: "Campaigns",
                        data: $scope.campaigns
                    }],
                    title: {
                        text: 'Phishing Success Overview'
                    },
                    size: {
                        height: 300
                    },
                    credits: {
                        enabled: false
                    },
                    loading: false,
                }
                $scope.average_chart = {
                    options: {
                        chart: {
                            type: 'pie'
                        },
                        tooltip: {
                            formatter: function() {
                                return this.point.y + "%"
                            },
                            style: {
                                padding: 10,
                                fontWeight: 'bold'
                            }
                        },
                        plotOptions: {
                            pie: {
                                innerSize: '60%',
                                allowPointSelect: true,
                                cursor: 'pointer',
                                dataLabels: {
                                    enabled: false
                                },
                                showInLegend: true
                            }
                        },
                    },
                    series: [{
                        data: [{
                            name: "Successful Phishes",
                            color: "#e74c3c",
                            y: avg
                        }, {
                            name: "Unsuccessful Phishes",
                            color: "#7cb5ec",
                            y: 100 - avg
                        }]
                    }],
                    title: {
                        text: 'Average Phishing Results'
                    },
                    size: {
                        height: 300
                    },
                    credits: {
                        enabled: false
                    },
                    loading: false,
                }
                params.total(Math.min(campaigns.length, 5));
                $defer.resolve(campaigns.slice(0, params.total()));
            })
        }
    });
})
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
            $scope.successFlash("Campaign deleted successfully")
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
    $scope.ok = function(campaign) {
        $modalInstance.dismiss("")
        $scope.saveCampaign(campaign)
    }
};

app.controller('CampaignResultsCtrl', function($scope, $filter, CampaignService, GroupService, ngTableParams, $http, $window) {
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
                var result_series = []
                angular.forEach(campaign.results, function(result, key) {
                    var new_entry = true;
                    for (var i = 0; i < result_series.length; i++) {
                        if (result_series[i].name == result.status) {
                            result_series[i].y++;
                            new_entry = false;
                            break;
                        }
                    }
                    if (new_entry) {
                        result_series.push({
                            name: result.status,
                            y: 1
                        })
                    }
                });
                angular.forEach(campaign.timeline, function(e, key) {
                    e.x = new Date(e.time);
                    e.y = 0;
                });
                $scope.email_chart = {
                    options: {
                        chart: {
                            type: 'pie'
                        },
                        tooltip: {
                            formatter: function() {
                                return this.point.name + " : " + this.point.y
                            },
                            style: {
                                padding: 10,
                                fontWeight: 'bold'
                            }
                        },
                        plotOptions: {
                            pie: {
                                allowPointSelect: true,
                                cursor: 'pointer',
                                dataLabels: {
                                    enabled: false
                                },
                                showInLegend: true
                            }
                        }
                    },
                    series: [{
                        data: result_series
                    }],
                    title: {
                        text: 'Email Status'
                    },
                    size: {
                        height: 300
                    },
                    credits: {
                        enabled: false
                    },
                    loading: false,
                }
                $scope.timeline_chart = {
                    options: {
                        global: {
                            useUTC: false
                        },
                        chart: {
                            type: 'scatter',
                            zoomType: "x"
                        },
                        tooltip: {
                            formatter: function() {
                                var label = "Event: " + this.point.message + "<br/>";
                                if (this.point.email) {
                                    label += "Email: " + this.point.email + "<br/>";
                                }
                                label += "Date: " + $filter("date")(this.point.x, "medium");
                                return label
                            },
                            style: {
                                padding: 10,
                                fontWeight: 'bold'
                            }
                        },
                        plotOptions: {
                            series: {
                                cursor: 'pointer',
                            }
                        },
                        yAxis: {
                            labels: {
                                enabled: false
                            },
                            title: {
                                text: "Events"
                            }
                        },
                        xAxis: {
                            type: 'datetime',
                            dateTimeLabelFormats: { // don't display the dummy year
                                day: "%e of %b",
                                hour: "%l:%M",
                                second: '%l:%M:%S',
                                minute: '%l:%M'
                            },
                            max: Date.now(),
                            title: {
                                text: 'Date'
                            }
                        },
                    },
                    series: [{
                        name: "Events",
                        data: $scope.campaign.timeline
                    }],
                    title: {
                        text: 'Campaign Timeline'
                    },
                    size: {
                        height: 300
                    },
                    credits: {
                        enabled: false
                    },
                    loading: false,
                }
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

var GroupModalCtrl = function($scope, $modalInstance, $upload) {
    $scope.onFileSelect = function($file) {
        $scope.upload = $upload.upload({
            url: '/api/import/group',
            data: {},
            file: $file,
        }).progress(function(evt) {
            console.log('percent: ' + parseInt(100.0 * evt.loaded / evt.total));
        }).success(function(data, status, headers, config) {
            angular.forEach(data, function(record, key) {
                $scope.group.targets.push({
                    email: record.email
                });
            });
            $scope.editGroupTableParams.reload();
            //.error(...)
        });
    };
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
    $scope.ok = function(group) {
        $modalInstance.dismiss('')
        $scope.saveGroup(group)
    };
}

app.controller('TemplateCtrl', function($scope, $modal, TemplateService, ngTableParams) {
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
                attachments: []
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
        }, function(response) {
            if (response.success) {
                $scope.successFlash(response.message)
            } else {
                $scope.errorFlash(response.message)
            }
            $scope.mainTableParams.reload();
        });
    }
})

var TemplateModalCtrl = function($scope, $upload, $modalInstance) {
    $scope.onFileSelect = function($files) {
        console.log($files)
        angular.forEach($files, function(file, key) {
            var reader = new FileReader();
            reader.onload = function(e) {
                $scope.template.attachments.push({
                        name : file.name,
                        content : reader.result.split(",")[1],
                        type : file.type || "application/octet-stream"
                })
                $scope.$apply();
            }
            reader.onerror = function(e) {
                console.log(e)
            }
            reader.readAsDataURL(file)
        })
    }
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
    $scope.ok = function(template) {
        $modalInstance.dismiss('')
        $scope.saveTemplate(template)
    };
    $scope.removeFile = function(file) {
        $scope.template.attachments.splice($scope.template.attachments.indexOf(file), 1);
    }
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
            .success(function(response) {
                if (response.success) {
                    $scope.user.api_key = response.data;
                    $window.user.api_key = response.data;
                    $scope.successFlash(response.message)
                }
            })
    }
    $scope.save_settings = function() {
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
                } else {
                    $scope.errorFlash(data.message)
                }
            })
    }
})
