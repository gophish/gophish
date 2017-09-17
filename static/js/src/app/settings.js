$(document).ready(function () {
    $("#apiResetForm").submit(function (e) {
        $.post("/api/reset", $(this).serialize())
            .done(function (data) {
                api_key = data.data
                successFlash(data.message)
                $("#api_key").val(api_key)
            })
            .fail(function (data) {
                errorFlash(data.message)
            })
        return false
    })
    $("#settingsForm").submit(function (e) {
        $.post("/settings", $(this).serialize())
            .done(function (data) {
                successFlash(data.message)
            })
            .fail(function (data) {
                errorFlash(data.responseJSON.message)
            })
        return false
    })
    var use_map = localStorage.getItem('gophish.use_map')
    $("#use_map").prop('checked', JSON.parse(use_map))
    $("#use_map").on('change', function () {
        localStorage.setItem('gophish.use_map', JSON.stringify(this.checked))
    })
})