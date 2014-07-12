var app = angular.module('gophish', ['ngRoute', 'ngTable', 'ngResource', 'ui.bootstrap', 'highcharts-ng', 'angularFileUpload']);

app.config(function($routeProvider) {
    $routeProvider

    // route for the home page
    .when('/', {
        templateUrl: 'js/app/partials/dashboard.html',
        controller: 'DashboardCtrl'
    })

    .when('/campaigns', {
        templateUrl: 'js/app/partials/campaigns.html',
        controller: 'CampaignCtrl'
    })

    .when('/campaigns/:id', {
        templateUrl: 'js/app/partials/campaign_results.html',
        controller: 'CampaignResultsCtrl'
    })

    .when('/users', {
        templateUrl: 'js/app/partials/users.html',
        controller: 'GroupCtrl'
    })

    .when('/templates', {
        templateUrl: 'js/app/partials/templates.html',
        controller: 'TemplateCtrl'
    })

    .when('/settings', {
        templateUrl: 'js/app/partials/settings.html',
        controller: 'SettingsCtrl'
    })
});

app.config( [
    '$compileProvider',
    function( $compileProvider )
    {   
        $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|ftp|mailto|data):/);
        // Angular before v1.2 uses $compileProvider.urlSanitizationWhitelist(...)
    }
]);

app.filter('cut', function() {
    return function(value, max, tail) {
        if (!value) return '';
        max = parseInt(max, 10);
        truncd = []
        for(var i=0; i < Math.min(value.length,max); i++) {
            if (i == max-1) truncd.push("...")
            else truncd.push(value[i].email);
        }
        return truncd;
    };
});

// Example provided by http://docs.angularjs.org/api/ng/type/ngModel.NgModelController
app.directive('contenteditable', function() {
    return {
        restrict: 'A', // only activate on element attribute
        require: '?ngModel', // get a hold of NgModelCtrl
        link: function(scope, element, attrs, ngModel) {
            if (!ngModel) return; // do nothing if no ng-model

            // Specify how UI should be updated
            ngModel.$render = function() {
                element.html(ngModel.$viewValue || '');
            };

            // Listen for change events to enable binding
            element.on('blur keyup change', function() {
                scope.$apply(read);
            });

            // Write data to the model

            function read() {
                var html = element.html();
                // When we clear the content editable the browser leaves a <br> behind
                // If strip-br attribute is provided then we strip this out
                //if (attrs.stripBr && html == '<br>') {
                //    html = '';
                //}
                ngModel.$setViewValue(html);
            }
        }
    };
});
