const path = require('path');

module.exports = {
    context: path.resolve(__dirname, 'static', 'js', 'src', 'app'),
    entry: {
        passwords: './passwords',
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