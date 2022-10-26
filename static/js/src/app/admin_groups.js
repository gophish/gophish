let admin_groups = [];

const save = (id) => {
    let newAdminGroup = {
        name: $("#name").val(),
        users: $("#users").select2("data")
    };

    if (id !== -1) {
        newAdminGroup.id = id;
        api.adminGroupsId.put(newAdminGroup)
            .success((data) => {
                successFlash(`Administration Groupe ${escapeHtml(newAdminGroup.name)} updated successfully!`)
                load();
                dismiss();
                $("#modal").modal('hide');
            })
            .error((data) => {
                modalError(data.responseJSON.message);
            });
    } else {
        api.adminGroups.post(newAdminGroup)
            .success(() => {
                successFlash(`Admin group ${escapeHtml(newAdminGroup.name)} created successfully!`);
                load();
                dismiss();
                $("#modal").modal('hide');
            })
            .error((data) => {
                modalError(data.responseJSON.message)
            });
    }
}

const edit = (id) => {
    $("#modalSubmit").unbind('click').click(() => {
        save(id);
    });

    api.users.get()
        .success((users) => {
            $("#users").select2({
                data: users.map(function(user) {
                    user.text = user.username;

                    return user
                })
            });
            $("#users").val(null).trigger('change');
            if (id === -1) {
                $("#adminGroupModalLabel").text("New Administration Group");
            } else {
                $("#adminGroupModalLabel").text("Edit Administration Group");
                api.adminGroupsId.get(id)
                    .success((groupAdmin) => {
                        $("#name").val(groupAdmin.name);
                        const alreadySelectedUsers = users.map(function(user) {
                            if (groupAdmin.users 
                                && groupAdmin.users.find(u => u.id === user.id) !== undefined)
                                return user.id;

                            return null
                        });
                        $("#users").val(alreadySelectedUsers).trigger('change');
                    })
                    .error(function () {
                        errorFlash("Error fetching Administration Group")
                    });
            }
        });
};

const load = () => {
    $("#adminGroupsTable").hide();
    $("#loading").show();
    api.adminGroups.get()
        .success((ags) => {
            admin_groups = ags;
            $("#loading").hide();
            $("#adminGroupsTable").show();
            let adminGroupsTable = $("#adminGroupsTable").DataTable({
                destroy: true,
                columnDefs: [{
                    orderable: false,
                    targets: "no-sort"
                }]
            });
            adminGroupsTable.clear();
            adminGroupsRows = [];
            $.each(admin_groups, (i, group) => {
                adminGroupsRows.push([
                    escapeHtml(group.name),
                    "<div class='pull-right'>\
                        <button class='btn btn-primary edit_button' data-toggle='modal' data-backdrop='static' data-target='#modal' data-admin-group-id='" + group.id + "'>\
                            <i class='fa fa-pencil'></i>\
                        </button>\
                        <button class='btn btn-danger delete_button' data-admin-group-id='" + group.id + "'>\
                            <i class='fa fa-trash-o'></i>\
                        </button>\
                    </div>"
                ])
            });
            adminGroupsTable.rows.add(adminGroupsRows).draw();
        })
        .error(() => {
            errorFlash("Error fetching administration groups");
        });
}

const dismiss = () => {
    $("#name").val("")
    $("#modal\\.flashes").empty()
};

const deleteAdminGroup = (id) => {
    const adminGroup = admin_groups.find(x => x.id == id);

    if (!user) return

    Swal.fire({
        title: "Are you sure?",
        text: `This will delete the administration group '${escapeHtml(adminGroup.name)}'.\n\nThis can't be undone!`,
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function () {
            return new Promise((resolve, reject) => {
                api.adminGroupsId.delete(id)
                    .success((m) => {
                        resolve()
                    })
                    .error((data) => {
                        reject(data.responseJSON.message)
                    });
            })
            .catch(error => {
                Swal.showValidationMessage(error);
            });
        }
    }).then((result) => {
        if (result.value) {
            Swal.fire(
                'Administration Group Deleted!'
                `The administration group ${escapeHtml(group.name)} has been deleted!`,
                'success'
            );
        }

        $('button:contains("OK")').on('click', function () {
            location.reload();
        });
    });
};

$(document).ready(function () {
    load();
    $("#modal").on("hide.bs.modal", function () {
        dismiss();
    });
    $.fn.select2.defaults.set("width", "100%");
    $.fn.select2.defaults.set("dropdownParent", $("#users-select"));
    $.fn.select2.defaults.set("theme", "bootstrap");

    $("#new_button").on("click", function() {
        edit(-1);
    });

    $("#adminGroupsTable").on('click', '.delete_button', function (e) {
        deleteAdminGroup($(this).attr('data-admin-group-id'));
    });

    $("#adminGroupsTable").on('click', '.edit_button', function (e) {
        edit($(this).attr('data-admin-group-id'));
    });
})
