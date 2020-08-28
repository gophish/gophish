var map = null
var doPoll = true;
var customEventsPieCharts = true; // Include pie charts for custom events or not
                               // Setting to true will add the custom events to statusMapping as wel as adding HTML chart elements.

// statuses is a helper map to point result statuses to ui classes
var statuses = {

    "Email Sent": {
        color: "#1abc9c",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "Emails Sent": {
        color: "#1abc9c",
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
        color: "#f9bf3b",
        label: "label-warning",
        icon: "fa-envelope-open",
        point: "ct-point-opened"
    },
    "Clicked Link": {
        color: "#F39C12",
        label: "label-clicked",
        icon: "fa-mouse-pointer",
        point: "ct-point-clicked"
    },
    "Success": {
        color: "#f05b4f",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    //not a status, but is used for the campaign timeline and user timeline
    "Email Reported": {
        color: "#45d6ef",
        label: "label-info",
        icon: "fa-bullhorn",
        point: "ct-point-reported"
    },
    "Error": {
        color: "#6c7a89",
        label: "label-default",
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Error Sending Email": {
        color: "#6c7a89",
        label: "label-default",
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Submitted Data": {
        color: "#f05b4f",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Unknown": {
        color: "#6c7a89",
        label: "label-default",
        icon: "fa-question",
        point: "ct-point-error"
    },
    "Sending": {
        color: "#428bca",
        label: "label-primary",
        icon: "fa-spinner",
        point: "ct-point-sending"
    },
    "Retrying": {
        color: "#6c7a89",
        label: "label-default",
        icon: "fa-clock-o",
        point: "ct-point-error"
    },
    "Scheduled": {
        color: "#428bca",
        label: "label-primary",
        icon: "fa-clock-o",
        point: "ct-point-sending"
    },
    "Campaign Created": {
        label: "label-success",
        icon: "fa-rocket"
    }
}

var statusMapping = {
    "Email Sent": "sent",
    "Email Opened": "opened",
    "Clicked Link": "clicked",
    "Submitted Data": "submitted_data",
    "Email Reported": "reported",
    //"Opened Word Document" : "opened_word_document"
}

// This is an underwhelming attempt at an enum
// until I have time to refactor this appropriately.
var progressListing = [
    "Email Sent",
    "Email Opened",
    "Clicked Link",
    "Submitted Data"
]

var campaign = {}
var bubbles = []

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#resultsTable").dataTable().DataTable().clear().draw()
}

// Deletes a campaign after prompting the user
function deleteCampaign() {
    Swal.fire({
        title: "Are you sure?",
        text: "This will delete the campaign. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete Campaign",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        showLoaderOnConfirm: true,
        preConfirm: function () {
            return new Promise(function (resolve, reject) {
                api.campaignId.delete(campaign.id)
                    .success(function (msg) {
                        resolve()
                    })
                    .error(function (data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function (result) {
        if(result.value){
            Swal.fire(
                'Campaign Deleted!',
                'This campaign has been deleted!',
                'success'
            );
        }
        $('button:contains("OK")').on('click', function () {
            location.href = '/campaigns'
        })
    })
}

// Completes a campaign after prompting the user
function completeCampaign() {
    Swal.fire({
        title: "Are you sure?",
        text: "Gophish will stop processing events for this campaign",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Complete Campaign",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        showLoaderOnConfirm: true,
        preConfirm: function () {
            return new Promise(function (resolve, reject) {
                api.campaignId.complete(campaign.id)
                    .success(function (msg) {
                        resolve()
                    })
                    .error(function (data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function (result) {
        if (result.value){
            Swal.fire(
                'Campaign Completed!',
                'This campaign has been completed!',
                'success'
            );
            $('#complete_button')[0].disabled = true;
            $('#complete_button').text('Completed!')
            doPoll = false;
        }
    })
}

// Exports campaign results as a CSV file
function exportAsCSV(scope) {
    exportHTML = $("#exportButton").html()
    var csvScope = null
    var filename = campaign.name + ' - ' + capitalize(scope) + '.csv'
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
    var csvString = Papa.unparse(csvScope, {
        'escapeFormulae': true
    })
    var csvData = new Blob([csvString], {
        type: 'text/csv;charset=utf-8;'
    });
    if (navigator.msSaveBlob) {
        navigator.msSaveBlob(csvData, filename);
    } else {
        var csvURL = window.URL.createObjectURL(csvData);
        var dlLink = document.createElement('a');
        dlLink.href = csvURL;
        dlLink.setAttribute('download', filename)
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
    $.each(Object.keys(details.payload), function (i, param) {
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
    Swal.fire({
        title: 'Where do you want the credentials submitted to?',
        input: 'text',
        showCancelButton: true,
        inputPlaceholder: "http://example.com/login",
        inputValue: url || "",
        inputValidator: function (value) {
            return new Promise(function (resolve, reject) {
                if (value) {
                    resolve();
                } else {
                    reject('Invalid URL.');
                }
            });
        }
    }).then(function (result) {
        if (result.value){
            url = result.value
            submitForm()
        }
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

/**
 * Returns an HTML string that displays the OS and browser that clicked the link
 * or submitted credentials.
 * 
 * @param {object} event_details - The "details" parameter for a campaign
 *  timeline event
 * 
 */
var renderDevice = function (event_details) {
    var ua = UAParser(details.browser['user-agent'])
    var detailsString = '<div class="timeline-device-details">'

    var deviceIcon = 'laptop'
    if (ua.device.type) {
        if (ua.device.type == 'tablet' || ua.device.type == 'mobile') {
            deviceIcon = ua.device.type
        }
    }

    var deviceVendor = ''
    if (ua.device.vendor) {
        deviceVendor = ua.device.vendor.toLowerCase()
        if (deviceVendor == 'microsoft') deviceVendor = 'windows'
    }

    var deviceName = 'Unknown'
    if (ua.os.name) {
        deviceName = ua.os.name
        if (deviceName == "Mac OS") {
            deviceVendor = 'apple'
        } else if (deviceName == "Windows") {
            deviceVendor = 'windows'
        }
        if (ua.device.vendor && ua.device.model) {
            deviceName = ua.device.vendor + ' ' + ua.device.model
        }
    }

    if (ua.os.version) {
        deviceName = deviceName + ' (OS Version: ' + ua.os.version + ')'
    }

    deviceString = '<div class="timeline-device-os"><span class="fa fa-stack">' +
        '<i class="fa fa-' + escapeHtml(deviceIcon) + ' fa-stack-2x"></i>' +
        '<i class="fa fa-vendor-icon fa-' + escapeHtml(deviceVendor) + ' fa-stack-1x"></i>' +
        '</span> ' + escapeHtml(deviceName) + '</div>'

    detailsString += deviceString

    var deviceBrowser = 'Unknown'
    var browserIcon = 'info-circle'
    var browserVersion = ''

    
    if (ua.browser && ua.browser.name) {
        deviceBrowser = ua.browser.name
        // Handle the "mobile safari" case
        deviceBrowser = deviceBrowser.replace('Mobile ', '')
        if (deviceBrowser) {
            browserIcon = deviceBrowser.toLowerCase()
            if (browserIcon == 'ie') browserIcon = 'internet-explorer'
        }
        browserVersion = '(Version: ' + ua.browser.version + ')'
    }

    var browserString = '<div class="timeline-device-browser"><span class="fa fa-stack">' +
        '<i class="fa fa-' + escapeHtml(browserIcon) + ' fa-stack-1x"></i></span> ' +
        deviceBrowser + ' ' + browserVersion + '</div>'

    detailsString += browserString
    detailsString += '</div>'
    return detailsString
}

function renderTimeline(data) {

    record = {
        "id": data[0],
        "first_name": data[2],
        "last_name": data[3],
        "email": data[4],
        "position": data[5],
        "status": data[6],
        "reported": data[7],
        "send_date": data[8]
    }
    results = '<div class="timeline col-sm-12 well well-lg">' +
        '<h6>Timeline for ' + escapeHtml(record.first_name) + ' ' + escapeHtml(record.last_name) +
        '</h6><span class="subtitle">Email: ' + escapeHtml(record.email) +
        '<br>Result ID: ' + escapeHtml(record.id) + '</span>' +
        '<div class="timeline-graph col-sm-6">'
    $.each(campaign.timeline, function (i, event) {
        if (!event.email || event.email == record.email) {
            // Add the event
            
            // Handle custom event as a special case
            // We could collapse the first half into the regular code, but for now it feels neater to keep it separate and live with the code re-use (*waves @ jordan*)
            if (event.message == "Custom Event"){

                if (event.details) { // Should always be data, otherwise we can ignore the event
                    details = JSON.parse(event.details)
                    message = details.payload.title

                    results += '<div class="timeline-entry">' +
                    '    <div class="timeline-bar"></div>'
                    results +=
                        '    <div class="timeline-icon ' + statuses[message].label+ '">' +
                        '    <i class="fa ' + statuses[message].icon + '"></i></div>' +
                        '    <div class="timeline-message">' + escapeHtml(message) + // This is the case that makes code reuse tricky, as we want the title of the custom event from the payload. TODO Give some more thought. Perhaps we scrap 'Custom Event' and put the title in the message, and have some other indicator of the requirement to uniquely parse the contents of 'payload' e.g. payload['ae'] = 1 
                        '    <span class="timeline-date">' + moment.utc(event.time).local().format('MMMM Do YYYY h:mm:ss a') + '</span>'

                        // Check if user agent present && requested to display it
                        if ("browser" in details && details.payload.ua == 1){
                            deviceView = renderDevice(details)
                            if (deviceView) {
                                results += deviceView
                            }
                        }

                        //Iterate over sub-items
                        if ("sub_text" in details.payload) {
                            results += '<div class="timeline-device-details">'
                            details.payload.sub_text.forEach(function (text, index) {
                                results += "<div>"
                                // Check if there's an associated icon
                                if ("sub_icon" in details.payload && details.payload.sub_icon.length >= index){
                                    results = results + '<span class="' + details.payload.sub_icon[index] + '"></span> ' //+ text
                                } 
                                results += text
                                results += "</div>"
                            })
                            results += '</div>'
                        }
                    results += '</div></div>'
                } // End custom event processsing

        
            } else { // else, if regular event
             

                results += '<div class="timeline-entry">' +
                    '    <div class="timeline-bar"></div>'
                results +=
                    '    <div class="timeline-icon ' + statuses[event.message].label + '">' +
                    '    <i class="fa ' + statuses[event.message].icon + '"></i></div>' +
                    '    <div class="timeline-message">' + escapeHtml(event.message) +
                    '    <span class="timeline-date">' + moment.utc(event.time).local().format('MMMM Do YYYY h:mm:ss a') + '</span>'
                if (event.details) {
                    details = JSON.parse(event.details)

                    if (event.message == "Clicked Link" || event.message == "Submitted Data" || event.message == "Email Opened") {
                        deviceView = renderDevice(details)
                        if (deviceView) {
                            results += deviceView
                        }
                    }
                    if (event.message == "Submitted Data") {
                        results += '<div class="timeline-replay-button"><button onclick="replay(' + i + ')" class="btn btn-success">'
                        results += '<i class="fa fa-refresh"></i> Replay Credentials</button></div>'
                        results += '<div class="timeline-event-details"><i class="fa fa-caret-right"></i> View Details</div>'
                    }
                    if (details.payload) {
                        results += '<div class="timeline-event-results">'
                        results += '    <table class="table table-condensed table-bordered table-striped">'
                        results += '        <thead><tr><th>Parameter</th><th>Value(s)</tr></thead><tbody>'
                        $.each(Object.keys(details.payload), function (i, param) {
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
            
        }
    })
    // Add the scheduled send event at the bottom
    if (record.status == "Scheduled" || record.status == "Retrying") {
        results += '<div class="timeline-entry">' +
            '    <div class="timeline-bar"></div>'
        results +=
            '    <div class="timeline-icon ' + statuses[record.status].label + '">' +
            '    <i class="fa ' + statuses[record.status].icon + '"></i></div>' +
            '    <div class="timeline-message">' + "Scheduled to send at " + record.send_date + '</span>'
    }
    results += '</div></div>'
    return results
}

var renderTimelineChart = function (chartopts) {
    return Highcharts.chart('timeline_chart', {
        chart: {
            zoomType: 'x',
            type: 'line',
            height: "200px"
        },
        title: {
            text: 'Campaign Timeline'
        },
        xAxis: {
            type: 'datetime',
            dateTimeLabelFormats: {
                second: '%l:%M:%S',
                minute: '%l:%M',
                hour: '%l:%M',
                day: '%b %d, %Y',
                week: '%b %d, %Y',
                month: '%b %Y'
            }
        },
        yAxis: {
            min: 0,
            max: 2,
            visible: false,
            tickInterval: 1,
            labels: {
                enabled: false
            },
            title: {
                text: ""
            }
        },
        tooltip: {
            formatter: function () {
                return Highcharts.dateFormat('%A, %b %d %l:%M:%S %P', new Date(this.x)) +
                    '<br>Event: ' + this.point.message + '<br>Email: <b>' + this.point.email + '</b>'
            }
        },
        legend: {
            enabled: false
        },
        plotOptions: {
            series: {
                marker: {
                    enabled: true,
                    symbol: 'circle',
                    radius: 3
                },
                cursor: 'pointer',
            },
            line: {
                states: {
                    hover: {
                        lineWidth: 1
                    }
                }
            }
        },
        credits: {
            enabled: false
        },
        series: [{
            data: chartopts['data'],
            dashStyle: "shortdash",
            color: "#cccccc",
            lineWidth: 1,
            turboThreshold: 0
        }]
    })
}

/* Renders a pie chart using the provided chartops */
var renderPieChart = function (chartopts) {
    return Highcharts.chart(chartopts['elemId'], {
        chart: {
            type: 'pie',
            events: {
                load: function () {
                    var chart = this,
                        rend = chart.renderer,
                        pie = chart.series[0],
                        left = chart.plotLeft + pie.center[0],
                        top = chart.plotTop + pie.center[1];
                    this.innerText = rend.text(chartopts['data'][0].count, left, top).
                    attr({
                        'text-anchor': 'middle',
                        'font-size': '24px',
                        'font-weight': 'bold',
                        'fill': chartopts['colors'][0],
                        'font-family': 'Helvetica,Arial,sans-serif'
                    }).add();
                },
                render: function () {
                    this.innerText.attr({
                        text: chartopts['data'][0].count
                    })
                }
            }
        },
        title: {
            text: chartopts['title']
        },
        plotOptions: {
            pie: {
                innerSize: '80%',
                dataLabels: {
                    enabled: false
                }
            }
        },
        credits: {
            enabled: false
        },
        tooltip: {
            formatter: function () {
                if (this.key == undefined) {
                    return false
                }
                return '<span style="color:' + this.color + '">\u25CF</span>' + this.point.name + ': <b>' + this.y + '%</b><br/>'
            }
        },
        series: [{
            data: chartopts['data'],
            colors: chartopts['colors'],
        }]
    })
}

/* Updates the bubbles on the map

@param {campaign.result[]} results - The campaign results to process
*/
var updateMap = function (results) {
    if (!map) {
        return
    }
    bubbles = []
    $.each(campaign.results, function (i, result) {
        // Check that it wasn't an internal IP
        if (result.latitude == 0 && result.longitude == 0) {
            return true;
        }
        newIP = true
        $.each(bubbles, function (i, bubble) {
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

/**
 * Creates a status label for use in the results datatable
 * @param {string} status 
 * @param {moment(datetime)} send_date 
 */
function createStatusLabel(status, send_date) {

    if (status in statuses){
        var label = statuses[status].label || "label-default";
    } else {
        var label = "label-default"
    }

    var statusColumn = "<span class=\"label " + label + "\">" + status + "</span>"
    // Add the tooltip if the email is scheduled to be sent
    if (status == "Scheduled" || status == "Retrying") {
        var sendDateMessage = "Scheduled to send at " + send_date
        statusColumn = "<span class=\"label " + label + "\" data-toggle=\"tooltip\" data-placement=\"top\" data-html=\"true\" title=\"" + sendDateMessage + "\">" + status + "</span>"
    }
    return statusColumn
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
        .success(function (c) {
            campaign = c

            updateCustomEventData(campaign, false) // Update data structures with new custom event specifications 

            /* Update the timeline */
            var timeline_series_data = []
            $.each(campaign.timeline, function (i, event) {

                // Handle custom event
                if (event.message == "Custom Event") {
                    details = JSON.parse(event.details)
                    message = details.payload.title

                } else {
                    message = event.message
                    //color = statuses[event.message].color
                }

                var event_date = moment.utc(event.time).local()
                timeline_series_data.push({
                    email: event.email,
                    message: message, //event.message,
                    x: event_date.valueOf(),
                    y: 1,
                    marker: {
                        fillColor: statuses[message].color //statuses[event.message].color
                    }
                })
            })
            var timeline_chart = $("#timeline_chart").highcharts()
            timeline_chart.series[0].update({
                data: timeline_series_data
            })
            /* Update the results donut chart */
            //var email_series_data = {}
            // Load the initial data
            //Object.keys(statusMapping).forEach(function (k) {
            //    email_series_data[k] = 0
            //});
            
            /*
            $.each(campaign.results, function (i, result) {

                // Don't count custom events, we do this independently to avoid backfill logic.
                if (progressListing.includes(result.status)) {
                    email_series_data[result.status]++;
                }

                if (result.reported) {
                    email_series_data['Email Reported']++
                }
                // Backfill status values
                var step = progressListing.indexOf(result.status)
                for (var i = 0; i < step; i++) {
                    email_series_data[progressListing[i]]++
                }
            })*/

            // New function for counting events. Doesn't handle backfill, yet.
            email_series_data = countCampaignEvents(campaign)
 

            $.each(email_series_data, function (status, count) {
                var email_data = []
                if (!(status in statusMapping)) {
                    return true
                }
                email_data.push({
                    name: status,
                    y: Math.floor((count / campaign.results.length) * 100),
                    count: count
                })
                email_data.push({
                    name: '',
                    y: 100 - Math.floor((count / campaign.results.length) * 100)
                })
                var chart = $("#" + statusMapping[status] + "_chart").highcharts()
                chart.series[0].update({
                    data: email_data
                })
            })

            /* Update the datatable */
            resultsTable = $("#resultsTable").DataTable()
            resultsTable.rows().every(function (i, tableLoop, rowLoop) {
                var row = this.row(i)
                var rowData = row.data()
                var rid = rowData[0]
                $.each(campaign.results, function (j, result) {
                    if (result.id == rid) {
                        rowData[8] = moment(result.send_date).format('MMMM Do YYYY, h:mm:ss a')
                        rowData[7] = result.reported
                        rowData[6] = result.status

                        resultsTable.row(i).data(rowData)
                        if (row.child.isShown()) {
                            $(row.node()).find("#caret").removeClass("fa-caret-right")
                            $(row.node()).find("#caret").addClass("fa-caret-down")
                            row.child(renderTimeline(row.data()))
                        }
                        return false
                    }
                })
            })
            resultsTable.draw(false)
            /* Update the map information */
            updateMap(campaign.results)
            $('[data-toggle="tooltip"]').tooltip()
            $("#refresh_message").hide()
            $("#refresh_btn").show()
        })
}

function load() {

    campaign.id = window.location.pathname.split('/').slice(-1)[0]
    var use_map = JSON.parse(localStorage.getItem('gophish.use_map'))
    api.campaignId.results(campaign.id)
        .success(function (c) {
            campaign = c
            if (campaign) {

                updateCustomEventData(campaign, true) // Update data structures with new custom event specifications

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
                // Setup viewing the details of a result
                $("#resultsTable").on("click", ".timeline-event-details", function () {
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
                            "targets": [0, 8]
                        },
                        {
                            "render": function (data, type, row) {
                                return createStatusLabel(data, row[8])
                            },
                            "targets": [6]
                        },
                        {
                            className: "text-center",
                            "render": function (reported, type, row) {
                                if (type == "display") {
                                    if (reported) {
                                        return "<i class='fa fa-check-circle text-center text-success'></i>"
                                    }
                                    return "<i role='button' class='fa fa-times-circle text-center text-muted' onclick='report_mail(\"" + row[0] + "\", \"" + campaign.id + "\");'></i>"
                                }
                                return reported
                            },
                            "targets": [7]
                        }
                    ]
                });
                resultsTable.clear();
                //var email_series_data = {}
                var timeline_series_data = []
                //Object.keys(statusMapping).forEach(function (k) {
                //    email_series_data[k] = 0
                //});
                
                $.each(campaign.results, function (i, result) {

                    resultsTable.row.add([
                        result.id,
                        "<i id=\"caret\" class=\"fa fa-caret-right\"></i>",
                        escapeHtml(result.first_name) || "",
                        escapeHtml(result.last_name) || "",
                        escapeHtml(result.email) || "",
                        escapeHtml(result.position) || "",
                        result.status,
                        result.reported,
                        moment(result.send_date).format('MMMM Do YYYY, h:mm:ss a')
                    ])
                    

                    /*
                    // Don't count custom events, we do this independently to avoid backfill logic.
                    if (progressListing.includes(result.status)) {
                        email_series_data[result.status]++;
                    }

                    if (result.reported) {
                        email_series_data['Email Reported']++
                    }

                    //TODO: At some point need to figure out backfilling with custom events
                    // Possibly just backfill Email sent and Email Opened before getting into more complex
                    // data structures

                    // Backfill status values
                    var step = progressListing.indexOf(result.status)
                    for (var i = 0; i < step; i++) {
                        email_series_data[progressListing[i]]++
                    }
                    */
                    
                })

                // New function for counting events. Doesn't handle backfill, yet.
                email_series_data = countCampaignEvents(campaign)
  
                resultsTable.draw();
                // Setup tooltips
                $('[data-toggle="tooltip"]').tooltip()
                // Setup the individual timelines
                $('#resultsTable tbody').on('click', 'td.details-control', function () {
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
                $.each(campaign.timeline, function (i, event) {
                    if (event.message == "Campaign Created") {
                        return true
                    }
                    var event_date = moment.utc(event.time).local()

                    // Handle custom event
                    if (event.message == "Custom Event") {
                            details = JSON.parse(event.details)
                            message = details.payload.title
                    } else {
                        message = event.message
                    }

                    timeline_series_data.push({
                        email: event.email,
                        message: message, //event.message,
                        x: event_date.valueOf(),
                        y: 1,
                        marker: {
                            fillColor: statuses[message].color //statuses[event.message].color
                        }
                    })
                    
                })
                renderTimelineChart({
                    data: timeline_series_data
                })
                $.each(email_series_data, function (status, count) {
                    var email_data = []
                    if (!(status in statusMapping)) {
                        return true
                    }
                    email_data.push({
                        name: status,
                        y: Math.floor((count / campaign.results.length) * 100),
                        count: count
                    })
                    email_data.push({
                        name: '',
                        y: 100 - Math.floor((count / campaign.results.length) * 100)
                    })

                    var chart = renderPieChart({
                        elemId: statusMapping[status] + '_chart',
                        title: status,
                        name: status,
                        data: email_data,
                        colors: [statuses[status].color, '#dddddd']
                    })

                })

                if (use_map) {
                    $("#resultsMapContainer").show()
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
                updateMap(campaign.results)
            }
        })
        .error(function () {
            $("#loading").hide()
            errorFlash(" Campaign not found!")
        })
}

var setRefresh

function refresh() {
    if (!doPoll) {
        return;
    }
    $("#refresh_message").show()
    $("#refresh_btn").hide()
    poll()
    clearTimeout(setRefresh)
    setRefresh = setTimeout(refresh, 60000)
};

function report_mail(rid, cid) {
    Swal.fire({
        title: "Are you sure?",
        text: "This result will be flagged as reported (RID: " + rid + ")",
        type: "question",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Continue",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        showLoaderOnConfirm: true
    }).then(function (result) {
        if (result.value){
            api.campaignId.get(cid).success((function(c) {
                report_url = c.url;
                report_url += report_url.endsWith("/") ? "" : "/";
                report_url += "report?rid=" + rid;
                $.ajax({
                    url: report_url,
                    method: "GET",
                    success: function(data) {
                        refresh();
                    }
                });
            }));
        }
    })
}


/* updateCustomData will go through the supplied campaign and add custom event data to three data structure:
    statuses
    statusMapping 
    progressListing // Todo, needs more consideration on backfill

    The createPies boolean is used to allow us to create the pies on load() but not re-create them from calling poll(), as the
     highchart info gets overwritten. The problem with this is that if a new custom event comes in while the page is loaded
     the poll() won't add the pie. Need to investigate this. TODO

*/
function updateCustomEventData(campaign, createPies){


    var customEventNames = [] // Hold unique custom event names. Used to create HTML pie charts if customEventPieCharts set to true
    
    campaign.timeline.forEach(function(event) { // Step over each event

        
        if (event.message == "Custom Event") {

            details = JSON.parse(event.details) // TODO Validate this exists

            // 1. Add title, color, icon, and label properties to statuses dict
            title = "Custom Event"
            if ("title" in details.payload){
                title = String(details.payload.title)

            }
            statuses[title] = {"custom event" : 1} // Set true to be custom event, just so we can discern if we need to

            statuses[title]["color"] = "#00FFFF" // Default
            if ("color" in details.payload ){
                color = String(details.payload.color)
                if  (!(/^#[0-9A-F]{6}$/i.test(color))) {
                    color = "#00FFFF" // Default to Cyan if the color is invalid
                }
                statuses[title]["color"] = color
            }

            statuses[title]["icon"] = "fa fa-info" // Default
            if ("icon" in details.payload ){
                icon = String(details.payload.icon)
                statuses[title]["icon"] = icon
            }

            statuses[title]["label"] = "label-info" // Default
            if ("label" in details.payload ){
                label = String(details.payload.label)
                statuses[title]["label"] = label
            }

            if (!customEventNames.includes(title)){
                customEventNames.push(title)
            }

            /* How to handle progressListing needs more thought, and probably  */
            // Add the title to the progressListing array (if it's not already in there)
            //if (!progressListing.includes(title)) {
            //    progressListing.push(title)
            //}

            
        }
    })

    // 2.0 If customEventsPieChart is enabled we add to statusMapping and add HTML charts for the event
    if (customEventsPieCharts == true && createPies == true) {

        //2.1 Create HTML elements

        // Split the array into multiple arrays, each of size 5. This let's us create pie chart rows of five
        customEventNames.sort()
        chunkedArbEvents = Array.from({ length: Math.ceil(customEventNames.length / 5) }, (v, i) => customEventNames.slice(i * 5, i * 5 + 5) );
        
        $("#custompie").html('') // i. Clear the div class
        html = ''
        chunkedArbEvents.forEach(function(chunk){

            rowhtml = '<div class="row">\n\t<div style="height:200px;" class="col-lg-1 col-md-1"></div>\n'
            chunk.forEach(function(title){
                sanitizedEventName = title.toLowerCase().replace(/ /g, "_") // Convert Opened Word Document to opened_word_document.
                sanitizedEventName = escapeHtml(sanitizedEventName) // Should maybe do more tests on this. Or even use a short random string rather than the name. e.g {"Opened Word Document" : "7a2f87"}
                //i. Add the HTML element
                rowhtml += '\t<div id="' + sanitizedEventName + '_chart" style="height:200px;" class="col-lg-2 col-md-2"></div>\n'

                //ii. Add to statusMapping
                statusMapping[title] = sanitizedEventName


            })
            rowhtml += '\t<div style="height:200px;" class="col-lg-1 col-md-1"></div>\n</div>\n'

            html += rowhtml

        })
        $("#custompie").html(html)

    }

}


// countCampaignEvents will return a dict of title:count of custom and regular events from a campaign
// Todo: Need to implement backfill logic
function countCampaignEvents(campaign) {

    // Add all the default events to a counter dict
    eventsCounter = {}
    Object.keys(statusMapping).forEach(function (k) {
        eventsCounter[k] = 0
    });
    
    
    campaign.timeline.forEach(function(event){
        if (event.message == "Custom Event"){
            details = JSON.parse(event.details)
            title = details.payload.title[0]
        } else {
            title = event.message
            // Backfill logic for non custom events. Todo
            
        }
        if (title in eventsCounter) {
            eventsCounter[title] += 1
        } else {
            eventsCounter[title] = 1
        }

        // Backfill logic here for custom? 

    })
    return eventsCounter
}

$(document).ready(function () {
    Highcharts.setOptions({
        global: {
            useUTC: false
        }
    })
    load();

    // Start the polling loop
    setRefresh = setTimeout(refresh, 60000)
})