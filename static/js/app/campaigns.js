// labels is a map of campaign statuses to
// CSS classes
var labels = {
    "In progress": "label-primary",
    "Queued": "label-info",
    "Completed": "label-success",
    "Emails Sent": "label-success",
    "Error": "label-danger"
}

var campaigns = []

// Save attempts to POST to /campaigns/
function save() {
    groups = []
    $.each($("#groupTable").DataTable().rows().data(), function(i, group) {
        groups.push({
            name: group[0]
        })
    })
    var campaign = {
            name: $("#name").val(),
            template: {
                name: $("#template").val()
            },
            url: $("#url").val(),
            page: {
                name: $("#page").val()
            },
            smtp: {
                from_address: $("input[name=from]").val(),
                host: $("input[name=host]").val(),
                username: $("input[name=username]").val(),
                password: $("input[name=password]").val(),
            },
            groups: groups
        }
        // Submit the campaign
    api.campaigns.post(campaign)
        .success(function(data) {
            successFlash("Campaign successfully launched!")
            window.location = "/campaigns/" + campaign.id.toString()
        })
        .error(function(data) {
            $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
        })
}

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#groupTable").dataTable().DataTable().clear().draw()
}

function deleteCampaign(idx) {
    if (confirm("Delete " + campaigns[idx].name + "?")) {
        api.campaignId.delete(campaigns[idx].id)
            .success(function(data) {
                successFlash(data.message)
                load()
            })
    }
}

function edit(campaign) {
    // Clear the bloodhound instance
    group_bh.clear();
    template_bh.clear();
    page_bh.clear();
    if (campaign == "new") {
        api.groups.get()
            .success(function(groups) {
                if (groups.length == 0) {
                    modalError("No groups found!")
                    return false;
                } else {
                    group_bh.add(groups)
                }
            })
        api.templates.get()
            .success(function(templates) {
                if (templates.length == 0) {
                    modalError("No templates found!")
                    return false
                } else {
                    template_bh.add(templates)
                }
            })
        api.pages.get()
            .success(function(pages) {
                if (pages.length == 0) {
                    modalError("No pages found!")
                    return false
                } else {
                    page_bh.add(pages)
                }
            })
    }
}

$(document).ready(function() {
    api.campaigns.get()
        .success(function(cs) {
	    campaigns = cs
            $("#loading").hide()
            if (campaigns.length > 0) {
                $("#campaignTable").show()
                campaignTable = $("#campaignTable").DataTable();
                $.each(campaigns, function(i, campaign) {
                    label = labels[campaign.status] || "label-default";
                    campaignTable.row.add([
                        campaign.name,
                        moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<span class=\"label " + label + "\">" + campaign.status + "</span>",
                        "<div class='pull-right'><a class='btn btn-primary' href='/campaigns/" + campaign.id + "'>\
                    <i class='fa fa-bar-chart'></i>\
                    </a>\
                    <button class='btn btn-danger' onclick='deleteCampaign(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            $("#loading").hide()
            errorFlash("Error fetching campaigns")
        })
    $("#groupForm").submit(function() {
            groupTable.row.add([
                $("#groupSelect").val(),
                '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
            ]).draw()
            $("#groupTable").on("click", "span>i.fa-trash-o", function() {
                groupTable.row($(this).parents('tr'))
                    .remove()
                    .draw();
            })
            return false;
        })
        // Create the group typeahead objects
    groupTable = $("#groupTable").DataTable()
    group_bh = new Bloodhound({
        datumTokenizer: function(g) {
            return Bloodhound.tokenizers.whitespace(g.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    group_bh.initialize()
    $("#groupSelect.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "groups",
            source: group_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No groups matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + data.name + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, group) {
            $("#groupSelect").typeahead('val', group.name)
        })
        .bind('typeahead:autocomplete', function(ev, group) {
            $("#groupSelect").typeahead('val', group.name)
        });
    // Create the template typeahead objects
    template_bh = new Bloodhound({
        datumTokenizer: function(t) {
            return Bloodhound.tokenizers.whitespace(t.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    template_bh.initialize()
    $("#template.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "templates",
            source: template_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No templates matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + data.name + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, template) {
            $("#template").typeahead('val', template.name)
        })
        .bind('typeahead:autocomplete', function(ev, template) {
            $("#template").typeahead('val', template.name)
        });
    // Create the landing page typeahead objects
    page_bh = new Bloodhound({
        datumTokenizer: function(p) {
            return Bloodhound.tokenizers.whitespace(p.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    page_bh.initialize()
    $("#page.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "pages",
            source: page_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No pages matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + data.name + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, page) {
            $("#page").typeahead('val', page.name)
        })
        .bind('typeahead:autocomplete', function(ev, page) {
            $("#page").typeahead('val', page.name)
        });
})
