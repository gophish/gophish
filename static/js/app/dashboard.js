var overview_chart_options = {
    animationEasing:"linear"
}

function load(){
    api.campaigns.get()
    .success(function(campaigns){
        if (campaigns.length > 0){
            var overview_ctx = $("#overview_chart").get(0).getContext("2d");
            // Create the overview chart data
            var overview_data = {labels:[],data:[]}
            var average = 0
            $("#emptyMessage").hide()
            $("#campaignTable").show()
            campaignTable = $("#campaignTable").DataTable();
            $.each(campaigns, function(i, campaign){
                // Add it to the table
                campaignTable.row.add([
                    campaign.created_date,
                    campaign.name,
                    campaign.status
                ]).draw()
                // Add it to the chart data
                overview_data.labels.push(camaign.created_date)
                $.each(campaign.results, function(j, result){
                    if (result.status == "Success"){
                        campaign.y++;
                    }
                })
                campaign.y = Math.floor((campaign.y / campaign.results.length) * 100)
                average += campaign.y
                overview_data.data.push(campaign.y)
            })
            average = Math.floor(average / campaigns.length);
            var overview_chart = new Chart(overview_ctx).Line(campaigns, overview_chart_options);
        }
    })
    .error(function(){
        errorFlash("Error fetching campaigns")
    })
}

$(document).ready(function(){
    load()
})
