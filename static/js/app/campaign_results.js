var map = null

// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope"
    },
    "Email Opened": {
        slice: "ct-slice-donut-opened",
        legend: "ct-legend-opened",
        label: "label-warning",
        icon: "fa-envelope"
    },
    "Clicked Link": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-mouse-pointer"
    },
    "Success": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-exclamation"
    },
    "Error": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-times"
    },
    "Unknown": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-question"
    },
    "Campaign Created": {
        label: "label-success",
        icon: "fa-rocket"
    }
}

var campaign = {}

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#resultsTable").dataTable().DataTable().clear().draw()
}

// Deletes a campaign after prompting the user
function deleteCampaign() {
    if (confirm("Are you sure you want to delete: " + campaign.name + "?")) {
        api.campaignId.delete(campaign.id)
            .success(function(msg) {
                location.href = '/campaigns'
            })
            .error(function(e) {
                $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
                <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
            })
    }
}

// Exports campaign results as a CSV file
function exportAsCSV() {
    exportHTML = $("#exportButton").html()
    $("#exportButton").html('<i class="fa fa-spinner fa-spin"></i>')
    var csvString = Papa.unparse(campaign.results, {})
    var csvData = new Blob([csvString], {
        type: 'text/csv;charset=utf-8;'
    });
    if (navigator.msSaveBlob) {
        navigator.msSaveBlob(csvData, 'results.csv');
    } else {
        var csvURL = window.URL.createObjectURL(csvData);
        var dlLink = document.createElement('a');
        dlLink.href = csvURL;
        dlLink.setAttribute('download', 'results.csv');
        dlLink.click();
    }
    $("#exportButton").html(exportHTML)
}

