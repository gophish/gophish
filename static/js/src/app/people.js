function save(e) {
  var t = {};

  t.username = $("#username").val();
  t.email = $("#email").val();
  t.current_password = $("#curpassword").val();
  t.new_password = $("#password").val();
  t.confirm_new_password = $("#confirm_password").val();
  t.api_key = $("#hidden_api_key").val();
  t.id = e;
  t.role = parseInt($("#roles").val());
  t.partner = parseInt($("#partner").val());

  api.userId
    .post(t)
    .success(function(e) {
      successFlash("User updated successfully!"), dismiss();
      location.reload();
    })
    .error(function(e) {
      modalError(e.responseJSON.message);
    });
}

function edit(i) {
  $("#modalSubmit")
    .unbind("click")
    .click(function() {
      save(i);
    });

  api.userId.get(i).success(function(e) {
    $("input[type=text], textarea").val("");

    $("#username").val(e.username);
    $("#email").val(e.email);
    $("#hidden_hash").val(e.hash);
    $("#hidden_uid").val(e.id);
    $("#hidden_api_key").val(e.api_key);

    var partner = e.partner;

    api.roles.get().success(function(r) {
      $("#roles")
        .find("option")
        .each(function(i) {
          if ($(this).val() !== "") {
            $(this).remove();
          }
        });

      $.each(r, function(e, rr) {
        $("#roles").append(
          '<option value="' + rr.rid + '" >' + rr.name + "</option>"
        );
      });
    });

    //populate the roles
    api.rolesByUserId.get(i).success(function(r) {
      $("#roles option").prop("selected", false);
      $("#roles option[value=" + r.rid + "]").prop("selected", true);

      // hide partner field for non-customers and non-child-users
      if (r.rid !== 3 && r.rid !== 4) {
        $("#partner-container").css("display", "none");
      } else {
        $("#partner-container").css("display", "");
      }
    });

    //populate the partners
    api.users.partners().success(function(p) {
      $("#partner")
        .find("option")
        .each(function(i) {
          if ($(this).val() !== "") {
            $(this).remove();
          }
        });

      $.each(p, function(e, pp) {
        var selected = "";
        if (partner == pp.id) {
          selected = 'selected = "selected"';
        } else {
          selected = "";
        }

        $("#partner").append(
          '<option value="' +
            pp.id +
            '"  ' +
            selected +
            ">" +
            pp.username +
            "</option>"
        );
      });
    });

    // conditionally show and hide partner field depending on the selected role
    $("#roles").change(function() {
      if ($(this).val() == 3 || $(this).val() == 4) {
        $("#partner-container").css("display", "");
      } else {
        $("#partner-container").css("display", "none");
      }
    });
  });
}

function deleteUser(e) {
  swal({
    title: "Are you sure?",
    text: "This will delete the user. This can't be undone!",
    type: "warning",
    animation: !1,
    showCancelButton: !0,
    confirmButtonText: "Delete ",
    confirmButtonColor: "#428bca",
    reverseButtons: !0,
    allowOutsideClick: !1,
    preConfirm: function() {
      return new Promise(function(a, t) {
        api.userId
          .delete(e)
          .success(function(e) {
            a();
          })
          .error(function(e) {
            t(e.responseJSON.message);
          });
      });
    }
  }).then(function() {
    swal("User Deleted!", "This user has been deleted!", "success"),
      $('button:contains("OK")').on("click", function() {
        location.reload();
      });
  });
}

function dismiss() {
  $("#modal\\.flashes").empty(),
    $("#username").val(""),
    $("#email").val(""),
    $("#curpassword").val(""),
    $("#password").val(""),
    $("#confirm_password").val(""),
    $("#partner").val(""),
    $("#roles").val(""),
    $("#modal").modal("hide");
}

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
  }),
    $("#send_by_date").datetimepicker({
      widgetPositioning: {
        vertical: "bottom"
      },
      showTodayButton: !0,
      useCurrent: !1
    }),
    $(".modal").on("hidden.bs.modal", function(e) {
      $(this).removeClass("fv-modal-stack"),
        $("body").data("fv_open_modals", $("body").data("fv_open_modals") - 1);
    }),
    $(".modal").on("shown.bs.modal", function(e) {
      void 0 === $("body").data("fv_open_modals") &&
        $("body").data("fv_open_modals", 0),
        $(this).hasClass("fv-modal-stack") ||
          ($(this).addClass("fv-modal-stack"),
          $("body").data(
            "fv_open_modals",
            $("body").data("fv_open_modals") + 1
          ),
          $(this).css("z-index", 1040 + 10 * $("body").data("fv_open_modals")),
          $(".modal-backdrop")
            .not(".fv-modal-stack")
            .css("z-index", 1039 + 10 * $("body").data("fv_open_modals")),
          $(".modal-backdrop")
            .not("fv-modal-stack")
            .addClass("fv-modal-stack"));
    }),
    $(document).on("hidden.bs.modal", ".modal", function() {
      $(".modal:visible").length && $(document.body).addClass("modal-open");
    }),
    $("#modal").on("hidden.bs.modal", function(e) {
      dismiss();
    }),
    api.users
      .get()
      .success(function(e) {
        (people = e),
          $("#loading").hide(),
          people.length > 0
            ? ($("#peopleTable").show(),
              (peopleTable = $("#peopleTable").DataTable({
                columnDefs: [
                  {
                    orderable: !1,
                    targets: "no-sort"
                  }
                ],
                order: [[0, "asc"]]
              })),
              $.each(people, function(e, a) {
                // label = labels[a.status] || "label-default";
                // var t;
                // if (moment(a.launch_date).isAfter(moment())) {
                //     t = "Scheduled to start: " + moment(a.launch_date).format("MMMM Do YYYY, h:mm:ss a");
                //     var n = t + "<br><br>Number of recipients: " + a.stats.total
                // } else {
                //     t = "Launch Date: " + moment(a.launch_date).format("MMMM Do YYYY, h:mm:ss a");
                //     var n = t + "<br><br>Number of recipients: " + a.stats.total + "<br><br>Emails opened: " + a.stats.opened + "<br><br>Emails clicked: " + a.stats.clicked + "<br><br>Submitted Credentials: " + a.stats.submitted_data + "<br><br>Errors : " + a.stats.error + "Reported : " + a.stats.reported
                // }

                peopleTable.row
                  .add([
                    a.username,
                    a.email,
                    a.role,
                    "<div class='pull-right'><span data-toggle='modal' data-backdrop='static' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='' onclick='edit(" +
                      a.id +
                      ")' data-original-title='Edit Page'>  <i class='fa fa-pencil'></i> </button> </span>  <span data-backdrop='static' data-target='#modal'><button class='btn btn-danger' onclick='deleteUser(" +
                      a.id +
                      ")' data-toggle='tooltip' data-placement='left' title='Delete User'> <i class='fa fa-trash-o'></i></button></span></div>"
                  ])
                  .draw();
              }))
            : $("#emptyMessage").show();
      })
      .error(function() {
        $("#loading").hide(), errorFlash("Error fetching peoples");
      }),
    $.fn.select2.defaults.set("width", "100%"),
    $.fn.select2.defaults.set("dropdownParent", $("#modal_body")),
    $.fn.select2.defaults.set("theme", "bootstrap"),
    $.fn.select2.defaults.set("sorter", function(e) {
      return e.sort(function(e, a) {
        return e.text.toLowerCase() > a.text.toLowerCase()
          ? 1
          : e.text.toLowerCase() < a.text.toLowerCase()
          ? -1
          : 0;
      });
    });
});
