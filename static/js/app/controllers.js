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
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.main.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};;
        $scope.flashes.main.push({
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

        modalInstance.result.then(function(message) {
            $scope.successFlash(message)
            $scope.campaign = {
                name: '',
                groups: [],
            };
        }, function() {
            $scope.campaign = {
                name: '',
                groups: [],
            };
        });
        $scope.mainTableParams.reload()
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

    $scope.deleteCampaign = function(campaign) {
        var deleteCampaign = new CampaignService(campaign);
        deleteCampaign.$delete({
            id: deleteCampaign.id
        }, function(response) {
            if (response.success) {
                $scope.successFlash(response.message)
            } else {
                $scope.errorFlash(response.message)
            }
            $scope.mainTableParams.reload();
        });
    }
});

var CampaignModalCtrl = function($scope, CampaignService, $modalInstance) {
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.modal.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }
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
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
    $scope.ok = function(campaign) {
        var newCampaign = new CampaignService(campaign);
        newCampaign.$save({}, function() {
            $modalInstance.close("Campaign added successfully")
            $scope.campaigns.push(newCampaign);
            $scope.mainTableParams.reload()
        }, function(response) {
            $scope.errorFlash(response.data.message)
        });
    }
};

app.controller('CampaignResultsCtrl', function($scope, $filter, CampaignService, GroupService, ngTableParams, $http, $window) {
    id = $window.location.hash.split('/')[2];
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.main.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};;
        $scope.flashes.main.push({
            "type": "success",
            "message": message,
            "icon": "fa-check-circle"
        })
    }

    $scope.delete = function(campaign) {
        if (confirm("Delete campaign?")) {
            var deleteCampaign = new CampaignService(campaign);
            deleteCampaign.$delete({
                id: deleteCampaign.id
            }, function(response) {
                if (response.success) {
                    $scope.successFlash(response.message)
                } else {
                    $scope.errorFlash(response.message)
                }
                $scope.mainTableParams.reload();
            });
        }
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
})

app.controller('GroupCtrl', function($scope, $modal, GroupService, ngTableParams) {
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.main.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};;
        $scope.flashes.main.push({
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
        modalInstance.result.then(function(message) {
            $scope.successFlash(message)
            $scope.group = {
                name: '',
                targets: [],
            };
        }, function() {
            $scope.group = {
                name: '',
                targets: [],
            };
        });
        $scope.mainTableParams.reload()

    };

    $scope.deleteGroup = function(group) {
        var deleteGroup = new GroupService(group);
        deleteGroup.$delete({
            id: deleteGroup.id
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

var GroupModalCtrl = function($scope, GroupService, $modalInstance, $upload) {
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.modal.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }
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
                    first_name : record.first_name,
                    last_name : record.last_name,
                    email: record.email,
                    position: record.position
                });
            });
            $scope.editGroupTableParams.reload();
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
    $scope.cancel = function() {
        $modalInstance.dismiss();
    };
    $scope.ok = function(group) {
        var newGroup = new GroupService(group);
        if ($scope.newGroup) {
            newGroup.$save({}, function() {
                $scope.groups.push(newGroup);
                $modalInstance.close("Group created successfully!")
            }, function(error){
                $scope.errorFlash(error.data.message)
            });
        } else {
            newGroup.$update({
                id: newGroup.id
            },function(){
                $modalInstance.close("Group updated successfully!")
            }, function(error){
                $scope.errorFlash(error.data.message)
            })
        }
    };
}

