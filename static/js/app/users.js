var groups = []
var LocalData = []
var Data_View = []
var targetsTable ;
var tableInfo = {};
tableInfo.start=0;
tableInfo.order=[{"column":0,"dir":"asc"}];
tableInfo.search={"regex":false,"value":""};
tableInfo.length=0;

// Save attempts to POST or PUT to /groups/
function save(idx) {
    var targets = []
    targets = LocalData;
    var group = {
            name: $("#name").val(),
            targets: targets
        }
        // Submit the group
    if (idx != -1) {
        // If we're just editing an existing group,
        // we need to PUT /groups/:id
        group.id = groups[idx].id
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
}

function removeDuplicates() {
    LocalData = LocalData.sort(
        function(a,b){
            return a.email>b.email?1:0;
        }
    );
    var newLocalData = [];
    var prev = {"email":""};
    $.each(LocalData, function(i, target) {
        if(prev.email != target.email){
            prev = target;
            newLocalData.push(target);
        }
    });
    LocalData = newLocalData;
}

function xssParsing(input) {
    return [escapeHtml(input["first_name"]),escapeHtml(input["last_name"]),escapeHtml(input["email"]).toLowerCase(),escapeHtml(input["position"]),'<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>']
}


function edit(idx) {

    targetsTable = $("#targetsTable").dataTable({
        processing: true,
        serverSide: true,
        ajax:{
          url:"/api/datatables/dummy",
          data:function(params){
              tableInfo.start = params.start;
              tableInfo.length = params.length;
              tableInfo.order = params.order;
              tableInfo.search = params.search;
              delete params.draw;
              delete params.columns;
              delete params.length;
              delete params.start;
              return;
          },
          dataSrc:function(json){
              Data_View = [];
              $.each(LocalData,function(i,val){
                  if (val.email.search(tableInfo.search.value)>=0 || val.first_name.search(tableInfo.search.value)>=0 || val.last_name.search(tableInfo.search.value)>=0 || val.position.search(tableInfo.search.value)>=0)
                    Data_View.push(val)
              });
              Data_View = Data_View.sort(
                  function(a,b){
                      if(tableInfo.order[0].column==0){
                        if(tableInfo.order[0].dir=="asc"){
                            return a.first_name>b.first_name?1:0;
                        } else {
                            return a.first_name<b.first_name?1:0;
                        }
                      } else if(tableInfo.order[0].column==1){
                          if(tableInfo.order[0].dir=="asc"){
                              return a.last_name>b.last_name?1:0;
                          } else {
                              return a.last_name<b.last_name?1:0;
                          }
                      } else if(tableInfo.order[0].column==2){
                          if(tableInfo.order[0].dir=="asc"){
                              return a.email>b.email?1:0;
                          } else {
                              return a.email<b.email?1:0;
                          }
                      } else if(tableInfo.order[0].column==3){
                          if(tableInfo.order[0].dir=="asc"){
                              return a.position>b.position?1:0;
                          } else {
                              return a.position<b.position?1:0;
                          }
                      }
                  }
              );
              var data = [];
              for(var i=tableInfo.start;i<tableInfo.start+tableInfo.length && i<Data_View.length;i++){
                  data.push(xssParsing(Data_View[i]));
              }
              json.recordsFiltered = Data_View.length;
              json.recordsTotal = Data_View.length;
              json.data = data;
              return json.data;
          }
      },
        // infoCallback:   function( settings, start, end, max, total, pre ) {
        //                     var info = "";
        //                     if(LocalData.length<=tableInfo.length){
        //                         info += "Showing "+(LocalData.length)+" to "+LocalData.length;
        //                     } else{
        //                         info += "Showing "+(tableInfo.start+1)+" to "+(tableInfo.start+1+tableInfo.length);
        //                     }
        //                     info += " of "+LocalData.length+" entries"
        //                     return info;
        //                 },
        // ordering: false,
        deferRender:    true,
        destroy: true, // Destroy any other instantiated table - http://datatables.net/manual/tech-notes/3#destroy
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    })
    $("#modalSubmit").unbind('click').click(function() {
        save(idx)
    })
    if (idx == -1) {
        group = {}
        LocalData = [];
    } else {
        group = groups[idx]
        $("#name").val(group.name)
        LocalData = group.targets;
        targetsTable.DataTable().draw();
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
            LocalData = LocalData.concat(data.result);
            removeDuplicates();
            targetsTable.DataTable().draw();
        }
    })
}

function deleteGroup(idx) {
    if (confirm("Delete " + groups[idx].name + "?")) {
        api.groupId.delete(groups[idx].id)
            .success(function(data) {
                successFlash(data.message)
                load()
            })
    }
}
function load() {
    $("#groupTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.groups.get()
        .success(function(gs) {
            $("#loading").hide()
            if (gs.length > 0) {
                groups = gs
                $("#emptyMessage").hide()
                $("#groupTable").show()
                groupTable = $("#groupTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                groupTable.clear();
                $.each(groups, function(i, group) {
                    var targets = ""
                    $.each(group.targets, function(i, target) {
                        targets += target.email + ", "
                        if (targets.length > 50) {
                            targets = targets.slice(0, -3) + "..."
                            return false;
                        }
                    })
                    groupTable.row.add([
                        escapeHtml(group.name),
                        escapeHtml(targets),
                        moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'><button class='btn btn-primary' data-toggle='modal' data-target='#modal' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-danger' onclick='deleteGroup(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function() {
    load()
    // Setup the event listeners
    // Handle manual additions
    $("#targetForm").submit(function() {
        LocalData.push({
            first_name: $("#firstName").val(),
            last_name: $("#lastName").val(),
            email: $("#email").val().toLowerCase(),
            position: $("#position").val()
        })
        removeDuplicates();
        targetsTable.DataTable().draw();
        // Reset user input.
        $("#targetForm>div>input").val('');
        $("#firstName").focus();
        return false;
    });
    // Handle Deletion
    $("#targetsTable").on("click", "span>i.fa-trash-o", function() {
        var primaryKey = Data_View[$("#targetsTable tbody tr").index($(this).parents('tr'))].email;
        for(var i=0;i<LocalData.length;i++){
            if(LocalData[i].email == primaryKey){
                LocalData.splice(i,1)
                break;
            }
        }
        // LocalData.splice(,1)
        targetsTable.DataTable().draw();
    });
    $("#modal").on("hide.bs.modal", function() {
        dismiss();
    });
});
