//TODO

let webhooks = []
const load = () => {
    $("#loading").show()
    api.webhooks.get()
        .success((wh) => {
            $("#loading").hide()
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
