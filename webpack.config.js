const path = require('path');

module.exports = {
    context: path.resolve(__dirname, 'static', 'js', 'src', 'app'),
    entry: {
        autocomplete: './autocomplete',
        campaign_results: './campaign_results',
        campaigns: './campaigns',
        dashboard: './dashboard',
        gophish: './gophish',
        groups: './groups',
        landing_pages: './landing_pages',
        passwords: './passwords',
        templates: './templates',
        sending_profiles: './sending_profiles',
        settings: './settings',
        users: './users',
        webhooks: './webhooks',
    },
    output: {
        path: path.resolve(__dirname, 'static', 'js', 'dist', 'app'),
        filename: '[name].min.js'
    },
    module: {
        rules: [{
            test: /\.js$/,
            exclude: /node_modules/,
            use: {
                loader: "babel-loader"
            }
        }]
    }
}