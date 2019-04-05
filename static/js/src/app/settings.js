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

    var hide_table = localStorage.getItem('gophish.hide_table')
    toggleSettings(hide_table=="false")
    $("#hide_table").prop('checked', JSON.parse(hide_table))
    $("#hide_table").on('change', function () {
        localStorage.setItem('gophish.hide_table', JSON.stringify(this.checked))
        toggleSettings(JSON.stringify(this.checked)=="false")
    })

    var show_scheduled = localStorage.getItem('gophish.show_scheduled')
    $("#show_scheduled").prop('checked', JSON.parse(show_scheduled))
    $("#show_scheduled").on('change', function () {
        localStorage.setItem('gophish.show_scheduled', JSON.stringify(this.checked))
    })

    var show_sending = localStorage.getItem('gophish.show_sending')
    $("#show_sending").prop('checked', JSON.parse(show_sending))
    $("#show_sending").on('change', function () {
        localStorage.setItem('gophish.show_sending', JSON.stringify(this.checked))
    })

    var show_email_opened = localStorage.getItem('gophish.show_email_opened')
    $("#show_email_opened").prop('checked', JSON.parse(show_email_opened))
    $("#show_email_opened").on('change', function () {
        localStorage.setItem('gophish.show_email_opened', JSON.stringify(this.checked))
    })

    var show_clicked_link = localStorage.getItem('gophish.show_clicked_link')
    $("#show_clicked_link").prop('checked', JSON.parse(show_clicked_link))
    $("#show_clicked_link").on('change', function () {
        localStorage.setItem('gophish.show_clicked_link', JSON.stringify(this.checked))
    })

    var show_submitted_data = localStorage.getItem('gophish.show_submitted_data')
    $("#show_submitted_data").prop('checked', JSON.parse(show_submitted_data))
    $("#show_submitted_data").on('change', function () {
        localStorage.setItem('gophish.show_submitted_data', JSON.stringify(this.checked))
    })

    var show_error = localStorage.getItem('gophish.show_error')
    $("#show_error").prop('checked', JSON.parse(show_error))
    $("#show_error").on('change', function () {
        localStorage.setItem('gophish.show_error', JSON.stringify(this.checked))
    })
})

function toggleSettings(position){
  if(position){
    $("#additonalOptions").css("pointer-events","none")
    $("#additonalOptions").css("opacity",".4")
  }
  else{
    $("#additonalOptions").css("pointer-events","")
    $("#additonalOptions").css("opacity","1")
  }
}
