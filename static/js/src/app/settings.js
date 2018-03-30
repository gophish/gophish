$(document).ready(function () {
    $("#apiResetForm").submit(function (e) {
        api.reset()
            .success(function (response) {
                user.api_key = response.data
                successFlash(response.message)
                $("#api_key").val(user.api_key)
            })
            .error(function (data) {
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