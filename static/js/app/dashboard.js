var overview_chart_options = {
    animationEasing:"linear",
    customTooltips: function(tooltip) {
        console.log(tooltip)
    }
}

function load(){
    api.campaigns.get()
    .success(function(campaigns){
        if (campaigns.length > 0){
            var overview_ctx = $("#overview_chart").get(0).getContext("2d");
            // Create the overview chart data
            var overview_data = {labels:[],datasets:[{data:[]}]}
            var average = 0
            $("#emptyMessage").hide()
            $("#campaignTable").show()
            campaignTable = $("#campaignTable").DataTable();
            $.each(campaigns, function(i, campaign){
                var campaign_date = moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a')
                // Add it to the table
                campaignTable.row.add([
                    campaign.name,
                    campaign_date,
                    campaign.status
                ]).draw()
                // Add it to the chart data
                overview_data.labels.push(campaign_date)
                campaign.y = 0
                $.each(campaign.results, function(j, result){
                    if (result.status == "Success"){
                        campaign.y++;
                    }
                })
                campaign.y = Math.floor((campaign.y / campaign.results.length) * 100)
                average += campaign.y
                overview_data.datasets[0].data.push({y:campaign.y, test: "test"})
            })
            average = Math.floor(average / campaigns.length);
            var overview_chart = new Chart(overview_ctx).Line(overview_data, overview_chart_options);
        }
    })
    .error(function(){
        errorFlash("Error fetching campaigns")
    })
}

$(document).ready(function(){
    load()
})
