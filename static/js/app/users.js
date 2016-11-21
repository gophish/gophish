var groups = []

function zipBlob(filename, blob, callback) {
  // use a zip.BlobWriter object to write zipped data into a Blob object
  zip.createWriter(new zip.BlobWriter("application/zip"), function(zipWriter) {
    // use a BlobReader object to read the data stored into blob variable
    zipWriter.add(filename, new zip.BlobReader(blob), function() {
      // close the writer and calls callback function
      zipWriter.close(callback);
    });
  }, onerror);
}

function unzipBlob(blob, callback) {
  // use a zip.BlobReader object to read zipped data stored into blob variable
  zip.createReader(new zip.BlobReader(blob), function(zipReader) {
    // get entries from the zip file
    zipReader.getEntries(function(entries) {
      // get data from the first file
      entries[0].getData(new zip.BlobWriter("text/plain"), function(data) {
        // close the reader and calls callback function with uncompressed data as parameter
        zipReader.close();
        callback(data);
      });
    });
  }, onerror);
}

function onerror(message) {
  console.error(message);
}

// Save attempts to POST or PUT to /groups/
function save(idx) {
    var targets = []
    $.each($("#targetsTable").DataTable().rows().data(), function(i, target) {
        targets.push({
            first_name: unescapeHtml(target[0]),
            last_name: unescapeHtml(target[1]),
            email: unescapeHtml(target[2]),
            position: unescapeHtml(target[3])
        })
    })
    var group = {
            name: $("#name").val(),
            targets: targets
        }
        // Submit the group
    if (idx != -1) {
        // If we're just editing an existing group,
        // we need to PUT /groups/:id
        group.id = groups[idx].id
        api.groupId.put(group)
            .success(function(data) {
                successFlash("Group updated successfully!")
                load()
                dismiss()
                $("#modal").modal('hide')
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    } else {
        // Else, if this is a new group, POST it
        // to /groups
        api.groups.post(group)
            .success(function(data) {
                successFlash("Group added successfully!")
                load()
                dismiss()
                $("#modal").modal('hide')
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    }
}

function dismiss() {
    $("#targetsTable").dataTable().DataTable().clear().draw()
    $("#name").val("")
    $("#modal\\.flashes").empty()
}

function edit(idx) {
    targets = $("#targetsTable").dataTable({
        destroy: true, // Destroy any other instantiated table - http://datatables.net/manual/tech-notes/3#destroy
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    })
    $("#modalSubmit").unbind('click').click(function() {
        save(idx)
    })
    if (idx == -1) {
        group = {}
    } else {
        group = groups[idx]
        $("#name").val(group.name)
        $.each(group.targets, function(i, record) {
            targets.DataTable()
                .row.add([
                    escapeHtml(record.first_name),
                    escapeHtml(record.last_name),
                    escapeHtml(record.email),
                    escapeHtml(record.position),
                    '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
                ]).draw()
        });
    }
    // Handle file uploads
    $("#csvupload").fileupload({
        dataType: "json",
        add: function(e, data) {
            $("#modal\\.flashes").empty()
            var acceptFileTypes = /(csv|txt)$/i;
            var filename = data.originalFiles[0]['name']
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                modalError("Unsupported file extension (use .csv or .txt)")
                return false;
            }
            var reader = new FileReader();
            reader.readAsText(data.files[0]);
            reader.onload = function(e) {
                // console.log(e.target.result);
                var blob = new Blob([e.target.result], {
                  type : "text/plain"
                });
                zipBlob("users.txt", blob, function(zippedBlob) {
                  // saveAs(zippedBlob,"hello.zip")
                  var fd = new FormData();
                  fd.append('fname', 'compressed.zip');
                  fd.append('data', zippedBlob);
                  $.ajax({
                      type: 'POST',
                      url: '/api/import/group',
                      data: fd,
                      processData: false,
                      contentType: false
                  }).done(function(data) {
                    console.log(data);
                    $.each(data, function(i, record) {
                        addTarget(
                            record.first_name,
                            record.last_name,
                            record.email,
                            record.position);
                    });
                    targets.DataTable().draw();
                  });
                });
            };
            // data.submit();
        },
        done: function(e, data) {
            $.each(data.result, function(i, record) {
                addTarget(
                    record.first_name,
                    record.last_name,
                    record.email,
                    record.position);
            });
            targets.DataTable().draw();
        }
    })
}

function deleteGroup(idx) {
    if (confirm("Delete " + groups[idx].name + "?")) {
        api.groupId.delete(groups[idx].id)
            .success(function(data) {
                successFlash(data.message)
                load()
            })
    }
}

function addTarget(firstNameInput, lastNameInput, emailInput, positionInput) {
    // Create new data row.
    var email = escapeHtml(emailInput).toLowerCase();
    var newRow = [
        escapeHtml(firstNameInput),
        escapeHtml(lastNameInput),
        email,
        escapeHtml(positionInput),
        '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
    ];

    // Check table to see if email already exists.
    var targetsTable = targets.DataTable();
    var existingRowIndex = targetsTable
        .column(2, {
            order: "index"
        }) // Email column has index of 2
        .data()
        .indexOf(email);
    // Update or add new row as necessary.
    if (existingRowIndex >= 0) {
        targetsTable
            .row(existingRowIndex, {
                order: "index"
            })
            .data(newRow);
    } else {
        targetsTable.row.add(newRow);
    }
}

function load() {
    $("#groupTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.groups.get()
        .success(function(gs) {
            $("#loading").hide()
            if (gs.length > 0) {
                groups = gs
                $("#emptyMessage").hide()
                $("#groupTable").show()
                groupTable = $("#groupTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                groupTable.clear();
                $.each(groups, function(i, group) {
                    var targets = ""
                    $.each(group.targets, function(i, target) {
                        targets += target.email + ", "
                        if (targets.length > 50) {
                            targets = targets.slice(0, -3) + "..."
                            return false;
                        }
                    })
                    groupTable.row.add([
                        escapeHtml(group.name),
                        escapeHtml(targets),
                        moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'><button class='btn btn-primary' data-toggle='modal' data-target='#modal' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='deleteGroup(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function() {
    zip.workerScriptsPath = '/js/zip/';
    load()
        // Setup the event listeners
        // Handle manual additions
    $("#targetForm").submit(function() {
        addTarget(
            $("#firstName").val(),
            $("#lastName").val(),
            $("#email").val(),
            $("#position").val());
        targets.DataTable().draw();

        // Reset user input.
        $("#targetForm>div>input").val('');
        $("#firstName").focus();
        return false;
    });
    // Handle Deletion
    $("#targetsTable").on("click", "span>i.fa-trash-o", function() {
        targets.DataTable()
            .row($(this).parents('tr'))
            .remove()
            .draw();
    });
    $("#modal").on("hide.bs.modal", function() {
        dismiss();
    });
});
