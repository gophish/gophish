function errorFlash(message) {
    $("#flashes").empty()
    $("#flashes").append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
        <i class=\"fa fa-exclamation-circle\"></i> " + message + "</div>"
    )
}

function successFlash(message) {
    $("#flashes").empty()
    $("#flashes").append("<div style=\"text-align:center\" class=\"alert alert-success\">\
        <i class=\"fa fa-check-circle\"></i> " + message + "</div>"
    )
}

function modalError(message){
    $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
        <i class=\"fa fa-exclamation-circle\"></i> " + message + "</div>")
}

function query(endpoint, method, data) {
    return $.ajax({
        url: "/api" + endpoint + "?api_key=" + user.api_key,
        async: false,
        method: method,
        data: JSON.stringify(data),
        dataType:"json",
        contentType: "application/json"
    })
}

/*
Define our API Endpoints
*/
var api = {
    // campaigns contains the endpoints for /campaigns
    campaigns : {
        // get() - Queries the API for GET /campaigns
        get: function(){
            return query("/campaigns/", "GET", {})
        },
        // post() - Posts a campaign to POST /campaigns
        post: function(data){
            return query("/campaigns/", "POST", data)
        }
    },
    // campaignId contains the endpoints for /campaigns/:id
    campaignId : {
        // get() - Queries the API for GET /campaigns/:id
        get: function(id){
            return query("/campaigns/" + id, "GET", {})
        },
        // delete() - Deletes a campaign at DELETE /campaigns/:id
        delete: function(id){
            return query("/campaigns/" + id, "DELETE", data)
        }
    },
    // groups contains the endpoints for /groups
    groups : {
        // get() - Queries the API for GET /groups
        get: function(){
            return query("/groups/", "GET", {})
        },
        // post() - Posts a campaign to POST /groups
        post: function(group){
            return query("/groups/", "POST", group)
        }
    },
    // groupId contains the endpoints for /groups/:id
    groupId : {
        // get() - Queries the API for GET /groups/:id
        get: function(id){
            return query("/groups/" + id, "GET", {})
        },
        // put() - Puts a campaign to PUT /groups/:id
        put: function (group){
            return query("/groups/" + group.id, "PUT", group)
        },
        // delete() - Deletes a campaign at DELETE /groups/:id
        delete: function(id){
            return query("/groups/" + id, "DELETE", {})
        }
    },
    // templates contains the endpoints for /templates
    templates : {
        // get() - Queries the API for GET /templates
        get: function(){
            return query("/templates/", "GET", {})
        },
        // post() - Posts a campaign to POST /templates
        post: function(template){
            return query("/templates/", "POST", template)
        }
    },
    // templateId contains the endpoints for /templates/:id
    templateId : {
        // get() - Queries the API for GET /templates/:id
        get: function(id){
            return query("/templates/" + id, "GET", {})
        },
        // put() - Puts a campaign to PUT /templates/:id
        put: function (template){
            return query("/templates/" + template.id, "PUT", template)
        },
        // delete() - Deletes a campaign at DELETE /templates/:id
        delete: function(id){
            return query("/templates/" + id, "DELETE", {})
        }
    },
    // pages contains the endpoints for /pages
    pages : {
        // get() - Queries the API for GET /pages
        get: function(){
            return query("/pages/", "GET", {})
        },
        // post() - Posts a campaign to POST /pages
        post: function(page){
            return query("/pages/", "POST", page)
        }
    },
    // templateId contains the endpoints for /templates/:id
    pageId : {
        // get() - Queries the API for GET /templates/:id
        get: function(id){
            return query("/pages/" + id, "GET", {})
        },
        // put() - Puts a campaign to PUT /templates/:id
        put: function (page){
            return query("/pages/" + page.id, "PUT", page)
        },
        // delete() - Deletes a campaign at DELETE /templates/:id
        delete: function(id){
            return query("/pages/" + id, "DELETE", {})
        }
    },
    // import handles all of the "import" functions in the api
    import_email : function(raw) {
	return query("/import/email", "POST", {}) 
    },
    clone_site : function(req){
	return query("/import/site", "POST", req)
    }
}
