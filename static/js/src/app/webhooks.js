//TODO

let webhooks = []
const load = () => {
    $("#webhookTable").hide()
    $("#loading").show()
    api.webhooks.get()
        .success((wh) => {
            webhooks = wh;
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
                        <button class="btn btn-primary validate_button" data-backdrop="static" data-webhook-id="${webhook.id}">
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

$("#apiTestWebhookForm").submit(function(e) {
    api.webhookId.validate("TODO - id")
        .success(function(response) {
            successFlash(response.message)
        })
        .error(function (data) {
            errorFlash(data.message)
        })
    return false
})

$("#webhooksForm").submit(function(e) {
    
})

$(document).ready(function() {
    load()
});
