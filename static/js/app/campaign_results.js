var map = null
var doPoll = true;

// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
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

var campaign = {}
var bubbles = []

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#resultsTable").dataTable().DataTable().clear().draw()
}

// Deletes a campaign after prompting the user
function deleteCampaign() {
    swal({
        title: "Are you sure?",
        text: "This will delete the campaign. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete Campaign",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function() {
            return new Promise(function(resolve, reject) {
                api.campaignId.delete(campaign.id)
                    .success(function(msg) {
                        resolve()
                    })
                    .error(function(data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function() {
        swal(
            'Campaign Deleted!',
            'This campaign has been deleted!',
            'success'
        );
        $('button:contains("OK")').on('click', function() {
            location.href = '/campaigns'
        })
    })
}

// Completes a campaign after prompting the user
function completeCampaign() {
    swal({
        title: "Are you sure?",
        text: "Gophish will stop processing events for this campaign",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Complete Campaign",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function() {
            return new Promise(function(resolve, reject) {
                api.campaignId.complete(campaign.id)
                    .success(function(msg) {
                        resolve()
                    })
                    .error(function(data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function() {
        swal(
            'Campaign Completed!',
            'This campaign has been completed!',
            'success'
        );
        $('#complete_button')[0].disabled = true;
        $('#complete_button').text('Completed!')
        doPoll = false;
    })
}

// Exports campaign results as a CSV file
function exportAsCSV(scope) {
    exportHTML = $("#exportButton").html()
    var csvScope = null
    switch (scope) {
        case "results":
            csvScope = campaign.results
            break;
        case "events":
            csvScope = campaign.timeline
            break;
    }
    if (!csvScope) {
        return
    }
    $("#exportButton").html('<i class="fa fa-spinner fa-spin"></i>')
    var csvString = Papa.unparse(csvScope, {})
    var csvData = new Blob([csvString], {
        type: 'text/csv;charset=utf-8;'
    });
    if (navigator.msSaveBlob) {
        navigator.msSaveBlob(csvData, scope + '.csv');
    } else {
        var csvURL = window.URL.createObjectURL(csvData);
        var dlLink = document.createElement('a');
        dlLink.href = csvURL;
        dlLink.setAttribute('download', scope + '.csv');
        document.body.appendChild(dlLink)
        dlLink.click();
        document.body.removeChild(dlLink)
    }
    $("#exportButton").html(exportHTML)
}

function replay(event_idx) {
    request = campaign.timeline[event_idx]
    details = JSON.parse(request.details)
    url = null
    form = $('<form>').attr({
            method: 'POST',
            target: '_blank',
        })
        /* Create a form object and submit it */
    $.each(Object.keys(details.payload), function(i, param) {
            if (param == "rid") {
                return true;
            }
            if (param == "__original_url") {
                url = details.payload[param];
                return true;
            }
            $('<input>').attr({
                name: param,
            }).val(details.payload[param]).appendTo(form);
        })
        /* Ensure we know where to send the user */
        // Prompt for the URL
    swal({
        title: 'Where do you want the credentials submitted to?',
        input: 'text',
        showCancelButton: true,
        inputPlaceholder: "http://example.com/login",
        inputValue: url || "",
        inputValidator: function(value) {
            return new Promise(function(resolve, reject) {
                if (value) {
                    resolve();
                } else {
                    reject('Invalid URL.');
                }
            });
        }
    }).then(function(result) {
        url = result
        submitForm()
    })
    return
    submitForm()

    function submitForm() {
        form.attr({
            action: url
        })
        form.appendTo('body').submit().remove()
    }
}

function renderTimeline(data) {
    record = {
        "first_name": data[2],
        "last_name": data[3],
        "email": data[4],
        "position": data[5]
    }
    results = '<div class="timeline col-sm-12 well well-lg">' +
        '<h6>Timeline for ' + escapeHtml(record.first_name) + ' ' + escapeHtml(record.last_name) +
        '</h6><span class="subtitle">Email: ' + escapeHtml(record.email) + '</span>' +
        '<div class="timeline-graph col-sm-6">'
    $.each(campaign.timeline, function(i, event) {
        if (!event.email || event.email == record.email) {
            // Add the event
            results += '<div class="timeline-entry">' +
                '    <div class="timeline-bar"></div>'
            results +=
                '    <div class="timeline-icon ' + statuses[event.message].label + '">' +
                '    <i class="fa ' + statuses[event.message].icon + '"></i></div>' +
                '    <div class="timeline-message">' + escapeHtml(event.message) +
                '    <span class="timeline-date">' + moment(event.time).format('MMMM Do YYYY h:mm a') + '</span>'
            if (event.details) {
                if (event.message == "Submitted Data") {
                    results += '<div class="timeline-replay-button"><button onclick="replay(' + i + ')" class="btn btn-success">'
                    results += '<i class="fa fa-refresh"></i> Replay Credentials</button></div>'
                    results += '<div class="timeline-event-details"><i class="fa fa-caret-right"></i> View Details</div>'
                }
                details = JSON.parse(event.details)
                if (details.payload) {
                    results += '<div class="timeline-event-results">'
                    results += '    <table class="table table-condensed table-bordered table-striped">'
                    results += '        <thead><tr><th>Parameter</th><th>Value(s)</tr></thead><tbody>'
                    $.each(Object.keys(details.payload), function(i, param) {
                        if (param == "rid") {
                            return true;
                        }
                        results += '    <tr>'
                        results += '        <td>' + escapeHtml(param) + '</td>'
                        results += '        <td>' + escapeHtml(details.payload[param]) + '</td>'
                        results += '    </tr>'
                    })
                    results += '       </tbody></table>'
                    results += '</div>'
                }
                if (details.error) {
                    results += '<div class="timeline-event-details"><i class="fa fa-caret-right"></i> View Details</div>'
                    results += '<div class="timeline-event-results">'
                    results += '<span class="label label-default">Error</span> ' + details.error
                    results += '</div>'
                }
            }
            results += '</div></div>'
        }
    })
    results += '</div></div>'
    return results
}


/* poll - Queries the API and updates the UI with the results
 *
 * Updates:
 * * Timeline Chart
 * * Email (Donut) Chart
 * * Map Bubbles
 * * Datatables
 */
function poll() {
    api.campaignId.results(campaign.id)
        .success(function(c) {
            campaign = c
                /* Update the timeline */
            var timeline_data = {
                series: [{
                    name: "Events",
                    data: []
                }]
            }
            $.each(campaign.timeline, function(i, event) {
                timeline_data.series[0].data.push({
                    meta: i,
                    x: new Date(event.time),
                    y: 1
                })
            })
            var timeline_chart = $("#timeline_chart")
            if (timeline_chart.get(0).__chartist__) {
                timeline_chart.get(0).__chartist__.update(timeline_data)
            }
            /* Update the results donut chart */
            var email_data = {
                series: []
            }
            var email_series_data = {}
            $.each(campaign.results, function(i, result) {
                if (!email_series_data[result.status]) {
                    email_series_data[result.status] = 1
                } else {
                    email_series_data[result.status]++;
                }
            })
            $("#email_chart_legend").html("")
            $.each(email_series_data, function(status, count) {
                email_data.series.push({
                    meta: status,
                    value: count
                })
                $("#email_chart_legend").append('<li><span class="' + statuses[status].legend + '"></span>' + status + '</li>')
            })
            var email_chart = $("#email_chart")
            if (email_chart.get(0).__chartist__) {
                email_chart.get(0).__chartist__.on('draw', function(data) {
                        data.element.addClass(statuses[data.meta].slice)
                    })
                    // Update with the latest data
                email_chart.get(0).__chartist__.update(email_data)
            }
            /* Update the datatable */
            resultsTable = $("#resultsTable").DataTable()
            resultsTable.rows().every(function(i, tableLoop, rowLoop) {
                    var row = this.row(i)
                    var rowData = row.data()
                    var rid = rowData[0]
                    $.each(campaign.results, function(j, result) {
                        if (result.id == rid) {
                            var label = statuses[result.status].label || "label-default";
                            rowData[6] = "<span class=\"label " + label + "\">" + result.status + "</span>"
                            resultsTable.row(i).data(rowData).draw(false)
                            if (row.child.isShown()) {
                                row.child(renderTimeline(row.data()))
                            }
                            return false
                        }
                    })
                })
                /* Update the map information */
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
        })
}

function load() {
    campaign.id = window.location.pathname.split('/').slice(-1)[0]
    api.campaignId.results(campaign.id)
        .success(function(c) {
            campaign = c
            if (campaign) {
                $("title").text(c.name + " - Gophish")
                $("#loading").hide()
                $("#campaignResults").show()
                    // Set the title
                $("#page-title").text("Results for " + c.name)
                if (c.status == "Completed") {
                    $('#complete_button')[0].disabled = true;
                    $('#complete_button').text('Completed!');
                    doPoll = false;
                }
                // Setup tooltips
                $('[data-toggle="tooltip"]').tooltip()
                    // Setup viewing the details of a result
                $("#resultsTable").on("click", ".timeline-event-details", function() {
                        // Show the parameters
                        payloadResults = $(this).parent().find(".timeline-event-results")
                        if (payloadResults.is(":visible")) {
                            $(this).find("i").removeClass("fa-caret-down")
                            $(this).find("i").addClass("fa-caret-right")
                            payloadResults.hide()
                        } else {
                            $(this).find("i").removeClass("fa-caret-right")
                            $(this).find("i").addClass("fa-caret-down")
                            payloadResults.show()
                        }
                    })
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
                            return moment(value).format('MMMM Do YYYY h:mm a')
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
                    "order": [
                        [2, "asc"]
                    ],
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }, {
                        className: "details-control",
                        "targets": [1]
                    }, {
                        "visible": false,
                        "targets": [0]
                    }]
                });
                resultsTable.clear();
                $.each(campaign.results, function(i, result) {
                        label = statuses[result.status].label || "label-default";
                        resultsTable.row.add([
                            result.id,
                            "<i class=\"fa fa-caret-right\"></i>",
                            escapeHtml(result.first_name) || "",
                            escapeHtml(result.last_name) || "",
                            escapeHtml(result.email) || "",
                            escapeHtml(result.position) || "",
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
                        row.invalidate('dom').draw(false)
                    } else {
                        // Open this row
                        $(this).find("i").removeClass("fa-caret-right")
                        $(this).find("i").addClass("fa-caret-down")
                        row.child(renderTimeline(row.data())).show();
                        tr.addClass('shown');
                        row.invalidate('dom').draw(false)
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
                $("#email_chart_legend").html("")
                $.each(email_series_data, function(status, count) {
                    email_data.series.push({
                        meta: status,
                        value: count
                    })
                    $("#email_chart_legend").append('<li><span class="' + statuses[status].legend + '"></span>' + status + '</li>')
                })
                var timeline_chart = new Chartist.Line('#timeline_chart', timeline_data, timeline_opts)
                timeline_chart.on('draw', function(data) {
                        if (data.type === "point") {
                            var point_style = statuses[campaign.timeline[data.meta].message].point
                            var circle = new Chartist.Svg("circle", {
                                cx: [data.x],
                                cy: [data.y],
                                r: 5,
                                fill: "#283F50",
                                meta: data.meta,
                                value: 1,
                            }, point_style + ' ct-timeline-point')
                            data.element.replace(circle)
                        }
                    })
                    // Setup the overview chart listeners
                $chart = $("#timeline_chart")
                var $toolTip = $chart
                    .append('<div class="chartist-tooltip"></div>')
                    .find('.chartist-tooltip')
                    .hide();
                $chart.on('mouseenter', '.ct-timeline-point', function() {
                    var $point = $(this)
                    cidx = $point.attr('meta')
                    html = "Event: " + campaign.timeline[cidx].message
                    if (campaign.timeline[cidx].email) {
                        html += '<br>' + "Email: " + escapeHtml(campaign.timeline[cidx].email)
                    }
                    $toolTip.html(html).show()
                });
                $chart.on('mouseleave', '.ct-timeline-point', function() {
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
                if (!map) {
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
                }
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
}
$(document).ready(function() {
    load();
    // Start the polling loop
    function refresh() {
        if (!doPoll) {
            return;
        }
        $("#refresh_message").show()
        poll()
        $("#refresh_message").hide()
        setTimeout(refresh, 10000)
    };
    // Start the polling loop
    setTimeout(refresh, 10000)
})
