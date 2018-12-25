function save(e) {
   
    var tg = {};

    tg.name = $("#name").val();
    tg.weight = parseInt($("#weight").val());
    

    -1 != e ? (tg.id = e, api.phishtags.put(tg).success(function(e) {
        successFlash("Template edited successfully!"), load(), dismiss()
    }).error(function(e) {
        modalError(e.responseJSON.message)
    })) : api.phishtags.post(tg).success(function(e) {

        successFlash("Category added successfully!"), load(), dismiss()
    }).error(function(e) {
        modalError(e.responseJSON.message)
    })
    
}

function dismiss() {
    $("#modal\\.flashes").empty(), $("#name").val(""), $("#weight").val(""), $("#modal").modal("hide")
}

function deleteTag(e) {
    confirm("Delete " + e.name + "?") && api.phishtagssingle.delete(e.id).success(function(e) {
        successFlash(e.message), load()
    })
}

function edit(e) {
    $("#modalSubmit").unbind("click").click(function() {
        save(e)
    });
    -1 != e ? (api.phishtags.single(e).success(function(tg) { 
        $("#name").val(tg.name), $("#weight").val(tg.weight)
     })) : ""
}

function load() {
    $("#categoriesTable").hide(), $("#emptyMessage").hide(), $("#loading").show(), api.phishtags.get().success(function(e) {
        
        categories = e, $("#loading").hide(), categories.length > 0 ? ($("#categoriesTable").show(), categoriesTable = $("#categoriesTable").DataTable({
            destroy: !0,
            columnDefs: [{
                orderable: !1,
                targets: "no-sort"
            }]
        }), categoriesTable.clear(), $.each(categories, function(e, a) {
            categoriesTable.row.add([escapeHtml(a.name), a.weight, "<div class='pull-right'><span data-toggle='modal' data-backdrop='static' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Edit Category' onclick='edit(" + a.id + ")'>                    <i class='fa fa-pencil'></i>                    </button></span>\t\t  <button class='btn btn-danger' data-toggle='tooltip' data-placement='left' title='Delete Category' onclick='deleteTag(" + a.id + ")'>                    <i class='fa fa-trash-o'></i>                    </button></div>"]).draw()
        }), $('[data-toggle="tooltip"]').tooltip()) : $("#emptyMessage").show()
    }).error(function() {
        $("#loading").hide(), errorFlash("Error fetching categories")
    })
}

var tags = [],
    icons = {
        "application/vnd.ms-excel": "fa-file-excel-o",
        "text/plain": "fa-file-text-o",
        "image/gif": "fa-file-image-o",
        "image/png": "fa-file-image-o",
        "application/pdf": "fa-file-pdf-o",
        "application/x-zip-compressed": "fa-file-archive-o",
        "application/x-gzip": "fa-file-archive-o",
        "application/vnd.openxmlformats-officedocument.presentationml.presentation": "fa-file-powerpoint-o",
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": "fa-file-word-o",
        "application/octet-stream": "fa-file-o",
        "application/x-msdownload": "fa-file-o"
    },
    deleteTag = function(e) {
        swal({
            title: "Are you sure?",
            text: "This will delete the tag. This can't be undone!",
            type: "warning",
            animation: !1,
            showCancelButton: !0,
            confirmButtonText: "Delete Tag",
            confirmButtonColor: "#428bca",
            reverseButtons: !0,
            allowOutsideClick: !1,
            preConfirm: function() {
                return new Promise(function(t, a) {
                    api.phishtags.delete(e).success(function(e) {
                        t()
                    }).error(function(e) {
                        a(e.responseJSON.message)
                    })
                })
            }
        }).then(function() {
            swal("Tag Deleted!", "This tag has been deleted!", "success"), $('button:contains("OK")').on("click", function() {
                location.reload()
            })
        })
    };

$(document).ready(function() {
    $(".modal").on("hidden.bs.modal", function(e) {
        $(this).removeClass("fv-modal-stack"), $("body").data("fv_open_modals", $("body").data("fv_open_modals") - 1)
    }), $(".modal").on("shown.bs.modal", function(e) {
        void 0 === $("body").data("fv_open_modals") && $("body").data("fv_open_modals", 0), $(this).hasClass("fv-modal-stack") || ($(this).addClass("fv-modal-stack"), $("body").data("fv_open_modals", $("body").data("fv_open_modals") + 1), $(this).css("z-index", 1040 + 10 * $("body").data("fv_open_modals")), $(".modal-backdrop").not(".fv-modal-stack").css("z-index", 1039 + 10 * $("body").data("fv_open_modals")), $(".modal-backdrop").not("fv-modal-stack").addClass("fv-modal-stack"))
    }), $.fn.modal.Constructor.prototype.enforceFocus = function() {
        $(document).off("focusin.bs.modal").on("focusin.bs.modal", $.proxy(function(e) {
            this.$element[0] === e.target || this.$element.has(e.target).length || $(e.target).closest(".cke_dialog, .cke").length || this.$element.trigger("focus")
        }, this))
    }, $(document).on("hidden.bs.modal", ".modal", function() {
        $(".modal:visible").length && $(document.body).addClass("modal-open")
    }), $("#modal").on("hidden.bs.modal", function(e) {
        dismiss()
    }), $("#sendTestEmailModal").on("hidden.bs.modal", function(e) {
        dismissSendTestEmailModal()
    }), $("#headersForm").on("submit", function() {
        return headerKey = $("#headerKey").val(), headerValue = $("#headerValue").val(), "" != headerKey && "" != headerValue && (addCustomHeader(headerKey, headerValue), $("#headersForm>div>input").val(""), $("#headerKey").focus(), !1)
    }), $("#headersTable").on("click", "span>i.fa-trash-o", function() {
        headers.DataTable().row($(this).parents("tr")).remove().draw()
    }), load()
});