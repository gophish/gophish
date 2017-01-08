$(document).ready(function() {
    $("#apiResetForm").submit(function(e) {
        $.post("/api/reset", $(this).serialize())
            .done(function(data) {
                api_key = data.data
                successFlash(data.message)
                $("#api_key").val(api_key)
            })
            .fail(function(data) {
                errorFlash(data.message)
            })
        return false
    })
    $("#settingsForm").submit(function(e) {
        $.post("/settings", $(this).serialize())
            .done(function(data) {
                successFlash(data.message)
            })
            .fail(function(data) {
                errorFlash(data.responseJSON.message)
            })
        return false
    })
})
