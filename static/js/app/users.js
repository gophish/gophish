// Save attempts to POST to /groups/
function save(){
    var targets = []
    $.each($("#targetsTable").DataTable().rows().data(), function(i, target){
        targets.push({
            first_name : target[0],
            last_name: target[1],
            email: target[2],
            position: target[3]
        })
    })
    var group = {
        name: $("#name").val(),
        targets: targets
    }
    console.log(group)
    // Submit the group
    api.groups.post(group)
    .success(function(data){
        successFlash("Group added successfully!")
        load()
        dismiss()
    })
    .error(function(data){
        modalError(data.responseJSON.message)
    })
}

function dismiss(){
    $("#targetsTable").dataTable().DataTable().clear().draw()
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
}

function edit(group){
    if (group == "new") {
        group = {}
    }
    // Handle file uploads
    targets = $("#targetsTable").dataTable()
    $("#csvupload").fileupload({
        dataType:"json",
        add: function(e, data){
            $("#modal\\.flashes").empty()
            var acceptFileTypes= /(csv|txt)$/i;
            var filename = data.originalFiles[0]['name']
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                modalError("Unsupported file extension (use .csv or .txt)")
                return false;
            }
            data.submit();
        },
        done: function(e, data){
            console.log(data.result)
            $.each(data.result, function(i, record) {
                targets.DataTable()
                .row.add([
                    record.first_name,
                    record.last_name,
                    record.email,
                    record.position,
                    '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
                ]).draw()
            });
        }
    })
    // Handle manual additions
    $("#targetForm").submit(function(){
        targets.DataTable()
        .row.add([
            $("#firstName").val(),
            $("#lastName").val(),
            $("#email").val(),
            $("#position").val(),
            '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
        ])
        .draw()
        $("#targetForm>div>input").val('')
        $("#firstName").focus()
        return false
    })
    // Handle Deletion
    $("#targetsTable").on("click", "span>i.fa-trash-o", function(){
        targets.DataTable()
        .row( $(this).parents('tr') )
        .remove()
        .draw();
    })
}

function load(){
    api.groups.get()
    .success(function(groups){
        if (groups.length > 0){
            $("#emptyMessage").hide()
            $("#groupTable").show()
            groupTable = $("#groupTable").DataTable();
            $.each(groups, function(i, group){
                var targets = ""
                $.each(group.targets, function(i, target){
                    targets += target.email + ", "
                    if (targets.length > 50) {
                        targets = targets.slice(0,-3) + "..."
                        return false;
                    }
                })
                groupTable.row.add([
                    group.name,
                    targets,
                    moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                    "<div class='pull-right'><button class='btn btn-primary' onclick='alert(\"test\")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='alert(\"test\")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                ]).draw()
            })
        }
    })
    .error(function(){
        errorFlash("Error fetching groups")
    })
}

$(document).ready(function(){
    load()
    $("#fileUpload").hover(function(){$("#fileUpload").tooltip('toggle')})
})
