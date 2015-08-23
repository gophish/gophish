/*
	landing_pages.js
	Handles the creation, editing, and deletion of landing pages
	Author: Jordan Wright <github.com/jordan-wright>
*/
var pages = []
// Save attempts to POST to /templates/
function save(idx){
    var page = {}
    page.name = $("#name").val()
    page.html = CKEDITOR.instances["html_editor"].getData();
    if (idx != -1){
        page.id = page[idx].id
        api.landing_pageId.put(page)
        .success(function(data){
            successFlash("Page edited successfully!")
            load()
            dismiss()
        })
    } else {
        // Submit the page
        api.landing_pages.post(page)
        .success(function(data){
            successFlash("Page added successfully!")
            load()
            dismiss()
        })
        .error(function(data){
            modalError(data.responseJSON.message)
        })
    }
}

function dismiss(){
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#name").val("")
    $("#html_editor").val("")
}

function deleteTemplate(idx){
    if (confirm("Delete " + pages[idx].name + "?")){
        api.landing_pageId.delete(pages[idx].id)
        .success(function(data){
            successFlash(data.message)
            load()
        })
    }
}

function edit(idx){
    $("#modalSubmit").unbind('click').click(function(){save(idx)})
    $("#html_editor").ckeditor()
    var page = {}
    if (idx != -1) {
        page = pages[idx]
        $("#name").val(page.name)
        $("#html_editor").val(page.html)
    }
}

function load(){
    $("#pagesTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.landing_pages.get()
    .success(function(ps){
        pages = ps
        $("#loading").hide()
        if (pages.length > 0){
            $("#pagesTable").show()
            pagesTable = $("#templateTable").DataTable();
            pagesTable.clear()
            $.each(pages, function(i, page){
                pagesTable.row.add([
                    page.name,
                    moment(page.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                    "<div class='pull-right'><button class='btn btn-primary' data-toggle='modal' data-target='#modal' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='deletePage(" + i + ")'>\
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
        errorFlash("Error fetching pages")
    })
}

$(document).ready(function(){
    load()
})
