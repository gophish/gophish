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
    $("#imapForm").submit(function (e) {
        var imapSettings = {}
        imapSettings.host = $("#imaphost").val()
        imapSettings.port = $("#imapport").val()
        imapSettings.username = $("#imapusername").val()
        imapSettings.password = $("#imappassword").val()
        imapSettings.enabled = $('#use_imap').prop('checked')
        imapSettings.tls = $('#use_tls').prop('checked')

        //To avoid unmarshalling error in controllers/api/imap.go. It would fail gracefully, but with a generic error. 
        if (imapSettings.port == ""){
            errorFlash("No IMAP Host specified")
            document.body.scrollTop = 0;
            document.documentElement.scrollTop = 0;
            return false
        }
        if (isNaN(imapSettings.port) || imapSettings.port <1 || imapSettings.port > 65535  ){ 
            errorFlash("Invalid IMAP Port")
            document.body.scrollTop = 0;
            document.documentElement.scrollTop = 0;
            return false
        }

        //api.IMAP.post(imapSettings).done(function (data) { // When using this API approach we get an error in the logs "http: TLS handshake error from 127.0.0.1:53858: remote error: tls: unknown certificate"
        query("/imap/", "POST", imapSettings, true).done(function (data) { //  so using this direct query() approach for now
                if (data.success == true) {
                    successFlashFade("Successfully updated IMAP settings.", 2)
                } else {
                    errorFlash("Unable to update IMAP settings.")
                }
            })
            .fail(function (data) {
                errorFlash(data.responseJSON.message)
            })
            .always(function (data){
                document.body.scrollTop = 0;
                document.documentElement.scrollTop = 0;
            })
        return false
    })

    $("#testimap").click(function() {
        var oldHTML = $("#testimap").html();
        // Disable inputs and change button text
        $("#imaphost").attr("disabled", true);
        $("#imapport").attr("disabled", true);
        $("#imapusername").attr("disabled", true);
        $("#imappassword").attr("disabled", true);
        $("#use_imap").attr("disabled", true);
        $("#use_tls").attr("disabled", true);
        $("#testimap").attr("disabled", true);
        $("#testimap").html("<i class='fa fa-circle-o-notch fa-spin'></i> Testing...");

        // Query test imap server endpoint
        var server = {}
        server.host = $("#imaphost").val()
        server.port = $("#imapport").val()
        server.username = $("#imapusername").val()
        server.password = $("#imappassword").val()
        server.tls = $('#use_tls').prop('checked')
        
        //api.IMAP.test(server).done(function() { // When using this API approach the button text does not change, and the inputs aren't disabled. I don't know why.
        query("/imap/test", "POST", server, true).done(function(data) { //  so using this direct query() approach for now
            if (data.success == true) {
                swal({
                    title: "Success",
                    text: "Logged into <b>" + $("#imaphost").val() + "</b>",
                    type: "success",
                })
            } else {
                swal({
                    title: "Failed!",
                    text: "Unable to login to <b>" + $("#imaphost").val() + "</b>.",
                    type: "error",
                    showCancelButton: true,
                    cancelButtonText: "Close",
                    confirmButtonText: "More Info",
                    confirmButtonColor: "#428bca",
                    allowOutsideClick: false,
                    preConfirm: function () {
                        swal({
                          title: "Error:",
                          text: data.message,
                        })
                    }
                })
            }
            
          })
          .fail(function() {
            swal({
                title: "Failed!",
                text: "An unecpected error occured.",
                type: "error",
            })
          })
          .always(function() {
            //Re-enable inputs and change button text
            $("#imaphost").attr("disabled", false);
            $("#imapport").attr("disabled", false);
            $("#imapusername").attr("disabled", false);
            $("#imappassword").attr("disabled", false);
            $("#use_imap").attr("disabled", false);
            $("#use_tls").attr("disabled", false);
            $("#testimap").attr("disabled", false);
            $("#testimap").html(oldHTML);
          });

      }); //end testclick

    $("#reporttab").click(function() {
        loadIMAPSettings()
    })

    function loadIMAPSettings(){
        api.IMAP.get()
        .success(function (imap) {
            $("#imapusername").val(imap.username)
            $("#imaphost").val(imap.host)
            $("#imapport").val(imap.port)
            $("#imappassword").val(imap.password)
            $('#use_tls').prop('checked', imap.tls)
            $('#use_imap').prop('checked', imap.enabled)
        })
        .error(function () {
            errorFlash("Error fetching IMAP settings")
        })
    }

    var use_map = localStorage.getItem('gophish.use_map')
    $("#use_map").prop('checked', JSON.parse(use_map))
    $("#use_map").on('change', function () {
        localStorage.setItem('gophish.use_map', JSON.stringify(this.checked))
    })

    loadIMAPSettings()

})