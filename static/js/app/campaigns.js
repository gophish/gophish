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
    campaigns.post(campaign)
    .success(function(data){
        successFlash("submitted!")
        console.log(data)
    })
    .error(function(data){
        $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
    })
}
$(document).ready(function(){
    var campaignData = {}
    campaigns.get()
    .success(function(data){
        successFlash("worked!")
        campaignData = data
    })
    .error(function(data){
        errorFlash("No work")
    })
    campaignTable = $("#campaignTable").DataTable();
    $.each(campaignData, function(i, campaign){
        campaignTable.row.add([
            campaign.created_date,
            campaign.name,
            campaign.status
        ]).draw()
    })
})
