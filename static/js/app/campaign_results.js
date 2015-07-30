// labels is a map of campaign statuses to
// CSS classes
var labels = {
    "Email Sent" : "label-primary",
    "Email Opened" : "label-info",
	"Success" : "label-success",
    "Clicked Link" : "label-success",
	"Error" : "label-danger"
}

var campaign = {}

function dismiss(){
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#resultsTable").dataTable().DataTable().clear().draw()
}

// Deletes a campaign after prompting the user
function deleteCampaign(){
    if (confirm("Are you sure you want to delete: " + campaign.name + "?")){
        api.campaignId.delete(campaign.id)
        .success(function(msg){
            console.log(msg)
        })
        .error(function(e){
            $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
                <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
        })
    }
}

$(document).ready(function(){
    campaign.id = window.location.pathname.split('/').slice(-1)[0]
    api.campaignId.get(campaign.id)
    .success(function(c){
        campaign = c
        if (campaign){
            // Setup our graphs
            var timeline_chart = {labels:[],series:[[]]}
            var email_chart = {series:[]}
            var timeline_opts = {
                axisX: {
                    showGrid: false
                },
                showArea: true,
                plugins: []
            }
            var email_opts = {
                donut : true,
                donutWidth: 40,
                chartPadding: 0,
                showLabel: false
            }
            $("#loading").hide()
            $("#campaignResults").show()
            resultsTable = $("#resultsTable").DataTable();
            $.each(campaign.results, function(i, result){
                label = labels[result.status] || "label-default";
                resultsTable.row.add([
                    result.first_name || "",
                    result.last_name || "",
                    result.email || "",
                    result.position || "",
                    "<span class=\"label " + label + "\">" + result.status + "</span>"
                ]).draw()
            })
        }
    })
    .error(function(){
        errorFlash(" Campaign not found!")
    })
})
