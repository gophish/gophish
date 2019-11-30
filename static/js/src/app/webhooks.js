let webhooks = [];

const dismiss = () => {
    $("#title").val("");
    $("#url").val("");
    $("#secret").val("");
    $("#flashes").empty();
}

const saveWebhook = (id) => {
    let wh = {
        title: $("#title").val(),
        url: $("#url").val(),
        secret: $("#secret").val()
    };
    if (id != -1) {
        wh.id = id;
        api.webhookId.put(wh)
            .success(function(data) {
                successFlash(`Webhook "${wh.title}" has been updated successfully!`);
                load();
                dismiss();
                $("#modal").modal("hide");
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    } else {
        api.webhooks.post(wh)
            .success(function(data) {
                successFlash(`Webhook "${wh.title}" has been created successfully!`);
                load();
                dismiss();
                $("#modal").modal("hide");
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    }
};

const load = () => {
    $("#webhookTable").hide();
    $("#loading").show();
    api.webhooks.get()
        .success((whs) => {
            webhooks = whs;
            $("#loading").hide()
            $("#webhookTable").show()
            let webhookTable = $("#webhookTable").DataTable({
                destroy: true,
                columnDefs: [{
                    orderable: false,
                    targets: "no-sort"
                }]
            });
            webhookTable.clear();
            $.each(webhooks, (i, webhook) => {
                webhookTable.row.add([
                    escapeHtml(webhook.title),
                    escapeHtml(webhook.url),
                    escapeHtml(webhook.is_active),
                    `
                      <div class="pull-right">
                        <button class="btn btn-primary ping_button" data-webhook-id="${webhook.id}">
                          Ping
                        </button>
                        <button class="btn btn-primary edit_button" data-toggle="modal" data-backdrop="static" data-target="#modal" data-webhook-id="${webhook.id}">
                          <i class="fa fa-pencil"></i>
                        </button>
                        <button class="btn btn-danger delete_button" data-webhook-id="${webhook.id}">
                          <i class="fa fa-trash-o"></i>
                        </button>
                      </div>
                    `
                ]).draw()
            })
        })
        .error(() => {
            errorFlash("Error fetching webhooks")
        })
};

const editWebhook = (id) => {
    $("#modalSubmit").unbind("click").click(() => {
        saveWebhook(id);
    });
    if (id !== -1) {
        api.webhookId.get(id)
          .success(function(wh) {
              //TODO
              $("#title").val(wh.title);
              $("#url").val(wh.url);
              $("#secret").val(wh.secret);
          })
          .error(function () {
              errorFlash("Error fetching webhook")
          });
    }
};

const deleteWebhook = (id) => {
    var wh = webhooks.find(x => x.id == id);
    if (!wh) {
        return
    }
    Swal.fire({
        title: "Are you sure?",
        text: `This will delete the webhook titled "${wh.title}"`,
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function () {
            return new Promise((resolve, reject) => {
                api.webhookId.delete(id)
                    .success((msg) => {
                        resolve()
                    })
                    .error((data) => {
                        reject(data.responseJSON.message)
                    })
            })
            .catch(error => {
                Swal.showValidationMessage(error)
              })
        }
    }).then(function(result) {
        if (result.value) {
            Swal.fire(
                "Webhook Deleted!",
                `The webhook has been deleted!`,
                "success"
            );
        }
        $("button:contains('OK')").on("click", function() {
            location.reload();
        })
    })
};

const pingUrl = (btn, whId) => {
    dismiss();
    btn.disabled = true;
    api.webhookId.ping(whId)
        .success(function(wh) {
            btn.disabled = false;
            successFlash(`Ping of "${wh.title}" webhook succeeded, please reload the page to fetch the updated data`);
        })
        .error(function() {
            btn.disabled = false;
            var wh = webhooks.find(x => x.id == whId);
            if (!wh) {
                return
            }
            errorFlash(`Ping of "${wh.title}" webhook failed`)
        });
};

$(document).ready(function() {
    load();
    $("#modal").on("hide.bs.modal", function() {
        dismiss();
    });
    $("#new_button").on("click", function() {
        editWebhook(-1);
    });
    $("#webhookTable").on("click", ".edit_button", function(e) {
        editWebhook($(this).attr("data-webhook-id"));
    });
    $("#webhookTable").on("click", ".delete_button", function(e) {
        deleteWebhook($(this).attr("data-webhook-id"));
    });
    $("#webhookTable").on("click", ".ping_button", function(e) {
        pingUrl(e.currentTarget, e.currentTarget.dataset.webhookId);
    });
});
