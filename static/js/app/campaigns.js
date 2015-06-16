$(document).ready(function(){
    campaigns.get()
    .success(function(data){
        successFlash("worked!")
        console.log(data)
    })
    .error(function(data){
        errorFlash("No work")
        console.log(data)
    })
    $("#table_id").DataTable();
})
