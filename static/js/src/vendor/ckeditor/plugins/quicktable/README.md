quicktable
==========
[![Build Status](https://travis-ci.org/ufdada/quicktable.svg?branch=master)](https://travis-ci.org/ufdada/quicktable)
[![devDependency Status](https://david-dm.org/ufdada/quicktable/dev-status.svg)](https://david-dm.org/ufdada/quicktable#info=devDependencies)
###A quicktable plugin for ckeditor 4

This plugin adds a quicktable feature to the existing table plugin.

The original code was submitted by [danyaPostfactum](https://github.com/danyaPostfactum) as a pull request for the table plugin. 
I just extracted the code and made a seperate plugin out of it and added some options to it (see sample in plugin directory)

__*The original table plugin is required for this to work!*__

####Building:
This requires **node.js** and **npm** to be installed.

Then open the directory with the *git-shell* and type `npm install` to install the required packages.

After that you can execute grunt with the following options:

1. `grunt test`
 - Linting the js and html files (check for syntax errors etc.). See [JS Hint Configuration File](https://raw.githubusercontent.com/ufdada/quicktable/master/.jshintrc) for options __*(default)*__
2. `grunt build`
 - `grunt test` and compressing the plugin into a zip file. The zip file is located in the *release* directory.
3. `grunt build-only`
 - Just generating the zip file without linting and markdown compile *(not recommended)*

####Installation:
Just copy the whole directory (quicktable) in the plugins directory

####Configuration:

```javascript
	CKEDITOR.replace( 'editor1', {
		qtRows: 20, // Count of rows in the quicktable (default: 8)
		qtColumns: 20, // Count of columns in the quicktable (default: 10)
		qtBorder: '1', // Border of the inserted table (default: '1')
		qtWidth: '90%', // Width of the inserted table (default: '500px')
		qtStyle: { 'border-collapse' : 'collapse' }, // Content of the style-attribute of the inserted table (default: null)
		qtClass: 'test', // Class of the inserted table (default: '')
		qtCellPadding: '0', // Cell padding of the inserted table (default: '1')
		qtCellSpacing: '0', // Cell spacing of the inserted table (default: '1')
		qtPreviewBorder: '4px double black', // Border of the preview table (default: '1px solid #aaa')
		qtPreviewSize: '4px', // Cell size of the preview table (default: '14px')
		qtPreviewBackground: '#c8def4' // Cell background of the preview table on hover (default: '#e5e5e5')
	});
```

####Known Issues:
- Some missing translations

For more Information see original post [here](https://github.com/ckeditor/ckeditor-dev/pull/92)
