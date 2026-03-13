var map = null
var doPoll = true;

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
var rangeView = {
    enabled: false,
    source: "dashboard",
    mode: "snapshot",
    start: null,
    end: null
}

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
            results += '<div class="timeline-entry">' +
                '    <div class="timeline-bar"></div>'
            results +=
                '    <div class="timeline-icon ' + statuses[event.message].label + '">' +
                '    <i class="fa ' + statuses[event.message].icon + '"></i></div>' +
                '    <div class="timeline-message">' + escapeHtml(event.message) +
                '    <span class="timeline-date">' + moment.utc(event.time).local().format('MMMM Do YYYY h:mm:ss a') + '</span>'
            if (event.details) {
                details = JSON.parse(event.details)
                if (event.message == "Clicked Link" || event.message == "Submitted Data") {
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

function buildDashboardStatsFromCampaign(results) {
    var stats = {
        total: results.length,
        sent: 0,
        opened: 0,
        clicked: 0,
        submitted_data: 0,
        email_reported: 0,
        error: 0
    }
    $.each(results, function (i, result) {
        switch (result.status) {
            case "Submitted Data":
                stats.submitted_data++
                stats.clicked++
                stats.opened++
                stats.sent++
                break
            case "Clicked Link":
                stats.clicked++
                stats.opened++
                stats.sent++
                break
            case "Email Opened":
                stats.opened++
                stats.sent++
                break
            case "Email Sent":
                stats.sent++
                break
            case "Error":
                stats.error++
                break
        }
        if (result.reported) {
            stats.email_reported++
        }
    })
    return stats
}

function getChartCountsFromStats(stats) {
    return {
        "Email Sent": stats.sent || 0,
        "Email Opened": stats.opened || 0,
        "Clicked Link": stats.clicked || 0,
        "Submitted Data": stats.submitted_data || 0,
        "Email Reported": stats.email_reported || 0
    }
}

function getPieChartData(count, total, status) {
    var safeTotal = total || 0
    var pct = 0
    if (safeTotal > 0) {
        pct = Math.floor((count / safeTotal) * 100)
    }
    return [{
        name: status,
        y: pct,
        count: count
    }, {
        name: '',
        y: 100 - pct
    }]
}

function updatePieChartsFromStats(stats) {
    var chartCounts = getChartCountsFromStats(stats)
    var total = stats.total || campaign.results.length
    $.each(chartCounts, function (status, count) {
        var chart = $("#" + statusMapping[status] + "_chart").highcharts()
        if (!chart) {
            return true
        }
        chart.series[0].update({
            data: getPieChartData(count, total, status)
        })
    })
}

function eventInSelectedWindow(eventMoment, startMoment, endMoment) {
    if (startMoment && eventMoment.isBefore(startMoment)) {
        return false
    }
    if (endMoment && eventMoment.isAfter(endMoment)) {
        return false
    }
    return true
}

function buildTimelineSeries(startMoment, endMoment) {
    var timelineSeriesData = []
    $.each(campaign.timeline, function (i, event) {
        if (event.message == "Campaign Created") {
            return true
        }
        var eventDate = moment.utc(event.time).local()
        if (!eventInSelectedWindow(eventDate, startMoment, endMoment)) {
            return true
        }
        timelineSeriesData.push({
            email: event.email,
            message: event.message,
            x: eventDate.valueOf(),
            y: 1,
            marker: {
                fillColor: statuses[event.message].color
            }
        })
    })
    return timelineSeriesData
}

function updateTimelineChartWindow(startMoment, endMoment) {
    var timelineChart = $("#timeline_chart").highcharts()
    if (!timelineChart) {
        return
    }
    timelineChart.series[0].update({
        data: buildTimelineSeries(startMoment, endMoment)
    })
}

function formatSummaryDate(value) {
    if (!value) {
        return ""
    }
    return moment.utc(value).local().format('MMMM Do YYYY, h:mm:ss a')
}

function renderStatsSummary(response) {
    if ($("#range_view_source").val() !== "actual") {
        $("#range_stats_summary").hide().empty()
        return
    }
    var modeLabel = response.mode == "range" ? "First reached within range" : "Snapshot at end"
    var windowLabel = "End: " + formatSummaryDate(response.end_date)
    if (response.mode == "range") {
        windowLabel = "Start: " + formatSummaryDate(response.start_date) + "<br>End: " + formatSummaryDate(response.end_date)
    }
    var actual = response.actual
    var dashboard = response.dashboard
    var html = "<strong>Historical Events View</strong><br>"
    html += "Mode: " + modeLabel + "<br>"
    html += windowLabel + "<br><br>"
    html += "Charts now show event-derived counts. Dashboard counts are shown below for comparison.<br><br>"
    html += "<table class=\"table table-condensed table-bordered\" style=\"margin-bottom: 0; background: #fff;\">"
    html += "<thead><tr><th>Source</th><th>Sent</th><th>Opened</th><th>Clicked</th><th>Submitted</th><th>Reported</th><th>Error</th></tr></thead>"
    html += "<tbody>"
    html += "<tr><td>Historical Events</td><td>" + actual.sent + "</td><td>" + actual.opened + "</td><td>" + actual.clicked + "</td><td>" + actual.submitted_data + "</td><td>" + actual.email_reported + "</td><td>" + actual.error + "</td></tr>"
    html += "<tr><td>Current Dashboard</td><td>" + dashboard.sent + "</td><td>" + dashboard.opened + "</td><td>" + dashboard.clicked + "</td><td>" + dashboard.submitted_data + "</td><td>" + dashboard.email_reported + "</td><td>" + dashboard.error + "</td></tr>"
    html += "</tbody></table>"
    $("#range_stats_summary").html(html).show()
}

function setDatePickerEnabled(pickerSelector, enabled) {
    var picker = $(pickerSelector).data("DateTimePicker")
    if (!picker) {
        return
    }
    if (enabled) {
        picker.enable()
    } else {
        picker.disable()
    }
}

function syncRangeControls() {
    var source = $("#range_view_source").val()
    var mode = $("#range_view_mode").val()
    var useHistorical = source == "actual"
    $("#range_view_mode").prop("disabled", !useHistorical)
    setDatePickerEnabled("#range_end_picker", useHistorical)
    setDatePickerEnabled("#range_start_picker", useHistorical && mode == "range")
}

function readRangeControlState() {
    var source = $("#range_view_source").val()
    var mode = $("#range_view_mode").val()
    var startDate = $("#range_start_picker").data("DateTimePicker").date()
    var endDate = $("#range_end_picker").data("DateTimePicker").date()
    return {
        source: source,
        mode: mode,
        start: startDate,
        end: endDate
    }
}

function applyDashboardView() {
    rangeView.enabled = false
    rangeView.source = "dashboard"
    rangeView.mode = "snapshot"
    rangeView.start = null
    rangeView.end = null
    updatePieChartsFromStats(buildDashboardStatsFromCampaign(campaign.results))
    updateTimelineChartWindow(null, null)
    $("#range_stats_summary").hide().empty()
}

function applyHistoricalView(triggerFlash) {
    var state = readRangeControlState()
    if (state.source != "actual") {
        applyDashboardView()
        return
    }
    if (!state.end) {
        if (triggerFlash !== false) {
            errorFlash("End date is required for historical analysis")
        }
        return
    }
    if (state.mode == "range" && !state.start) {
        if (triggerFlash !== false) {
            errorFlash("Start date is required for range analysis")
        }
        return
    }
    if (state.start && state.end && state.start.isAfter(state.end)) {
        if (triggerFlash !== false) {
            errorFlash("Start date must be before end date")
        }
        return
    }
    var params = {
        mode: state.mode,
        end: state.end.clone().utc().toISOString()
    }
    if (state.mode == "range" && state.start) {
        params.start = state.start.clone().utc().toISOString()
    }
    api.campaignId.rangeStats(campaign.id, params)
        .success(function (response) {
            rangeView.enabled = true
            rangeView.source = state.source
            rangeView.mode = state.mode
            rangeView.start = state.start ? state.start.clone() : null
            rangeView.end = state.end.clone()
            updatePieChartsFromStats(response.actual)
            updateTimelineChartWindow(rangeView.mode == "range" ? rangeView.start : null, rangeView.end)
            renderStatsSummary(response)
        })
        .error(function (data) {
            if (triggerFlash !== false) {
                errorFlash(data.responseJSON.message || "Error loading historical analysis")
            }
        })
}

function isZeroCampaignDate(raw) {
    if (!raw) {
        return true
    }
    var candidate = moment.utc(raw)
    return !candidate.isValid() || candidate.year() <= 1
}

function initializeRangeControls() {
    if ($("#range_start_picker").data("DateTimePicker")) {
        return
    }
    $("#range_start_picker").datetimepicker({
        widgetPositioning: {
            vertical: "bottom"
        },
        showTodayButton: true,
        useCurrent: false,
        format: "MMMM Do YYYY, h:mm:ss a"
    })
    $("#range_end_picker").datetimepicker({
        widgetPositioning: {
            vertical: "bottom"
        },
        showTodayButton: true,
        useCurrent: false,
        format: "MMMM Do YYYY, h:mm:ss a"
    })
    $("#range_view_source").on("change", function () {
        syncRangeControls()
    })
    $("#range_view_mode").on("change", function () {
        syncRangeControls()
    })
    $("#range_apply").on("click", function () {
        applyHistoricalView()
    })
    $("#range_reset").on("click", function () {
        $("#range_view_source").val("dashboard")
        $("#range_view_mode").val("snapshot")
        syncRangeControls()
        applyDashboardView()
    })
}

function setRangeControlDefaults() {
    var defaultStart = isZeroCampaignDate(campaign.launch_date) ? moment() : moment.utc(campaign.launch_date).local()
    var defaultEnd = moment()
    if (!isZeroCampaignDate(campaign.completed_date)) {
        defaultEnd = moment.utc(campaign.completed_date).local()
    }
    $("#range_view_source").val("dashboard")
    $("#range_view_mode").val("snapshot")
    $("#range_start_picker").data("DateTimePicker").date(defaultStart)
    $("#range_end_picker").data("DateTimePicker").date(defaultEnd)
    syncRangeControls()
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
    var label = statuses[status].label || "label-default";
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
            if (rangeView.enabled && rangeView.source == "actual") {
                applyHistoricalView(false)
            } else {
                updatePieChartsFromStats(buildDashboardStatsFromCampaign(campaign.results))
                updateTimelineChartWindow(null, null)
            }
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
	                $("title").text(c.name + " - Gophish")
	                $("#loading").hide()
	                $("#campaignResults").show()
                    initializeRangeControls()
                    setRangeControlDefaults()
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
	                })
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
	                renderTimelineChart({
	                    data: buildTimelineSeries(null, null)
	                })
                    var dashboardStats = buildDashboardStatsFromCampaign(campaign.results)
                    var chartCounts = getChartCountsFromStats(dashboardStats)
                    $.each(chartCounts, function (status, count) {
                        if (!(status in statusMapping)) {
                            return true
                        }
                        renderPieChart({
                            elemId: statusMapping[status] + '_chart',
                            title: status,
                            name: status,
                            data: getPieChartData(count, dashboardStats.total, status),
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
                report_url = new URL(c.url)
                report_url.pathname = '/report'
                report_url.search = "?rid=" + rid 
                fetch(report_url)
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! Status: ${response.status}`);
                    }
                    refresh();
                })
                .catch(error => {
                    let errorMessage = error.message;
                    if (error.message === "Failed to fetch") {
                        errorMessage = "This might be due to Mixed Content issues or network problems.";
                    }
                    Swal.fire({
                        title: 'Error',
                        text: errorMessage,
                        type: 'error',
                        confirmButtonText: 'Close'
                    });
                });
            }));
        }
    })
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
