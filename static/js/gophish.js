$(document).ready(function() {
	$(".editable-row").hover(
        function() {
            alert("test")
            $(this).find(".edit-row").removeClass("hidden");
        }, function() {
            $(this > ".edit-row").addClass("hidden");
        }
    )
})
    