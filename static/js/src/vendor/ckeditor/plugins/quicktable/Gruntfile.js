/* jshint node: true */
var fs = require('fs');

module.exports = function (grunt) {
	"use strict";

	grunt.initConfig({
		pkg: grunt.file.readJSON('package.json'), 
		jshint: {
			all: [
				"lang/*.js", 
				"*.js",
				"*.json"
			], 
			options: {
				jshintrc: '.jshintrc'
			}
		},
		lint5: {
			dirPath: "samples",
			templates: [
				"quicktable.html"
			]
			//,
			// ignoreList: [
			// ]
		},
		compress: {
			main: {
				options: {
					archive: 'release/<%= pkg.name %>-<%= pkg.version %>.zip',
					level: 9,
					pretty: true
				},
				files: [
					{
						src: [
							'**', 
							// Exclude files and folders
							'!node_modules/**',
							'!release/**',
							'!.*',
							'!*.log',
							'!Gruntfile.js',
							'!package.json',
							'!LICENSE',
							'!*.md',
							'!template.jst',
							'!*.zip'
						], 
						dest: '<%= pkg.name %>/'
					}
				]
			}
		},
		markdown: {
			all: {
				files: [
					{
						expand: true,
						src: '*.md',
						dest: 'release/docs/',
						ext: '.html'
					}
				],
				options: {
					template: 'template.jst',
					//preCompile: function(src, context) {},
					//postCompile: function(src, context) {},
					//templateContext: {},
					markdownOptions: {
						gfm: true,
						highlight: 'manual'
					}
				}
			}
		}
	});

	function replaceContent(file, searchArray) {
		fs.readFile(file, 'utf8', function (err,data) {
			if (err) {
				return grunt.log.writeln(err);
			}
			
			var result = data;
			for (var i = 0; i < searchArray.length;i++){
				result = result.replace(searchArray[i][0], searchArray[i][1]);
			}
			fs.writeFile(file, result, 'utf8', function (err) {
				if (err) {
					return grunt.log.writeln(err);
				}
			});
		});
	}
	
	grunt.loadNpmTasks('grunt-markdown');
	grunt.loadNpmTasks('grunt-contrib-compress');
	grunt.loadNpmTasks('grunt-lint5');
	grunt.loadNpmTasks('grunt-contrib-jshint');

	grunt.registerTask('test', ['jshint', 'lint5']);
	grunt.registerTask('build-only', ['beforeCompress', 'compress', 'afterCompress']);
	grunt.registerTask('build', ['test', 'beforeCompress', 'compress', 'afterCompress', 'markdown']);
	grunt.registerTask('default', ['test']);
	//Custom tasks
	grunt.registerTask('beforeCompress', 'Running before Compression', function() {
		replaceContent('samples/quicktable.html', [ 
			[/http\:\/\/cdn.ckeditor.com\/4.4.3\/full-all\//g, '../../../'],
			[/language: 'en'/g, '// language: \'en\''],
			[/<!-- REMOVE BEGIN -->/g, '<!-- REMOVE BEGIN --><!--']
		]);
	});
	grunt.registerTask('afterCompress', 'Running after Compression', function() {
		replaceContent('samples/quicktable.html', [
			[/\.\.\/\.\.\/\.\.\//g, 'http://cdn.ckeditor.com/4.4.3/full-all/'],
			[/\/\/ language: 'en'/g, 'language: \'en\''],
			[/<!-- REMOVE BEGIN --><!--/g, '<!-- REMOVE BEGIN -->']
		]);
	});
	
};