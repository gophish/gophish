$(document).ready(function() {
    $("#apiResetForm").submit(function(e) {
        return api.reset().success(function(e) {
            user.api_key = e.data, successFlash(e.message), $("#api_key").val(user.api_key)
        }).error(function(e) {
            errorFlash(e.message)
        }), !1
    }), $("#settingsForm").submit(function(e) {
        console.log($(this).serialize())
        return $.post("/settings", $(this).serialize()).done(function(e) {
            successFlash(e.message)
        }).fail(function(e) {
            errorFlash(e.responseJSON.message)
        }), !1
    });
    var e = localStorage.getItem("gophish.use_map");
    $("#use_map").prop("checked", JSON.parse(e)), $("#use_map").on("change", function() {
        localStorage.setItem("gophish.use_map", JSON.stringify(this.checked))
    })
});