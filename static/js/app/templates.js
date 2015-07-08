var template = {attachments:[]}
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
function save(){
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

function dismiss(){
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    template = {attachments:[]}
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

function edit(t){
    $("#html_editor").ckeditor()
    $("#attachmentsTable").show()
    $("#attachmentsTable").DataTable({
        "aoColumnDefs" : [{
            "targets" : [3,4],
            "sClass" : "datatable_hidden"
        }]
    });
    if (t == "new") {
        template = {attachments:[]}
    }
}

function load(){
    api.templates.get()
    .success(function(templates){
        if (templates.length > 0){
            $("#emptyMessage").hide()
            $("#templateTable").show()
            templateTable = $("#templateTable").DataTable();
            $.each(templates, function(i, template){
                templateTable.row.add([
                    template.name,
                    moment(template.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
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
        errorFlash("Error fetching templates")
    })
}

$(document).ready(function(){
    load()
})