app.controller('TemplateCtrl', function($scope, $modal, TemplateService, ngTableParams) {
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.main.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};;
        $scope.flashes.main.push({
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

        modalInstance.result.then(function(message) {
            $scope.successFlash(message)
            $scope.template = {
                name: '',
                html: '',
                text: '',
            };
        }, function() {
            $scope.template = {
                name: '',
                html: '',
                text: '',
            };
        });
    };

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

var TemplateModalCtrl = function($scope, TemplateService, $upload, $modalInstance, $modal) {
    $scope.editorOptions = {
        fullPage: true,
        allowedContent: true,
    }
    $scope.errorFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};
        $scope.flashes.modal.push({
            "type": "danger",
            "message": message,
            "icon": "fa-exclamation-circle"
        })
    }

    $scope.successFlash = function(message) {
        $scope.flashes = {"main" : [], "modal" : []};;
        $scope.flashes.modal.push({
            "type": "success",
            "message": message,
            "icon": "fa-check-circle"
        })
    }
    $scope.onFileSelect = function($files) {
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
        $modalInstance.dismiss();
    };
    $scope.ok = function(template) {
        var newTemplate = new TemplateService(template);
        // If it's a new template
        if ($scope.newTemplate) {
            // POST the template to /api/templates
            newTemplate.$save({}, function() {
                // If successful, push the template into the list
                $scope.templates.push(newTemplate);
                $scope.mainTableParams.reload()
                // Close the dialog, returning the template
                $modalInstance.close("Template created successfully!")
            }, function(error){
                // Otherwise, leave the dialog open, showing the error
                $scope.errorFlash(error.data.message)
            });
        } else {
            newTemplate.$update({
                id: newTemplate.id
            }, function(){
                $modalInstance.close("Template updated successfully!")
            }, function(error){
                $scope.errorFlash(error.data.message)
            })
        }
    };
    $scope.removeFile = function(file) {
        $scope.template.attachments.splice($scope.template.attachments.indexOf(file), 1);
    }

    $scope.importEmail = function() {
        var emailInstance = $modal.open({
            templateUrl: '/js/app/partials/modals/importEmailModal.html',
            controller: ImportEmailCtrl,
            scope: $scope
        });

        emailInstance.result.then(function(raw) {
            $scope.template.subject = raw;
        }, function() {});
    };
};

var ImportEmailCtrl = function($scope, $http, $modalInstance) {
    $scope.email = {}
    $scope.ok = function() {
        // Simple POST request example (passing data) :
        $http.post('/api/import/email', $scope.email.raw,
        { headers : {"Content-Type" : "text/plain"}}
    ).success(function(data) {console.log("Success: " + data)})
    .error(function(data) {console.log("Error: " + data)});
    $modalInstance.close($scope.email.raw)
};
$scope.cancel = function() {$modalInstance.dismiss()}
};

app.controller('LandingPageCtrl', function($scope, $modal, LandingPageService, ngTableParams) {
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
            LandingPageService.query(function(pages) {
                $scope.pages = pages
                params.total(pages.length)
                $defer.resolve(pages.slice((params.page() - 1) * params.count(), params.page() * params.count()));
            })
        }
    });

    $scope.editPage = function(page) {
        if (page === 'new') {
            $scope.newPage = true;
            $scope.page = {
                name: '',
                html: '',
            };

        } else {
            $scope.newPage = false;
            $scope.page = page;
        }
        var modalInstance = $modal.open({
            templateUrl: '/js/app/partials/modals/LandingPageModal.html',
            controller: LandingPageModalCtrl,
            scope: $scope
        });

        modalInstance.result.then(function(selectedItem) {
            $scope.selected = selectedItem;
        }, function() {
            console.log('closed')
        });
    };

    $scope.savePage = function(page) {
        var newPage = new LandingPageService(page);
        if ($scope.newPage) {
            newPage.$save({}, function() {
                $scope.pages.push(newPage);
                $scope.mainTableParams.reload()
            });
        } else {
            newPage.$update({
                id: newPage.id
            })
        }
        $scope.page = {
            name: '',
            html: '',
        };
    }
    $scope.deletePage = function(page) {
        var deletePage = new LandingPageService(page);
        deletePage.$delete({
            id: deletePage.id
        }, function(response) {
            if (response.success) {
                $scope.successFlash(response.message)
            } else {
                $scope.errorFlash(response.message)
            }
            $scope.mainTableParams.reload();
        });
    }
});

var LandingPageModalCtrl = function($scope, $modalInstance) {
    $scope.editorOptions = {
        fullPage: true,
        allowedContent: true,
        startupMode: "source"
    }
    $scope.cancel = function() {
        $modalInstance.dismiss('cancel');
    };
    $scope.ok = function(page) {
        $modalInstance.dismiss('')
        $scope.savePage(page)
    };
    $scope.csrf_token = csrf_token
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
