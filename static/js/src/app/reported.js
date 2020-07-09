let emails = []
let notes = {}

let statusBtnClass = {
                "Unknown" : "btn btn-warning dropdown-toggle statusbtn",
                "Safe" : "btn btn-success dropdown-toggle statusbtn",
                "Harmful" : "btn btn-danger dropdown-toggle statusbtn"
                }

const load = () => {
    $("#reportedTable").hide()
    $("#loading").show()
    api.reported.get()
        .success((em) => {
            emails = em
            $("#loading").hide()
            $("#reportedTable").show()
            let reportedTable = $("#reportedTable").DataTable({
                destroy: true,
                "aaSorting": [], // Disable default sort
                columnDefs: [{
                    orderable: false,
                    targets: "no-sort"
                }]
            });
            reportedTable.clear();
            $.each(emails, (i, email) => {

                statusBtn = '<div class="dropdown">\
                            <button id="btnstatus-'+email.id+'" class="' + statusBtnClass[email.status] + '" data-toggle="dropdown">' + email.status + '</button>\
                            <ul class="dropdown-menu" id="' + email.id + '">\
                                <li><a onclick="updateStatus(' + email.id + ', \'Safe\')" href="#">Safe</a></li>\
                                <li><a onclick="updateStatus(' + email.id + ', \'Harmful\')" href="#">Harmful</a></li>\
                                <li><a onclick="updateStatus(' + email.id + ', \'Unknown\')" href="#">Unknown</a></li>\
                                <li class="divider"></li>\
                                <li><a onclick="loadNotes(' + email.id + ')" class="notes" href="#">Notes</a></li>\
                                </ul>\
                            </div>'

                notes[email.id] = email.notes

                attHTML =
                "<div class='dropdown'>\
                <button class='btn btn-primary dropdown-toggle' type='button' data-toggle='dropdown'>" + email.attachments.length + " files\
                <span class='caret'></span></button>"

                if (email.attachments.length > 0 ){
                    attHTML += "<ul class='dropdown-menu'>"
                    $.each(email.attachments, function(i, a){    
                        attHTML += "<li id="+ a.id +"><a href=\"/reported/attachment/\" onclick=\"window.open('/reported/attachment/" + a.id + "', 'newwindow', 'width=640, height=480'); return false;\">" + a.filename + " ("+ humanFileSize(a.size) +")</a></li>"
                    });
                    attHTML += "</ul>"
                    
                }
                attHTML += "</div>"

                subj = escapeHtml(email.reported_subject)
                if (subj.length > 24) {
                  subj = subj.substring(0, 24) + "..."
                }

                reportedTable.row.add([
                    "<span data-toggle='tooltip' title='" + escapeHtml(email.reported_by_name) + "'> " + escapeHtml(email.reported_by_email) + "</span>",
                    //escapeHtml(email.reported_subject.substring(0, 24) + "..."),
                    subj,
                    moment.utc(email.reported_time).fromNow(),
                    attHTML,
                    statusBtn,
                    "<div class=''><button class='btn btn-primary edit_button' onclick='viewEmail(" + email.id + ")' data-backdrop='static' data-user-id='" + email.id + "'>\
                    <i class='fa fa-eye'></i>\
                    </button>\
                    <button class='btn btn-danger delete_button' onclick='deleteEmail(" + email.id + ")' data-user-id='" + email.id + "'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                ]).draw()
            })
        })
        .error(() => {
            errorFlash("Error fetching reported emails")
        })
}

$(document).ready(function () {

    load()
    window.setInterval(function(){  //Refresh every 10 seconds
          load()
    }, 10000);
  }); 

function updateStatus(emailID, newstatus){
  
    // Update button
    btn = $("#btnstatus-" + emailID)
    btn.attr('class', statusBtnClass[newstatus])        
    btn.text(newstatus)
    btn.val(newstatus)
  
    //Update server side value
    email = {
                id: parseInt(emailID),
                status: newstatus
            }
  
    api.reported.put(email)
      .error(function (data) {
          Swal.fire({
            type: 'error',
            title: data.responseJSON.message
          })
      })
  
  } 
    
function loadNotes(emailID){

  email = {
    id: parseInt(emailID),
  }

  $("#notes").val(notes[emailID]);
  $("#notes-emailid").val(emailID);
  $("#modal\\.flashes").empty()
  $('#modal').modal('show');

}   

function deleteEmail(id) {

  Swal.fire({
      title: "Are you sure?",
      text: "This will delete the email from here, but not from your mail server.",
      type: "warning",
      animation: false,
      showCancelButton: true,
      confirmButtonText: "Delete",
      confirmButtonColor: "#428bca",
      reverseButtons: true,
      allowOutsideClick: false,
      preConfirm: function () {
          return new Promise((resolve, reject) => {
              api.reported.delete(id)
                  .success((msg) => {
                      resolve()
                  })
                  .error((data) => {
                      reject(data.responseJSON.message)
                  })
          })
          .catch(error => {
              Swal.showValidationMessage(error)
            })
      }
  }).then(function (result) {
        load()    
  })
}


function viewEmail(id){
  
  $("#modal-email\\.flashes").empty()
  $('#modal-email').modal('show');

  api.reported.getone(id)
        .success((em) => {

          if (em.length > 0 ) { // Should always be one, but safe to check.
            rtext = em[0].reported_text.replace(/(?:\r\n|\r|\n)/g, '<br>');
            rhtml = em[0].reported_html

            //$("#email-plaintext").attr('value', btoa(rtext))
            //$("#email-html").attr('value', btoa(rhtml))
            $("#email-plaintext").data("value", rtext)
            $("#email-html").data("value", rhtml)


            $("#email-body").attr("srcdoc", rtext); // Load plaintext by default
          } else {
            modalError("Error loading email")
          }

      })
      .error(function (data) {
        modalError("Error loading email: " + data.responseJSON.message)
      })


};

function viewplaintext() {
  rtext = $("#email-plaintext").data("value")
  $("#email-body").attr("srcdoc", rtext);
}

function viewhtml() {
  rhtml = $("#email-html").data("value")
  $("#email-body").attr("srcdoc", rhtml);
}

$("#modalSubmit").unbind('click').click(() => {
    
    emailID = $("#notes-emailid").attr('value')
    newnotes = $("#notes").val()
    notes[emailID] = newnotes

    email = {
              "id": parseInt(emailID),
              "notes": notes[emailID]
            }

    api.reported.put(email)
    .success(function (data) {
      $("#modal").modal('hide')
      })
      .error(function (data) {
        modalError("Error saving notes: " + data.responseJSON.message)
      })

})
  
// Convert attachment byte file size to human readable format
function humanFileSize(bytes, si=true, dp=0) {
    const thresh = si ? 1000 : 1024;
  
    if (Math.abs(bytes) < thresh) {
      return bytes + ' B';
    }
  
    const units = si 
      ? ['kb', 'mb', 'gb', 'tb', 'pb', 'eb', 'zb', 'yb'] 
      : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    let u = -1;
    const r = 10**dp;
  
    do {
      bytes /= thresh;
      ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);
  
  
    return bytes.toFixed(dp) + ' ' + units[u];
  }
