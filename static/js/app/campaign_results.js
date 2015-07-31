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
            var timeline_data = {labels:[],series:[]}
            var email_data = {series:[]}
            var timeline_opts = {
                axisX: {
                    showGrid: false,
                },
                axisY: {
                    type: Chartist.FixedScaleAxis,
                    ticks: [0, 1, 2],
                    low: 0,
                    showLabel: false
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
            // Setup the results table
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
            // Setup the graphs
            $.each(campaign.timeline, function(i, event){
                timeline_data.labels.push(moment(event.time).format('MMMM Do YYYY h:mm'))
                timeline_data.series.push([{meta : i, value: 1}])
            })
            var timeline_chart = new Chartist.Line('#timeline_chart', timeline_data, timeline_opts)

            // Setup the overview chart listeners
            $chart = $("#timeline_chart")
            var $toolTip = $chart
            .append('<div class="chartist-tooltip"></div>')
            .find('.chartist-tooltip')
            .hide();

            $chart.on('mouseenter', '.ct-point', function() {
                var $point = $(this)
                value = $point.attr('ct:value')
                cidx = $point.attr('ct:meta')
                html = "Event: " + campaign.timeline[cidx].message
                if (campaign.timeline[cidx].email) {
                    html += '<br>' + "Email: " + campaign.timeline[cidx].email
                }
                $toolTip.html(html).show()
            });

            $chart.on('mouseleave', '.ct-point', function() {
                $toolTip.hide();
            });
            $chart.on('mousemove', function(event) {
                $toolTip.css({
                    left: (event.offsetX || event.originalEvent.layerX) - $toolTip.width() / 2 - 10,
                    top: (event.offsetY + 0 || event.originalEvent.layerY) - $toolTip.height() - 40
                });
            });
            $("#loading").hide()
            $("#campaignResults").show()
        }
    })
    .error(function(){
        errorFlash(" Campaign not found!")
    })
})
