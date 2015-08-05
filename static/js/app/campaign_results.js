// labels is a map of campaign statuses to
// CSS classes
var labels = {
    "Email Sent" : "label-primary",
    "Email Opened" : "label-info",
    "Clicked Link" : "label-success",
	"Error" : "label-danger"
}

//
var classes = {
    "Email Sent" : {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent"
    },
    "Email Opened" : {
        slice: "ct-slice-donut-opened",
        legend: "ct-legend-opened"
    },
    "Clicked Link" : {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked"
    },
    "Error" : {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error"
    }
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
            // Set the title
            $("#page-title").text("Results for " + c.name)
            // Setup our graphs
            var timeline_data = {series:[{
                name: "Events",
                data: []
            }]}
            var email_data = {series:[]}
            var email_legend = {}
            var email_series_data = {}
            var timeline_opts = {
                axisX: {
                    showGrid: false,
                    type: Chartist.FixedScaleAxis,
                    divisor: 5,
                    labelInterpolationFnc: function(value){
                        return moment(value).format('MMMM Do YYYY h:mm')
                    }
                },
                axisY: {
                    type: Chartist.FixedScaleAxis,
                    ticks: [0, 1, 2],
                    low: 0,
                    showLabel: false
                },
                showArea: false,
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
                if (!email_series_data[result.status]){
                    email_series_data[result.status] = 1
                } else {
                    email_series_data[result.status]++;
                }
            })
            // Setup the graphs
            $.each(campaign.timeline, function(i, event){
                console.log(moment(event.time).format('MMMM Do YYYY h:mm'))
                timeline_data.series[0].data.push({meta : i, x: new Date(event.time), y:1})
            })
            console.log(timeline_data)
            $.each(email_series_data, function(status, count){
                email_data.series.push({meta: status, value: count})
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
                    top: (event.offsetY + 70 || event.originalEvent.layerY) - $toolTip.height() - 40
                });
            });
            var email_chart = new Chartist.Pie("#email_chart", email_data, email_opts)
            email_chart.on('draw', function(data){
                // We don't want to create the legend twice
                if (!email_legend[data.meta]) {
                    console.log(data.meta)
                    $("#email_chart_legend").append('<li><span class="' + classes[data.meta].legend + '"></span>' + data.meta + '</li>')
                    email_legend[data.meta] = true
                }
                data.element.addClass(classes[data.meta].slice)
            })
            // Setup the average chart listeners
            $piechart = $("#email_chart")
            var $pietoolTip = $piechart
            .append('<div class="chartist-tooltip"></div>')
            .find('.chartist-tooltip')
            .hide();

            $piechart.on('mouseenter', '.ct-slice-donut', function() {
                var $point = $(this)
                value = $point.attr('ct:value')
                label = $point.attr('ct:meta')
                $pietoolTip.html(label + ': ' + value.toString()).show();
            });

            $piechart.on('mouseleave', '.ct-slice-donut', function() {
                $pietoolTip.hide();
            });
            $piechart.on('mousemove', function(event) {
                $pietoolTip.css({
                    left: (event.offsetX || event.originalEvent.layerX) - $pietoolTip.width() / 2 - 10,
                    top: (event.offsetY + 40 || event.originalEvent.layerY) - $pietoolTip.height() - 80
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
