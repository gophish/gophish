

var labels = {
        "In progress": "label-primary",
        Queued: "label-info",
        Completed: "label-success",
        "Emails Sent": "label-success",
        Error: "label-danger"
    },
    campaigns = [],
    campaign = {};

$(document).ready(function() {
    $("#launch_date").datetimepicker({
        widgetPositioning: {
            vertical: "bottom"
        },
        showTodayButton: !0,
        defaultDate: moment()
    }), $("#send_by_date").datetimepicker({
        widgetPositioning: {
            vertical: "bottom"
        },
        showTodayButton: !0,
        useCurrent: !1
    }), $(".modal").on("hidden.bs.modal", function(e) {
        $(this).removeClass("fv-modal-stack"), $("body").data("fv_open_modals", $("body").data("fv_open_modals") - 1)
    }), $(".modal").on("shown.bs.modal", function(e) {
        void 0 === $("body").data("fv_open_modals") && $("body").data("fv_open_modals", 0), $(this).hasClass("fv-modal-stack") || ($(this).addClass("fv-modal-stack"), $("body").data("fv_open_modals", $("body").data("fv_open_modals") + 1), $(this).css("z-index", 1040 + 10 * $("body").data("fv_open_modals")), $(".modal-backdrop").not(".fv-modal-stack").css("z-index", 1039 + 10 * $("body").data("fv_open_modals")), $(".modal-backdrop").not("fv-modal-stack").addClass("fv-modal-stack"))
    }), $(document).on("hidden.bs.modal", ".modal", function() {
        $(".modal:visible").length && $(document.body).addClass("modal-open")
    }), $("#modal").on("hidden.bs.modal", function(e) {
        dismiss()
    }), api.roles.get().success(function(e) {

        console.log(e)

        roles = e, $("#loading").hide(), roles.length > 0 ? ($("#rolesTable").show(), peopleTable = $("#rolesTable").DataTable({
            columnDefs: [{
                orderable: !1,
                targets: "no-sort"
            }],
            order: [
                [0, "asc"]
            ]
        }), 

        $.each(roles, function(e, a) {
            console.log(a);
            // label = labels[a.status] || "label-default";
            // var t;
            // if (moment(a.launch_date).isAfter(moment())) {
            //     t = "Scheduled to start: " + moment(a.launch_date).format("MMMM Do YYYY, h:mm:ss a");
            //     var n = t + "<br><br>Number of recipients: " + a.stats.total
            // } else {
            //     t = "Launch Date: " + moment(a.launch_date).format("MMMM Do YYYY, h:mm:ss a");
            //     var n = t + "<br><br>Number of recipients: " + a.stats.total + "<br><br>Emails opened: " + a.stats.opened + "<br><br>Emails clicked: " + a.stats.clicked + "<br><br>Submitted Credentials: " + a.stats.submitted_data + "<br><br>Errors : " + a.stats.error + "Reported : " + a.stats.reported
            // }
            peopleTable.row.add([a.name, a.weight]).draw() 
        })) : $("#emptyMessage").show()
    }).error(function() {
        $("#loading").hide(), errorFlash("Error fetching peoples")
    }), $.fn.select2.defaults.set("width", "100%"), $.fn.select2.defaults.set("dropdownParent", $("#modal_body")), $.fn.select2.defaults.set("theme", "bootstrap"), $.fn.select2.defaults.set("sorter", function(e) {
        return e.sort(function(e, a) {
            return e.text.toLowerCase() > a.text.toLowerCase() ? 1 : e.text.toLowerCase() < a.text.toLowerCase() ? -1 : 0
        })
    })
});