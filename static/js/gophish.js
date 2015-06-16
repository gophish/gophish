function errorFlash(message) {
    $("#flashes").append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
        <i class=\"fa fa-exclamation-circle\"></i>" + message + "</div>"
    )
}

function successFlash(message) {
    $("#flashes").append("<div style=\"text-align:center\" class=\"alert alert-success\">\
        <i class=\"fa fa-check-circle\"></i> " + message + "</div>"
    )
}

function api(endpoint, method, data) {
    return $.ajax({
        url: "/api" + endpoint + "?api_key=" + user.api_key,
        async: false,
        method: method,
        data: data,
        dataType:"json"
    })
}

/*
Define our API Endpoints
*/
// campaigns contains the endpoints for /campaigns
var campaigns = {
    // get() - Queries the API for GET /campaigns
    get: function(){
        return api("/campaigns", "GET", {})
    },
    // post() - Posts a campaign to POST /campaigns
    post: function(data){
        return api("/campaigns", "POST", data)
    }
}

// campaignId contains the endpoints for /campaigns/:id
var campaignId = {
    // get() - Queries the API for GET /campaigns/:id
    get: function(id){
        return api("/campaigns/" + id, "GET", {})
    },
    // post() - Posts a campaign to POST /campaigns/:id
    post: function(campaign){
        return api("/campaigns/" + campaign.id, "POST", data)
    },
    // put() - Puts a campaign to PUT /campaigns/:id
    put: function (campaign){
        return api("/campaigns/" + campaign.id, "PUT", data)
    },
    // delete() - Deletes a campaign at DELETE /campaigns/:id
    delete: function(id){
        return api("/campaigns/" + id, "DELETE", data)
    }
}

// groups contains the endpoints for /groups
var groups = {
    // get() - Queries the API for GET /groups
    get: function(){
        return api("/groups", "GET", {})
    },
    // post() - Posts a campaign to POST /groups
    post: function(group){
        return api("/groups", "POST", group)
    }
}

// groupId contains the endpoints for /groups/:id
var groupId = {
    // get() - Queries the API for GET /groups/:id
    get: function(id){
        return api("/groups/" + id, "GET", {})
    },
    // post() - Posts a campaign to POST /groups/:id
    post: function(group){
        return api("/groups/" + group.id, "POST", data)
    },
    // put() - Puts a campaign to PUT /groups/:id
    put: function (group){
        return api("/groups/" + group.id, "PUT", data)
    },
    // delete() - Deletes a campaign at DELETE /groups/:id
    delete: function(id){
        return api("/groups/" + id, "DELETE", data)
    }
}

// templates contains the endpoints for /templates
var templates = {
    // get() - Queries the API for GET /templates
    get: function(){
        return api("/templates", "GET", {})
    },
    // post() - Posts a campaign to POST /templates
    post: function(template){
        return api("/templates", "POST", template)
    }
}

// templateId contains the endpoints for /templates/:id
var templateId = {
    // get() - Queries the API for GET /templates/:id
    get: function(id){
        return api("/templates/" + id, "GET", {})
    },
    // post() - Posts a campaign to POST /templates/:id
    post: function(template){
        return api("/templates/" + template.id, "POST", data)
    },
    // put() - Puts a campaign to PUT /templates/:id
    put: function (template){
        return api("/templates/" + template.id, "PUT", data)
    },
    // delete() - Deletes a campaign at DELETE /templates/:id
    delete: function(id){
        return api("/templates/" + id, "DELETE", data)
    }
}
