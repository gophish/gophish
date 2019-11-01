let webhooks = []


//TODO
const load = () => {
    $("#loading").show()
    api.webhooks.get()
        .success((us) => {
            $("#loading").hide()
        })
        .error(() => {
            errorFlash("Error fetching webhooks")
        })
}

$(document).ready(function () {
    load()
});
