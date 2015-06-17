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

function groupAdd(name){
    groups.append({
        name: name
    })
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
