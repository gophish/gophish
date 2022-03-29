var groups = []

function dismiss() {
    $("#targetsTable").dataTable().DataTable().clear().draw()
    $("#name").val("")
    $("#modal\\.flashes").empty()
}

function instantiateDataTable(id) {
     $('#targetsTable').dataTable( {
        destroy: true, // Destroy any other instantiated table - http://datatables.net/manual/tech-notes/3#destroy
        select: true,
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }],
        "processing": true,
        "serverSide": true,
        "ajax": {
                    url: "/api/groups/" + id,
                    'beforeSend': function (request) {
                        request.setRequestHeader("Authorization", "Bearer " + user.api_key);
                    }
                },
        columns: [
            { data: 'first_name', render: escapeHtml},
            { data: 'last_name', render: escapeHtml },
            { data: 'email', render: escapeHtml },
            { data: 'position',  render: escapeHtml },
            { data: null,
                render: function ( data, type, row ) {
                    return '<span style="cursor:pointer;"><i class="fa fa-trash-o" id="' + data.id + '"></i></span>';
                }
            }]
    } );
}

function edit(id) {

    $("#groupid").val(id)

    if (id == -1 ){ // New group
        $("#targetsTable_wrapper").hide()
        $("#targetsTable").hide()   
    } else {
        api.groupId.summary(id)
            .success(function (group) {
                $("#name").val(escapeHtml(group.name))
            })
            .error(function (data) {
                modalError("Error fetching group name")
            })
        instantiateDataTable(id)
    }

    // Handle file uploads
    csvurl = "/api/groups/" // New group
    method = "POST"
    if (id != -1) {
        csvurl = "/api/groups/" + id // Update existing group
        method = "PUT"
    }
    $("#csvupload").fileupload({
        url: csvurl,
        method: method,
        dataType: "json",
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + user.api_key);
        },
        add: function (e, data) {
            $("#modal\\.flashes").empty()
            name = $("#name").val()
            data.paramName = escapeHtml(name)  // Send group name

            if (name == "") {
                modalError("No group name supplied")
                $("#name").focus();
                return false;
            }
            var acceptFileTypes = /(csv|txt)$/i;
            var filename = data.originalFiles[0]['name']
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                modalError("Unsupported file extension (use .csv or .txt)")
                return false;
            }
            data.submit();
        },
        fail: function(e, data) {
            modalError(data.jqXHR.responseJSON.message);
        },
        done: function (e, data) {
            if (!('id' in data.result)) {
                modalError("Failed to upload CSV file")
            } else {
                $("#targetsTable_wrapper").show()
                $("#targetsTable").show()
                edit(data.result.id)
                load()
            }
        }
    })
}

function saveGroupName() {
    id = parseInt($("#groupid").val())
    if (id == -1) { return }
    name = $("#name").val() // Check for length etc + handle escapes
    data = {"id": id, "name":name}

    api.groupId.rename(id, data)
        .success(function (msg) {
            load()
        })
        .error(function (data) {
            modalError(data.responseJSON.message)
        })
}

var downloadCSVTemplate = function () {
    var csvScope = [{
        'First Name': 'Example',
        'Last Name': 'User',
        'Email': 'foobar@example.com',
        'Position': 'Systems Administrator'
    }]
    var filename = 'group_template.csv'
    var csvString = Papa.unparse(csvScope, {})
    var csvData = new Blob([csvString], {
        type: 'text/csv;charset=utf-8;'
    });
    if (navigator.msSaveBlob) {
        navigator.msSaveBlob(csvData, filename);
    } else {
        var csvURL = window.URL.createObjectURL(csvData);
        var dlLink = document.createElement('a');
        dlLink.href = csvURL;
        dlLink.setAttribute('download', filename)
        document.body.appendChild(dlLink)
        dlLink.click();
        document.body.removeChild(dlLink)
    }
}