function renderTimeline(data) {
    record = {
        "first_name": data[1],
        "last_name": data[2],
        "email": data[3],
        "position": data[4]
    }
    results = '<div class="timeline col-sm-12 well well-lg">' +
        '<h6>Timeline for ' + record.first_name + ' ' + record.last_name +
        '</h6><span class="subtitle">Email: ' + record.email + '</span>' +
        '<div class="timeline-graph col-sm-6">'
    $.each(campaign.timeline, function(i, event) {
        if (!event.email || event.email == record.email) {
            // Add the event
            results += '<div class="timeline-entry">' +
                '    <div class="timeline-bar"></div>'
            results +=
                '    <div class="timeline-icon ' + statuses[event.message].label + '">' +
                '    <i class="fa ' + statuses[event.message].icon + '"></i></div>' +
                '    <div class="timeline-message">' + event.message +
                '    <span class="timeline-date">' + moment(event.time).format('MMMM Do YYYY h:mm') + '</span></div>'
            results += '</div>'
        }
    })
    results += '</div></div>'
    return results
}
$(document).ready(function() {
    campaign.id = window.location.pathname.split('/').slice(-1)[0]
    api.campaignId.get(campaign.id)
        .success(function(c) {
            campaign = c
            if (campaign) {
                // Set the title
                $("#page-title").text("Results for " + c.name)
                    // Setup tooltips
                $('[data-toggle="tooltip"]').tooltip()
                    // Setup our graphs
                var timeline_data = {
                    series: [{
                        name: "Events",
                        data: []
                    }]
                }
                var email_data = {
                    series: []
                }
                var email_legend = {}
                var email_series_data = {}
                var timeline_opts = {
                    axisX: {
                        showGrid: false,
                        type: Chartist.FixedScaleAxis,
                        divisor: 5,
                        labelInterpolationFnc: function(value) {
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
                        donut: true,
                        donutWidth: 40,
                        chartPadding: 0,
                        showLabel: false
                    }
                    // Setup the results table
                resultsTable = $("#resultsTable").DataTable({
                    destroy: true,
                    destroy: true,
                    "order": [
                        [1, "asc"]
                    ],
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }, {
                        className: "details-control",
                        "targets": [0]
                    }]
                });
                $.each(campaign.results, function(i, result) {
                        label = statuses[result.status].label || "label-default";
                        resultsTable.row.add([
                            "<i class=\"fa fa-caret-right\"></i>",
                            result.first_name || "",
                            result.last_name || "",
                            result.email || "",
                            result.position || "",
                            "<span class=\"label " + label + "\">" + result.status + "</span>"
                        ]).draw()
                        if (!email_series_data[result.status]) {
                            email_series_data[result.status] = 1
                        } else {
                            email_series_data[result.status]++;
                        }
                    })
                    // Setup the individual timelines
                $('#resultsTable tbody').on('click', 'td.details-control', function() {
                    var tr = $(this).closest('tr');
                    var row = resultsTable.row(tr);

                    if (row.child.isShown()) {
                        // This row is already open - close it
                        row.child.hide();
                        tr.removeClass('shown');
                        $(this).find("i").removeClass("fa-caret-down")
                        $(this).find("i").addClass("fa-caret-right")
                    } else {
                        // Open this row
                        $(this).find("i").removeClass("fa-caret-right")
                        $(this).find("i").addClass("fa-caret-down")
                        row.child(renderTimeline(row.data())).show();
                        tr.addClass('shown');
                    }
                });
                // Setup the graphs
                $.each(campaign.timeline, function(i, event) {
                    timeline_data.series[0].data.push({
                        meta: i,
                        x: new Date(event.time),
                        y: 1
                    })
                })
                $.each(email_series_data, function(status, count) {
                    email_data.series.push({
                        meta: status,
                        value: count
                    })
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
                email_chart.on('draw', function(data) {
                        // We don't want to create the legend twice
                        if (!email_legend[data.meta]) {
                            console.log(data.meta)
                            $("#email_chart_legend").append('<li><span class="' + statuses[data.meta].legend + '"></span>' + data.meta + '</li>')
                            email_legend[data.meta] = true
                        }
                        data.element.addClass(statuses[data.meta].slice)
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
                map = new Datamap({
                    element: document.getElementById("resultsMap"),
                    responsive: true,
                    fills: {
                        defaultFill: "#ffffff",
                        point: "#283F50"
                    },
                    geographyConfig: {
                        highlightFillColor: "#1abc9c",
                        borderColor: "#283F50"
                    },
                    bubblesConfig: {
                        borderColor: "#283F50"
                    }
                });
                bubbles = []
                $.each(campaign.results, function(i, result) {
                    // Check that it wasn't an internal IP
                    if (result.latitude == 0 && result.longitude == 0) {
                        return true;
                    }
                    newIP = true
                    $.each(bubbles, function(i, bubble) {
                        if (bubble.ip == result.ip) {
                            bubbles[i].radius += 1
                            newIP = false
                            return false
                        }
                    })
                    if (newIP) {
                        console.log("Adding bubble at: ")
                        console.log({
                            latitude: result.latitude,
                            longitude: result.longitude,
                            name: result.ip,
                            fillKey: "point"
                        })
                        bubbles.push({
                            latitude: result.latitude,
                            longitude: result.longitude,
                            name: result.ip,
                            fillKey: "point",
                            radius: 2
                        })
                    }
                })
                map.bubbles(bubbles)
            }
            // Load up the map data (only once!)
            $('a[data-toggle="tab"]').on('shown.bs.tab', function(e) {
                if ($(e.target).attr('href') == "#overview") {
                    if (!map) {
                        map = new Datamap({
                            element: document.getElementById("resultsMap"),
                            responsive: true,
                            fills: {
                                defaultFill: "#ffffff"
                            },
                            geographyConfig: {
                                highlightFillColor: "#1abc9c",
                                borderColor: "#283F50"
                            }
                        });
                    }
                }
            })
        })
        .error(function() {
            $("#loading").hide()
            errorFlash(" Campaign not found!")
        })
})
