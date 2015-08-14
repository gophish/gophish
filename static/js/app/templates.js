var templates = []
var icons = {
    "application/vnd.ms-excel" : "fa-file-excel-o",
    "text/plain" : "fa-file-text-o",
    "image/gif" : "fa-file-image-o",
    "image/png" : "fa-file-image-o",
    "application/pdf" : "fa-file-pdf-o",
    "application/x-zip-compressed" : "fa-file-archive-o",
    "application/x-gzip" : "fa-file-archive-o",
    "application/vnd.openxmlformats-officedocument.presentationml.presentation" : "fa-file-powerpoint-o",
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document" : "fa-file-word-o",
    "application/octet-stream" : "fa-file-o",
    "application/x-msdownload" : "fa-file-o"
}

// Save attempts to POST to /templates/
function save(idx){
    var template = {attachments:[]}
    template.name = $("#name").val()
    template.subject = $("#subject").val()
    template.html = CKEDITOR.instances["html_editor"].getData();
    template.text = $("#text_editor").val()
    // Add the attachments
    $.each($("#attachmentsTable").DataTable().rows().data(), function(i, target){
        template.attachments.push({
            name : target[1],
            content: target[3],
            type: target[4],
        })
    })
    if (idx != -1){
        template.id = templates[idx].id
        api.templateId.put(template)
        .success(function(data){
            successFlash("Template edited successfully!")
            load()
            dismiss()
        })
    } else {
        // Submit the template
        api.templates.post(template)
        .success(function(data){
            successFlash("Template added successfully!")
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
    $("#attachmentsTable").dataTable().DataTable().clear().draw()
    $("#name").val("")
    $("#text_editor").val("")
    $("#html_editor").val("")
}

function attach(files){
    attachmentsTable = $("#attachmentsTable").DataTable();
    $.each(files, function(i, file){
        var reader = new FileReader();
        /* Make this a datatable */
        reader.onload = function(e){
            var icon = icons[file.type] || "fa-file-o"
            // Add the record to the modal
            attachmentsTable.row.add([
                '<i class="fa ' + icon + '"></i>',
                file.name,
                '<span class="remove-row"><i class="fa fa-trash-o"></i></span>',
                reader.result.split(",")[1],
                file.type || "application/octet-stream"
            ]).draw()
        }
        reader.onerror = function(e) {
            console.log(e)
        }
        reader.readAsDataURL(file)
    })
}

function edit(idx){
    $("#modalSubmit").unbind('click').click(function(){save(idx)})
    $("#attachmentUpload").unbind('click').click(function(){this.value=null})
    $("#html_editor").ckeditor()
    $("#attachmentsTable").show()
    attachmentsTable = null
    if ( $.fn.dataTable.isDataTable('#attachmentsTable') ) {
        attachmentsTable = $('#attachmentsTable').DataTable();
    }
    else {
        attachmentsTable = $("#attachmentsTable").DataTable({
            "aoColumnDefs" : [{
                "targets" : [3,4],
                "sClass" : "datatable_hidden"
            }]
        });
    }
    var template = {attachments:[]}
    if (idx != -1) {
        template = templates[idx]
        $("#name").val(template.name)
        $("#html_editor").val(template.html)
        $("#text_editor").val(template.text)
        $.each(template.attachments, function(i, file){
            var icon = icons[file.type] || "fa-file-o"
            // Add the record to the modal
            attachmentsTable.row.add([
                '<i class="fa ' + icon + '"></i>',
                file.name,
                '<span class="remove-row"><i class="fa fa-trash-o"></i></span>',
                file.content,
                file.type || "application/octet-stream"
            ]).draw()
        })
    }
    // Handle Deletion
    $("#attachmentsTable").unbind('click').on("click", "span>i.fa-trash-o", function(){
        attachmentsTable.row( $(this).parents('tr') )
        .remove()
        .draw();
    })
}

function load(){
    $("#templateTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.templates.get()
    .success(function(ts){
        templates = ts
        $("#loading").hide()
        if (templates.length > 0){
            $("#templateTable").show()
            templateTable = $("#templateTable").DataTable();
            templateTable.clear()
            $.each(templates, function(i, template){
                templateTable.row.add([
                    template.name,
                    moment(template.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                    "<div class='pull-right'><button class='btn btn-primary' data-toggle='modal' data-target='#modal' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
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
        errorFlash("Error fetching templates")
    })
}

$(document).ready(function(){
    load()
})