var deleteGroup = function (id) {
    var group = groups.find(function (x) {
        return x.id === id
    })
    if (!group) {
        return
    }
    Swal.fire({
        title: "Are you sure?",
        text: "This will delete the group. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete " + escapeHtml(group.name),
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function () {
            return new Promise(function (resolve, reject) {
                api.groupId.delete(id)
                    .success(function (msg) {
                        resolve()
                    })
                    .error(function (data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function (result) {
        if (result.value){
            Swal.fire(
                'Group Deleted!',
                'This group has been deleted!',
                'success'
            );
        }
        $('button:contains("OK")').on('click', function () {
            location.reload()
        })
    })
}

function addTarget(firstNameInput, lastNameInput, emailInput, positionInput) {

    $("#modal\\.flashes").empty()
    groupId = $("#groupid").val()
    target = {
        "email": emailInput.toLowerCase(),
        "first_name": firstNameInput,
        "last_name": lastNameInput,
        "position": positionInput
    }

    if (groupId == -1){ // Create new group with target
        groupname = $("#name").val()
        groupname = escapeHtml(groupname)
        group = {"name": groupname, targets: [target]}

        api.groups.post(group)
        .success(function (data) {
            //ajax fertch and show table
            if (!('id' in data)) {
                modalError("Failed to add target")
            } else {
                instantiateDataTable(data.id)
                $('#targetsTable').DataTable().draw('page');
                $('#targetsTable_wrapper').show()
                $('#targetsTable').show()
                $("#groupid").val(data.id)
                load()
                edit(data.id)
                
            }
        })
        .error(function (data) {
            modalError(data.responseJSON.message)
        })

    } else { // Add single target to existing group
        api.groupId.addtarget(groupId, target)
            .success(function (data){
                load()
            })
            .error(function (data) {
                modalError(data.responseJSON.message)
            })
    }
}

function load() {
    $("#groupTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.groups.summary()
        .success(function (response) {
            $("#loading").hide()
            if (response.total > 0) {
                groups = response.groups
                $("#emptyMessage").hide()
                $("#groupTable").show()
                var groupTable = $("#groupTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                groupTable.clear();
                groupRows = []
                $.each(groups, function (i, group) {
                    groupRows.push([
                        escapeHtml(group.name),
                        escapeHtml(group.num_targets),
                        moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'><button class='btn btn-primary' data-toggle='modal' data-backdrop='static' data-target='#modal' onclick='edit(" + group.id +")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='deleteGroup(" + group.id + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ])
                })
                groupTable.rows.add(groupRows).draw()
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function () {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function () {
    load()
    // Setup the event listeners
    // Handle manual additions
    $("#targetForm").submit(function () {

        // Validate group name is present
        if ($("#name").val() == "") {
            modalError("No group name supplied")
            $("#name").focus();
            return false;
        }

        // Validate the form data
        var targetForm = document.getElementById("targetForm")
        if (!targetForm.checkValidity()) {
            targetForm.reportValidity()
            return
        }
        addTarget(
            $("#firstName").val(),
            $("#lastName").val(),
            $("#email").val(),
            $("#position").val());

        $('#targetsTable').DataTable().draw();

        // Reset user input.
        $("#targetForm>div>input").val('');
        $("#firstName").focus();
        return false;
    });

    // Handle Deletion
    $("#targetsTable").on("click", "span>i.fa-trash-o", function () {
        // We allow emtpy groups with this new pagination model. TODO: Do we need to revisit this?
        targetId=parseInt(this.id)
        groupId=$("#groupid").val()
        target = {"id" : targetId}

        api.groupId.deletetarget(groupId, target)
            .success(function (msg) {
                load()
            })
            .error(function (data) {
                modalError("Failed to delete user. Please try again later.")
            })

        $('#targetsTable').DataTable()
            .row($(this).parents('tr'))
            .remove()
            .draw('page');

    });
    $("#modal").on("hide.bs.modal", function () {
        dismiss();
    });
    $("#csv-template").click(downloadCSVTemplate)
});
