app.factory('CampaignService', function($resource) {
    return $resource('/api/campaigns/:id?api_key=' + user.api_key, {
        id: "@id"
    }, {
        update: {
            method: 'PUT'
        }
    });
});

app.factory('GroupService', function($resource) {
    return $resource('/api/groups/:id?api_key=' + user.api_key, {
        id: "@id"
    }, {
        update: {
            method: 'PUT'
        }
    });
});

app.factory('TemplateService', function($resource) {
    return $resource('/api/templates/:id?api_key=' + user.api_key, {
        id: "@id"
    }, {
        update: {
            method: 'PUT'
        }
    });
});