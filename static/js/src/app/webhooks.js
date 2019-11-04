let webhooks = []

//TODO


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
                    `
                      <div class="pull-right">
                        <button class="btn btn-primary validate_button" data-webhook-id="${webhook.id}">
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
}

const editWebhook = (id) => {
    $("#modalSubmit").unbind('click').click(() => {
        // saveWebhook(id) //TODO
    });
    api.webhookId.get(id)
        .success(function(wh) {
            //TODO
            // $("#username").val(user.username)


        })
        .error(function () {
            errorFlash("Error fetching webhook")
        });
}
const deleteWebhook = (id) => {
    var wh = webhooks.find(x => x.id == id)
    if (!wh) {
        return
    }
    Swal.fire({
        title: "Are you sure?",
        text: `This will delete the webhook '${wh.title}'`,
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
    }).then(function (result) {
        if (result.value){
            Swal.fire(
                'Webhook Deleted!',
                `The webhook '${webhook.title}' has been deleted!`,
                'success'
            );
        }
        $('button:contains("OK")').on('click', function () {
            location.reload()
        })
    })
}

$(document).ready(function() {
    load();
    $("#webhookTable").on('click', '.edit_button', function (e) {
        editWebhook($(this).attr('data-webhook-id'))
    });
    $("#webhookTable").on('click', '.delete_button', function (e) {
        deleteWebhook($(this).attr('data-webhook-id'))
    });
});
