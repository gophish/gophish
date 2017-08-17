// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
        color: "#1abc9c",
        label: "label-success",
        order: 3,
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "In progress": {
        label: "label-primary",
        order: 4
    },
    "Queued": {
        label: "label-info",
        order: 4
    },
    "Completed": {
        label: "label-success",
        order: 4
    },
    "Email Opened": {
        color: "#f9bf3b",
        label: "label-warning",
        order: 2,
        icon: "fa-envelope",
        point: "ct-point-opened"
    },
    "Clicked Link": {
        color: "#F39C12",
        label: "label-clicked",
        order: 1,
        icon: "fa-mouse-pointer",
        point: "ct-point-clicked"
    },
    "Success": {
        color: "#f05b4f",
        label: "label-danger",
        order: 4,
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Error": {
        color: "#6c7a89",
        label: "label-default",
        order: 5,
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Error Sending Email": {
        color: "#6c7a89",
        label: "label-default",
        order: 5,
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Submitted Data": {
        color: "#f05b4f",
        label: "label-danger",
        order: 0,
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Unknown": {
        color: "#6c7a89",
        label: "label-default",
        order: 5,
        icon: "fa-question",
        point: "ct-point-error"
    },
    "Sending": {
        color: "#428bca",
        label: "label-primary",
        order: 4,
        icon: "fa-spinner",
        point: "ct-point-sending"
    },
    "Campaign Created": {
        label: "label-success",
        order: 4,
        icon: "fa-rocket"
    }
}

function load() {
    $("#targetTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.campaigns.get()
        .success(function(response) {
            $("#loading").hide()
            if (response.length > 0) {
                campaigns = response
                targets = []
                $.each(campaigns, function(i, c) {
                    console.log(c)
                    $.each(c.results, function(j, result) {
                        targets.push({
                            status: result.status,
                            first_name: result.first_name,
                            last_name: result.last_name,
                            email: result.email,
                            campaign: {
                                id: c.id,
                                name: c.name
                            }
                        })
                    })
                })
                console.log(targets);
                $("#emptyMessage").hide()
                $("#targetTable").show()
                var targetTable = $("#targetTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                targetTable.clear();
                $.each(targets, function(i, target) {
                    label = statuses[target.status].label || "label-default";
                    order = statuses[target.status].order;
                    targetTable.row.add([
                        "<span class=\"label " + label + "\"><p style=\"display:none\">" + order + "</p>" + target.status + "</span>",
                        escapeHtml(target.first_name),
                        escapeHtml(target.last_name),
                        escapeHtml(target.email),
                        escapeHtml(target.campaign.name),
                        "<div class='pull-right'><a class='btn btn-primary' href='/campaigns/" + target.campaign.id + "' data-toggle='tooltip' data-placement='left' title='View Results'>\
                    <i class='fa fa-bar-chart'></i>\
                    </a></div"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function() {
    load()
});
