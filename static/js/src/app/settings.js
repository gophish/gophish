$(document).ready(function () {
    $('[data-toggle="tooltip"]').tooltip();
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
    //$("#imapForm").submit(function (e) {
    $("#savesettings").click(function() {
        var imapSettings = {}
        imapSettings.host = $("#imaphost").val()
        imapSettings.port = $("#imapport").val()
        imapSettings.username = $("#imapusername").val()
        imapSettings.password = $("#imappassword").val()
        imapSettings.enabled = $('#use_imap').prop('checked')
        imapSettings.tls = $('#use_tls').prop('checked')

        //Advanced settings
        imapSettings.folder = $("#folder").val()
        imapSettings.imap_freq = $("#imapfreq").val()
        imapSettings.restrict_domain = $("#restrictdomain").val()
        imapSettings.ignore_cert_errors = $('#ignorecerterrors').prop('checked')
        imapSettings.delete_reported_campaign_email = $('#deletecampaign').prop('checked')
        
        //To avoid unmarshalling error in controllers/api/imap.go. It would fail gracefully, but with a generic error.
        if (imapSettings.host == ""){
            errorFlash("No IMAP Host specified")
            document.body.scrollTop = 0;
            document.documentElement.scrollTop = 0;
            return false
        }
        if (imapSettings.port == ""){
            errorFlash("No IMAP Port specified")
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
        if (imapSettings.imap_freq == ""){
            imapSettings.imap_freq = "60"
        }

        api.IMAP.post(imapSettings).done(function (data) {
                if (data.success == true) {
                    successFlashFade("Successfully updated IMAP settings.", 2)
                } else {
                    errorFlash("Unable to update IMAP settings.")
                }
            })
            .success(function (data){
                loadIMAPSettings()
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

    $("#validateimap").click(function() {

        // Query validate imap server endpoint
        var server = {}
        server.host = $("#imaphost").val()
        server.port = $("#imapport").val()
        server.username = $("#imapusername").val()
        server.password = $("#imappassword").val()
        server.tls = $('#use_tls').prop('checked')
        server.ignore_cert_errors = $('#ignorecerterrors').prop('checked')

        //To avoid unmarshalling error in controllers/api/imap.go. It would fail gracefully, but with a generic error. 
        if (server.host == ""){
            errorFlash("No IMAP Host specified")
            document.body.scrollTop = 0;
            document.documentElement.scrollTop = 0;
            return false
        }
        if (server.port == ""){
            errorFlash("No IMAP Port specified")
            document.body.scrollTop = 0;
            document.documentElement.scrollTop = 0;
            return false
        }
        if (isNaN(server.port) || server.port <1 || server.port > 65535  ){
            errorFlash("Invalid IMAP Port")
            document.body.scrollTop = 0;
            document.documentElement.scrollTop = 0;
            return false
        }

        var oldHTML = $("#validateimap").html();
        // Disable inputs and change button text
        $("#imaphost").attr("disabled", true);
        $("#imapport").attr("disabled", true);
        $("#imapusername").attr("disabled", true);
        $("#imappassword").attr("disabled", true);
        $("#use_imap").attr("disabled", true);
        $("#use_tls").attr("disabled", true);
        $('#ignorecerterrors').attr("disabled", true);
        $("#folder").attr("disabled", true);
        $("#restrictdomain").attr("disabled", true);
        $('#deletecampaign').attr("disabled", true);
        $('#lastlogin').attr("disabled", true);
        $('#imapfreq').attr("disabled", true);
        $("#validateimap").attr("disabled", true);  
        $("#validateimap").html("<i class='fa fa-circle-o-notch fa-spin'></i> Testing...");
        
        api.IMAP.validate(server).done(function(data) {
            if (data.success == true) {
                Swal.fire({
                    title: "Success",
                    html: "Logged into <b>" + escapeHtml($("#imaphost").val()) + "</b>",
                    type: "success",
                })
            } else {
                Swal.fire({
                    title: "Failed!",
                    html: "Unable to login to <b>" + escapeHtml($("#imaphost").val()) + "</b>.",
                    type: "error",
                    showCancelButton: true,
                    cancelButtonText: "Close",
                    confirmButtonText: "More Info",
                    confirmButtonColor: "#428bca",
                    allowOutsideClick: false,
                }).then(function(result) {
                    if (result.value) {
                        Swal.fire({
                            title: "Error:",
                            text: data.message,
                        })
                    }
                  })
            }
            
          })
          .fail(function() {
            Swal.fire({
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
            $('#ignorecerterrors').attr("disabled", false);
            $("#folder").attr("disabled", false);
            $("#restrictdomain").attr("disabled", false);
            $('#deletecampaign').attr("disabled", false);
            $('#lastlogin').attr("disabled", false);
            $('#imapfreq').attr("disabled", false);
            $("#validateimap").attr("disabled", false);
            $("#validateimap").html(oldHTML);

          });

      }); //end testclick

    $("#reporttab").click(function() {
        loadIMAPSettings()
    })

    $("#advanced").click(function() {
        $("#advancedarea").toggle();
    })

    function loadIMAPSettings(){
        api.IMAP.get()
        .success(function (imap) {
            if (imap.length == 0){
                $('#lastlogindiv').hide()
            } else {
                imap = imap[0]
                if (imap.enabled == false){
                    $('#lastlogindiv').hide()
                } else {
                    $('#lastlogindiv').show()
                }
                $("#imapusername").val(imap.username)
                $("#imaphost").val(imap.host)
                $("#imapport").val(imap.port)
                $("#imappassword").val(imap.password)
                $('#use_tls').prop('checked', imap.tls)
                $('#ignorecerterrors').prop('checked', imap.ignore_cert_errors)
                $('#use_imap').prop('checked', imap.enabled)
                $("#folder").val(imap.folder)
                $("#restrictdomain").val(imap.restrict_domain)
                $('#deletecampaign').prop('checked', imap.delete_reported_campaign_email)
                $('#lastloginraw').val(imap.last_login)
                $('#lastlogin').val(moment.utc(imap.last_login).fromNow())
                $('#imapfreq').val(imap.imap_freq)
            }  

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
