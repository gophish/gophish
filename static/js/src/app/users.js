var groups = []

// Save attempts to POST or PUT to /groups/
function save(id) {
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
    if (id != -1) {
        // If we're just editing an existing group,
        // we need to PUT /groups/:id
        group.id = id
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
    $("#users").val("").change();
    $("#subgroups").val("").change();
}

function edit(id) {
    targets = $("#targetsTable").dataTable({
        destroy: true, // Destroy any other instantiated table - http://datatables.net/manual/tech-notes/3#destroy
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    })
    $("#modalSubmit").unbind('click').click(function() {
        save(id)
    })
    if (id == -1) {
        var group = {}
    } else {
        api.groupId.get(id)
            .success(function(group) {
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

            })
            .error(function() {
                errorFlash("Error fetching group")
            })
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
            data.submit();
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

function createRandomSubgroups() {
    api.groups.get()
        .success(function(groups) {
            if (groups.length == 0) {
                modalError("No groups found!")
                return false;
            } else {
                var group_s2 = $.map(groups, function(obj) {
                    obj.text = obj.name
                    obj.value = obj.id
                    return obj
                });
                $("#users.form-control").select2({
                    placeholder: "Select Group",
                    data: group_s2,
                });
            }
        });


    $("#subgroupModalSubmit").unbind('click').click(function() {
        var parentTargets = [];
        var parentGroup = {};
        var select = $("#users")[0];
        var subgroups = Number.parseInt($("#subgroups").val());
        if (select.options.selectedIndex != -1 && subgroups) {
            api.groupId.get($(select.options[select.options.selectedIndex]).val())
                .success(function(data) {
                    parentTargets = data.targets;
                    parentGroup = data;

                    totalTargets = parentTargets.length;

                    if (subgroups > 1 && subgroups <= totalTargets) {
                        var count = Math.floor(totalTargets / subgroups);
                        var extra = totalTargets % subgroups;
                        console.log(extra);
                        var newGroups = [];
                        var date = new Date();
                        for (var i = 0; i < subgroups; i++) {
                            newGroups.push({
                                name: parentGroup.name + "-" + (date.getTime() + i),
                                targets: []
                            })
                        }

                        for (var i = 0; i < count; i++) {
                            for (var j = 0; j < newGroups.length; j++) {
                                newGroups[j].targets.push(parentTargets.splice(Math.floor(Math.random() * parentTargets.length), 1)[0]);
                            }
                        }

                        var indices = [];
                        for (var i = 0; i < newGroups.length; i++) {
                            indices.push(i);
                        }
                        for (var i = 0; i < extra; i++) {
                            var index = indices.splice(Math.floor(Math.random() * newGroups.length), 1);
                            newGroups[index].targets.push(parentTargets.splice(Math.floor(Math.random() * parentTargets.length), 1)[0]);
                        }

                        for (var i = 0; i < newGroups.length; i++) {
                            api.groups.post(newGroups[i])
                                .success(function(data) {
                                    successFlash("Subgroups created successfully!")
                                    load()
                                    dismiss()
                                    $("#modal").modal('hide')
                                })
                                .error(function(data) {
                                    console.log(newGroups[i]);
                                    $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
                                })
                        }
                    } else {
                        modalError("Invalid number of subgroups!");
                    }
                })
                .error(function(data) {
                    modalError(data.responseJSON.message);
                })
        } else {
            modalError("Invalid options!");
        }
    })
}

function deleteGroup(id) {
    var group = groups.find(function(x){return x.id === id})
    if (!group) {
        console.log('wat');
        return
    }
    if (confirm("Delete " + group.name + "?")) {
        api.groupId.delete(id)
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
    $("#subgroupsButton").attr('disabled',false)
    $("#loading").show()
    api.groups.summary()
        .success(function(response) {
            $("#loading").hide()
            if (response.total > 0) {
                groups = response.groups
                $("#emptyMessage").hide()
                $("#subgroupsButton").attr('disabled',false)
                $("#groupTable").show()
                var groupTable = $("#groupTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                groupTable.clear();
                $.each(groups, function(i, group) {
                    groupTable.row.add([
                        escapeHtml(group.name),
                        escapeHtml(group.num_targets),
                        moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'><button class='btn btn-primary' data-toggle='modal' data-target='#modal' onclick='edit(" + group.id + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='deleteGroup(" + group.id + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
                $("#subgroupsButton").attr('disabled',true)
            }
        })
        .error(function() {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function() {
    $('.modal').on('hidden.bs.modal', function(event) {
        $(this).removeClass('fv-modal-stack');
        $('body').data('fv_open_modals', $('body').data('fv_open_modals') - 1);
    });
    $('.modal').on('shown.bs.modal', function(event) {
        // Keep track of the number of open modals
        if (typeof($('body').data('fv_open_modals')) == 'undefined') {
            $('body').data('fv_open_modals', 0);
        }
        // if the z-index of this modal has been set, ignore.
        if ($(this).hasClass('fv-modal-stack')) {
            return;
        }
        $(this).addClass('fv-modal-stack');
        // Increment the number of open modals
        $('body').data('fv_open_modals', $('body').data('fv_open_modals') + 1);
        // Setup the appropriate z-index
        $(this).css('z-index', 1040 + (10 * $('body').data('fv_open_modals')));
        $('.modal-backdrop').not('.fv-modal-stack').css('z-index', 1039 + (10 * $('body').data('fv_open_modals')));
        $('.modal-backdrop').not('fv-modal-stack').addClass('fv-modal-stack');
    });
    // Scrollbar fix - https://stackoverflow.com/questions/19305821/multiple-modals-overlay
    $(document).on('hidden.bs.modal', '.modal', function () {
        $('.modal:visible').length && $(document.body).addClass('modal-open');
    });
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
    $(".modal").on("hide.bs.modal", function() {
        dismiss();
    });

    // Select2 Defaults
    $.fn.select2.defaults.set("width", "100%");
    $.fn.select2.defaults.set("dropdownParent", $("#modal_body"));
    $.fn.select2.defaults.set("theme", "bootstrap");
});
