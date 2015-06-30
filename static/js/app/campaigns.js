// Save attempts to POST to /campaigns/
function save(){
    var campaign = {
        name: $("#name").val(),
        template:{
            name: $("#template").val()
        },
        smtp: {
            from_address: $("input[name=from]").val(),
            host: $("input[name=host]").val(),
            username: $("input[name=username]").val(),
            password: $("input[name=password]").val(),
        },
        groups: [{name : "Morning catch employees"}]
    }
    // Submit the campaign
    api.campaigns.post(campaign)
    .success(function(data){
        successFlash("Campaign successfully launched!")
        load()
    })
    .error(function(data){
        $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
    })
}

function dismiss(){
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
}

function edit(campaign){
    if (campaign == "new") {
        api.groups.get()
        .success(function(groups){
            if (groups.length == 0){
                modalError("No groups found!")
                return false;
            }
            else {
                // Create the group typeahead objects
                var groupTable = $("#groupTable").DataTable()
                var suggestion_template = Hogan.compile('<div>{{name}}</div>')
                var bh = new Bloodhound({
                    datumTokenizer: function(g) { return Bloodhound.tokenizers.whitespace(g.name) },
                    queryTokenizer: Bloodhound.tokenizers.whitespace,
                    local: groups
                })
                bh.initialize()
                $("#groups.typeahead.form-control").typeahead({
                    hint: true,
                    highlight: true,
                    minLength: 1
                },
                {
                    name: "groups",
                    source: bh,
                    templates: {
                        empty: function(data) {return '<div class="tt-suggestion">No groups matched that query</div>' },
                        suggestion: function(data){ return '<div>' + data.name + '</div>' }
                    }
                })
                .bind('typeahead:select', function(ev, group){
                    groupTable.row.add([
                        group.name,
                        '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
                    ]).draw()
                });
                //<span ng-click="removeGroup(group)" class="remove-row"><i class="fa fa-trash-o"></i>
                //</span>
            }
        })
    }
}

function load(){
    api.campaigns.get()
    .success(function(campaigns){
        if (campaigns.length > 0){
            $("#emptyMessage").hide()
            $("#campaignTable").show()
            campaignTable = $("#campaignTable").DataTable();
            $.each(campaigns, function(i, campaign){
                campaignTable.row.add([
                    campaign.created_date,
                    campaign.name,
                    campaign.status
                ]).draw()
            })
        }
    })
    .error(function(){
        errorFlash("Error fetching campaigns")
    })
}

$(document).ready(function(){
    load()
})
