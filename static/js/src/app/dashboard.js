var campaigns = []
    // labels is a map of campaign statuses to
    // CSS classes
var labels = {
    "In progress": "label-primary",
    "Queued": "label-info",
    "Completed": "label-success",
    "Emails Sent": "label-success",
    "Error": "label-danger"
}

function deleteCampaign(idx) {
    if (confirm("Delete " + campaigns[idx].name + "?")) {
        api.campaignId.delete(campaigns[idx].id)
            .success(function(data) {
                successFlash(data.message)
                location.reload()
            })
    }
}

$(document).ready(function() {
    api.campaigns.summary()
        .success(function(data) {
            $("#loading").hide()
            campaigns = data.campaigns
            if (campaigns.length > 0) {
                $("#dashboard").show()
                    // Create the overview chart data
                var overview_data = {
                    labels: [],
                    series: [
                        []
                    ]
                }
                var average_data = {
                    series: []
                }
                var overview_opts = {
                    axisX: {
                        showGrid: false
                    },
                    showArea: true,
                    plugins: [],
                    low: 0,
                    high: 100
                }
                var average_opts = {
                    donut: true,
                    donutWidth: 40,
                    chartPadding: 0,
                    showLabel: false
                }
                var average = 0
                campaignTable = $("#campaignTable").DataTable({
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }],
                    order: [
                        [1, "desc"]
                    ]
                });
                $.each(campaigns, function(i, campaign) {
                    var campaign_date = moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a')
                    var label = labels[campaign.status] || "label-default";
                    //section for tooltips on the status of a campaign to show some quick stats
                    var launchDate;
                    if (moment(campaign.launch_date).isAfter(moment())) {
                        launchDate = "Scheduled to start: " + moment(campaign.launch_date).format('MMMM Do YYYY, h:mm:ss a')
                        var quickStats = launchDate + "<br><br>" + "Number of recipients: " + campaign.stats.total
                    } else {
                        launchDate = "Launch Date: " + moment(campaign.launch_date).format('MMMM Do YYYY, h:mm:ss a')
                        var quickStats = launchDate + "<br><br>" + "Number of recipients: " + campaign.stats.total + "<br><br>" + "Emails opened: " + campaign.stats.opened + "<br><br>" + "Emails clicked: " + campaign.stats.clicked + "<br><br>" + "Submitted Credentials: " + campaign.stats.submitted_data + "<br><br>" + "Errors : " + campaign.stats.error
                    }
                    // Add it to the table
                    campaignTable.row.add([
                            escapeHtml(campaign.name),
                            campaign_date,
                            "<span class=\"label " + label + "\" data-toggle=\"tooltip\" data-placement=\"right\" data-html=\"true\" title=\"" + quickStats + "\">" + campaign.status + "</span>",
                            "<div class='pull-right'><a class='btn btn-primary' href='/campaigns/" + campaign.id + "' data-toggle='tooltip' data-placement='left' title='View Results'>\
                    <i class='fa fa-bar-chart'></i>\
                    </a>\
                    <button class='btn btn-danger' onclick='deleteCampaign(" + i + ")' data-toggle='tooltip' data-placement='left' title='Delete Campaign'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                        ]).draw()
                    $('[data-toggle="tooltip"]').tooltip()
                        // Add it to the chart data
                    campaign.y = 0
                    campaign.y += campaign.stats.clicked + campaign.stats.submitted_data
                    campaign.y = Math.floor((campaign.y / campaign.stats.total) * 100)
                    average += campaign.y
                        // Add the data to the overview chart
                    overview_data.labels.push(campaign_date)
                    overview_data.series[0].push({
                        meta: i,
                        value: campaign.y
                    })
                })
                average = Math.floor(average / data.total);
                average_data.series.push({
                    meta: "Unsuccessful Phishes",
                    value: 100 - average
                })
                average_data.series.push({
                        meta: "Successful Phishes",
                        value: average
                    })
                    // Build the charts
                var average_chart = new Chartist.Pie("#average_chart", average_data, average_opts)
                var overview_chart = new Chartist.Line('#overview_chart', overview_data, overview_opts)
                    // Setup the average chart listeners
                $piechart = $("#average_chart")
                var $pietoolTip = $piechart
                    .append('<div class="chartist-tooltip"></div>')
                    .find('.chartist-tooltip')
                    .hide();

                $piechart.on('mouseenter', '.ct-slice-donut', function() {
                    var $point = $(this)
                    value = $point.attr('ct:value')
                    label = $point.attr('ct:meta')
                    $pietoolTip.html(label + ': ' + value.toString() + "%").show();
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

                // Setup the overview chart listeners
                $chart = $("#overview_chart")
                var $toolTip = $chart
                    .append('<div class="chartist-tooltip"></div>')
                    .find('.chartist-tooltip')
                    .hide();

                $chart.on('mouseenter', '.ct-point', function() {
                    var $point = $(this)
                    value = $point.attr('ct:value') || 0
                    cidx = $point.attr('ct:meta')
                    $toolTip.html(campaigns[cidx].name + '<br>' + "Successes: " + value.toString() + "%").show();
                });

                $chart.on('mouseleave', '.ct-point', function() {
                    $toolTip.hide();
                });
                $chart.on('mousemove', function(event) {
                    $toolTip.css({
                        left: (event.offsetX || event.originalEvent.layerX) - $toolTip.width() / 2 - 10,
                        top: (event.offsetY + 40 || event.originalEvent.layerY) - $toolTip.height() - 40
                    });
                });
                $("#overview_chart").on("click", ".ct-point", function(e) {
                    var $cidx = $(this).attr('ct:meta');
                    window.location.href = "/campaigns/" + campaigns[cidx].id
                });
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            errorFlash("Error fetching campaigns")
        })
})
