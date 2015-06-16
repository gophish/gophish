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

var groups = {
    // get() - Queries the API for GET /campaigns
    get: function(){
        return api("/groups", "GET", {})
    },
    // post() - Posts a campaign to POST /campaigns
    post: function(group){
        return api("/groups", "POST", group)
    }
}

var groupId = {
    // get() - Queries the API for GET /groups/:id
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
