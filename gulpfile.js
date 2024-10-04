/*
 * gulpfile.js
 *
 * Description: The Gophish gulpfile
 */

var gulp = require('gulp'),
    rename = require('gulp-rename'),
    concat = require('gulp-concat'),
    uglify = require('gulp-uglify-es').default,
    cleanCSS = require('gulp-clean-css'),
    babel = require('gulp-babel'),

    js_directory = 'static/js/src/',
    css_directory = 'static/css/',
    app_directory = js_directory + 'app/',
    dest_js_directory = 'static/js/dist/',
    dest_css_directory = 'static/css/dist/';

vendorjs = function () {
    const pkg = require('./package.json');
    const libs = Object.keys(pkg.dependencies);
    return gulp.src(
        libs.map(function(lib) {
            return './node_modules/' + lib + '/**/*.js';
        })
    )
        .pipe(concat('vendor.js'))
        .pipe(rename({
            suffix: '.min'
        }))
        .pipe(uglify())
        .pipe(gulp.dest(dest_js_directory));
}

scripts = function () {
    // Gophish app files - non-ES6
    return gulp.src([
            app_directory + 'autocomplete.js',
            app_directory + 'campaign_results.js',
            app_directory + 'campaigns.js',
            app_directory + 'dashboard.js',
            app_directory + 'groups.js',
            app_directory + 'landing_pages.js',
            app_directory + 'sending_profiles.js',
            app_directory + 'settings.js',
            app_directory + 'templates.js',
            app_directory + 'gophish.js',
            app_directory + 'users.js',
            app_directory + 'webhooks.js',
            app_directory + 'passwords.js'
        ])
        .pipe(rename({
            suffix: '.min'
        }))
        .pipe(uglify().on('error', function (e) {
            console.log(e);
        }))
        .pipe(gulp.dest(dest_js_directory + 'app/'));
}

styles = function () {
    return gulp.src([
            css_directory + 'bootstrap.min.css',
            css_directory + 'main.css',
            css_directory + 'dashboard.css',
            css_directory + 'flat-ui.css',
            css_directory + 'dataTables.bootstrap.css',
            css_directory + 'font-awesome.min.css',
            css_directory + 'chartist.min.css',
            css_directory + 'bootstrap-datetime.css',
            css_directory + 'checkbox.css',
            css_directory + 'sweetalert2.min.css',
            css_directory + 'select2.min.css',
            css_directory + 'select2-bootstrap.min.css',
        ])
        .pipe(cleanCSS({
            compatibilty: 'ie9'
        }))
        .pipe(concat('gophish.css'))
        .pipe(gulp.dest(dest_css_directory));
}

exports.vendorjs = vendorjs
exports.scripts = scripts
exports.styles = styles
exports.build = gulp.parallel(vendorjs, scripts, styles)
exports.default = exports.build