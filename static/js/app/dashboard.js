var campaigns = []

$(document).ready(function(){
    api.campaigns.get()
    .success(function(cs){
        campaigns = cs
        if (campaigns.length > 0){
            // Create the overview chart data
            var overview_data = {labels:[],series:[[]]}
            var average_data = {series:[]}
            var overview_opts = {
                axisX: {
                    showGrid: false
                },
                showArea: true,
                plugins: []
            }
            var average_opts = {
                donut : true,
                donutWidth: 30
            }
            var average = 0
            $("#emptyMessage").hide()
            $("#campaignTable").show()
            campaignTable = $("#campaignTable").DataTable();
            $.each(campaigns, function(i, campaign){
                var campaign_date = moment(campaign.created_date).format('MMMM Do YYYY h:mm')
                // Add it to the table
                campaignTable.row.add([
                    campaign.name,
                    campaign_date,
                    campaign.status
                ]).draw()
                // Add it to the chart data
                campaign.y = 0
                $.each(campaign.results, function(j, result){
                    if (result.status == "Success"){
                        campaign.y++;
                    }
                })
                campaign.y = Math.floor((campaign.y / campaign.results.length) * 100)
                average += campaign.y
                // Add the data to the overview chart
                overview_data.labels.push(campaign_date)
                overview_data.series[0].push({meta : i, value: campaign.y})
            })
            average = Math.floor(average / campaigns.length);
            average_data.series.push({meta: "Successful Phishes", value: average})
            average_data.series.push({meta: "Unsuccessful Phishes", value: 100 - average})
            var average_chart = new Chartist.Pie("#average_chart", average_data, average_opts)
            var overview_chart = new Chartist.Line('#overview_chart', overview_data, overview_opts)
            $chart = $("#overview_chart")
            $chart.on("click", '.ct-point', function(d){console.log(d)});
            var $toolTip = $chart
            .append('<div class="chartist-tooltip"></div>')
            .find('.chartist-tooltip')
            .hide();

            $chart.on('mouseenter', '.ct-point', function() {
                var $point = $(this)
                value = $point.attr('ct:value')
                cidx = $point.attr('ct:meta')
                $toolTip.html(campaigns[cidx].name + '<br>' + "Successes: " + value.toString()).show();
            });

            $chart.on('mouseleave', '.ct-point', function() {
                $toolTip.hide();
            });
            $chart.on('mousemove', function(event) {
                $toolTip.css({
                    left: (event.offsetX || event.originalEvent.layerX) - $toolTip.width() / 2 - 10,
                    top: (event.offsetY || event.originalEvent.layerY) - $toolTip.height() - 40
                });
            });
            $("#overview_chart").on("click", ".ct-point", function(e) {
                var $cidx = $(this).attr('ct:meta');
                window.location.href = "/campaigns/" + campaigns[cidx].id
            });
        }
    })
    .error(function(){
        errorFlash("Error fetching campaigns")
    })
})
