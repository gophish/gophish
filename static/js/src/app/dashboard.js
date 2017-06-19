var campaigns = []

// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "Emails Sent": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "In progress": {
        label: "label-primary"
    },
    "Queued": {
        label: "label-info"
    },
    "Completed": {
        label: "label-success"
    },
    "Email Opened": {
        slice: "ct-slice-donut-opened",
        legend: "ct-legend-opened",
        label: "label-warning",
        icon: "fa-envelope",
        point: "ct-point-opened"
    },
    "Clicked Link": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-clicked",
        icon: "fa-mouse-pointer",
        point: "ct-point-clicked"
    },
    "Success": {
        slice: "ct-slice-donut-success",
        legend: "ct-legend-success",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Error": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Error Sending Email": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Submitted Data": {
        slice: "ct-slice-donut-success",
        legend: "ct-legend-success",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Unknown": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-question",
        point: "ct-point-error"
    },
    "Sending": {
        slice: "ct-slice-donut-sending",
        legend: "ct-legend-sending",
        label: "label-primary",
        icon: "fa-spinner",
        point: "ct-point-sending"
    },
    "Campaign Created": {
        label: "label-success",
        icon: "fa-rocket"
    }
}

var statsMapping = {
    "sent": "Email Sent",
    "opened": "Email Opened",
    "clicked": "Clicked Link",
    "submitted_data": "Submitted Data",
    "error": "Error"
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

function generateStatsPieChart(campaigns) {
    var stats_opts = {
        donut: true,
        donutWidth: 40,
        chartPadding: 0,
        showLabel: false
    }
    var stats_series_data = {}
    var stats_data = {
        series: []
    }
    var total = 0

    $.each(campaigns, function(i, campaign) {
        $.each(campaign.stats, function(status, count) {
            if (status == "total") {
                total += count
                return true
            }
            if (!stats_series_data[status]) {
                stats_series_data[status] = count;
            } else {
                stats_series_data[status] += count;
            }
        })
    })
    $.each(stats_series_data, function(status, count) {
        // I don't like this, but I guess it'll have to work.
        // Turns submitted_data into Submitted Data
        status_label = statsMapping[status]
        stats_data.series.push({
            meta: status_label,
            value: Math.floor((count / total) * 100)
        })
        $("#stats_chart_legend").append('<li><span class="' + statuses[status_label].legend + '"></span>' + status_label + '</li>')
    })

    var stats_chart = new Chartist.Pie("#stats_chart", stats_data, stats_opts)

    $piechart = $("#stats_chart")
    var $pietoolTip = $piechart
        .append('<div class="chartist-tooltip"></div>')
        .find('.chartist-tooltip')
        .hide();

    $piechart.get(0).__chartist__.on('draw', function(data) {
            data.element.addClass(statuses[data.meta].slice)
        })
        // Update with the latest data
    $piechart.get(0).__chartist__.update(stats_data)

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
                var overview_opts = {
                    axisX: {
                        showGrid: false
                    },
                    showArea: true,
                    plugins: [],
                    low: 0,
                    high: 100
                }
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
                        var label = statuses[campaign.status].label || "label-default";
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
                            // Add the data to the overview chart
                        overview_data.labels.push(campaign_date)
                        overview_data.series[0].push({
                            meta: i,
                            value: campaign.y
                        })
                    })
                    // Build the charts
                generateStatsPieChart(campaigns)
                var overview_chart = new Chartist.Line('#overview_chart', overview_data, overview_opts)

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
