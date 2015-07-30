// labels is a map of campaign statuses to
// CSS classes
var labels = {
    "In progress" : "label-primary",
    "Queued" : "label-info",
	"Completed" : "label-success",
	"Error" : "label-danger"
}

// Save attempts to POST to /campaigns/
function save(){
    var campaign = {
        name: $("#name").val(),
        template:{
            name: $("#template").val()
        },
        smtp: {
            from_address: $("input[name=from]").val(),
            host: $("input[name=host]").val(),
            username: $("input[name=username]").val(),
            password: $("input[name=password]").val(),
        },
        groups: [{name : "Test group"}]
    }
    // Submit the campaign
    api.campaigns.post(campaign)
    .success(function(data){
        successFlash("Campaign successfully launched!")
        load()
    })
    .error(function(data){
        $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
    })
}

function dismiss(){
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#groupTable").dataTable().DataTable().clear().draw()
}

function edit(campaign){
    // Clear the bloodhound instance
    bh.clear();
    if (campaign == "new") {
        api.groups.get()
        .success(function(groups){
            if (groups.length == 0){
                modalError("No groups found!")
                return false;
            }
            else {
                bh.add(groups)
            }
        })
    }
}

$(document).ready(function(){
    api.campaigns.get()
    .success(function(campaigns){
        $("#loading").hide()
        if (campaigns.length > 0){
            $("#campaignTable").show()
            campaignTable = $("#campaignTable").DataTable();
            $.each(campaigns, function(i, campaign){
                label = labels[campaign.status] || "label-default";
                campaignTable.row.add([
                    campaign.name,
                    moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a'),
                    "<span class=\"label " + label + "\">" + campaign.status + "</span>",
                    "<div class='pull-right'><button class='btn btn-primary' onclick='alert(\"test\")'>\
                    <i class='fa fa-bar-chart'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='alert(\"test\")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                ]).draw()
            })
        } else {
            $("#emptyMessage").show()
        }
    })
    .error(function(){
        $("#loading").hide()
        errorFlash("Error fetching campaigns")
    })
    $("#groupForm").submit(function(){
        groupTable.row.add([
            $("#groupSelect").val(),
            '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
        ]).draw()
        $("#groupTable").on("click", "span>i.fa-trash-o", function(){
            groupTable.row( $(this).parents('tr') )
            .remove()
            .draw();
        })
        return false;
    })
    // Create the group typeahead objects
    groupTable = $("#groupTable").DataTable()
    suggestion_template = Hogan.compile('<div>{{name}}</div>')
    bh = new Bloodhound({
        datumTokenizer: function(g) { return Bloodhound.tokenizers.whitespace(g.name) },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    bh.initialize()
    $("#groupSelect.typeahead.form-control").typeahead({
        hint: true,
        highlight: true,
        minLength: 1
    },
    {
        name: "groups",
        source: bh,
        templates: {
            empty: function(data) {return '<div class="tt-suggestion">No groups matched that query</div>' },
            suggestion: function(data){ return '<div>' + data.name + '</div>' }
        }
    })
    .bind('typeahead:select', function(ev, group){
        $("#groupSelect").typeahead('val', group.name)
    });
})
