// Save attempts to POST to /campaigns/
function save(){
    var targets = {}
    var group = {
        name: $("#name").val(),
        targets: targets
    }
    // Submit the campaign
    api.groups.post(group)
    .success(function(data){
        successFlash("Campaign successfully launched!")
        load()
    })
    .error(function(data){
        $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
    })
}

function groupAdd(name){
    groups.append({
        name: name
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
                groupTable.row.add([
                    group.Name,
                    group.targets,
                    group.modified_date
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
})
