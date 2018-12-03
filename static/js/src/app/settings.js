var public_keys = []


// Save attempts to POST or PUT to /public_keys/
function save(idx) {
	var key = {}
    key.name = $("#friendly_name").val()
    key.pub_key = $("#public_key").val()
    if (idx != -1) {
        key.id = public_keys[idx].id
        api.public_keys_id.put(key)
            .success(function (data) {
                successFlash("Public key edited successfully!")
                load()
                dismiss()
            }) 
			.error(function (data) {
                modalError(data.responseJSON.message)
            })
    } else {
        // Submit the page
        api.public_keys.post(key)
            .success(function (data) {
                successFlash("Public key added successfully!")
                load()
                dismiss()
            })
            .error(function (data) {
                modalError(data.responseJSON.message)
            })
    }
}

function dismiss() {
    $("#targetsTable").dataTable().DataTable().clear().draw()
    $("#friendly_name").val("")
	$("#public_key").val("")
	$("#modal").modal('hide')
    $("#modal\\.flashes").empty()
}

function copy(idx) {
    $("#modalSubmit").unbind('click').click(function () {
        save(-1)
    })
    var key = public_keys[idx]
    $("#friendly_name").val("Copy of " + key.name)
    $("#public_key").val(key.pub_key)
}

function edit(idx) {
	$("#modalSubmit").unbind('click').click(function () {
        save(idx)
    })
    var key = {}
    if (idx != -1) {
        key = public_keys[idx]
        $("#friendly_name").val(key.name)
        $("#public_key").val(key.pub_key)
    }
}

var deleteKey = function (idx) {
	  swal({
        title: "Are you sure?",
        text: "This will delete the public key. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete " + escapeHtml(public_keys[idx].name),
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function () {
            return new Promise(function (resolve, reject) {
                api.public_keys_id.delete(public_keys[idx].id)
                    .success(function (msg) {
                        resolve()
                    })
                    .error(function (data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function () {
        swal(
            'Public Key Deleted!',
            'This public key has been deleted!',
            'success'
        );
        $('button:contains("OK")').on('click', function () {
            load()
        })
    })
}

function load() {
	
    $("#tableContainer").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.public_keys.get()
        .success(function (publickeys) {
			public_keys = publickeys
            $("#loading").hide()
			
            if (public_keys.length > 0) {
				
                $("#emptyMessage").hide()
                $("#tableContainer").show()
                var publicKeysTable = $("#publicKeysTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                publicKeysTable.clear();
				
                $.each(publickeys, function (i, key) {
                    publicKeysTable.row.add([
                        escapeHtml(key.name),
                        escapeHtml(key.pub_key),
                         "<div class='pull-right'><span data-toggle='modal' data-backdrop='static' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Edit Page' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button></span>\
		   			 <span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Copy Page' onclick='copy(" + i + ")'>\
                    <i class='fa fa-copy'></i>\
                    </button></span>\
                    <button class='btn btn-danger' data-toggle='tooltip' data-placement='left' title='Delete Page' onclick='deleteKey(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function () {
            errorFlash("Error fetching public keys")
        })
}


$(document).ready(function () {
	load();

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